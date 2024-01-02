package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"
	"websocket-ssh-client/config"
)

var EOFBytes = []byte("EOF")

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
}

type ConnectionInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

var private ssh.Signer

func main() {
	log.Println("start...")
	var configFilePath string
	flag.StringVar(&configFilePath, "c", "config.yml", "Configuration file path")
	flag.Parse()

	// 读取配置
	config.ReadFile(configFilePath)

	host := config.CONFIG.App.Host
	port := strconv.Itoa(config.CONFIG.App.Port)
	endPoint := config.CONFIG.App.EndPoint

	// read private key only once
	privateBytes, err := ioutil.ReadFile("ssh/ssh_host_rsa_key")
	if err != nil {
		log.Fatal("Failed to load private key: ", err)
	}

	private, err = ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("Failed to parse private key: ", err)
	}
	sshServer(host, port, endPoint)
}

func sshServer(host string, port string, endPoint string) {
	address := host + ":" + port
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen on %s (%s)", address, err)
	} else {
		log.Println("listening on ", address)
	}

	for {
		socketConnection, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection (%s)", err)
			continue
		} else {
			log.Println("socketConnection:", socketConnection)
		}
		go sshServerConnection(socketConnection, endPoint)
	}
}

func sshServerConnection(connection net.Conn, endPoint string) {
	config := &ssh.ServerConfig{
		PasswordCallback: sshServerPasswordCallback,
	}

	config.AddHostKey(private)
	conn, chans, reqs, err := ssh.NewServerConn(connection, config)
	if err != nil {
		log.Printf("Failed to handshake (%s)", err)
		return
	}
	log.Printf("New SSH connection from %s (%s)", conn.RemoteAddr(), conn.ClientVersion())
	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		go handleChannel(newChannel, conn, endPoint)
	}
}

func sshServerPasswordCallback(conn ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
	fulldesc := conn.User()
	password := string(pass)
	fulldescArray := strings.Split(fulldesc, "@")
	username := fulldescArray[0]
	hostAndPort := strings.Split(fulldescArray[1], ":")
	host := hostAndPort[0]
	var port int
	if len(hostAndPort) > 1 {
		port, _ = strconv.Atoi(hostAndPort[1])
	} else {
		port = 22
	}

	log.Println(host, port, username, password)

	connInfo := ConnectionInfo{
		Username: username,
		Password: password,
		Host:     host,
		Port:     port,
	}

	jsonBytes, _ := json.Marshal(connInfo)
	base64Str := base64.StdEncoding.EncodeToString(jsonBytes)

	return &ssh.Permissions{
		Extensions: map[string]string{
			"json": base64Str,
		},
	}, nil // 密码验证成功
}

func handleChannel(newChannel ssh.NewChannel, conn *ssh.ServerConn, endPoint string) {
	log.Println("new channel:", newChannel)
	if newChannel.ChannelType() != "session" {
		newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
		return
	}

	ch, reqs, err := newChannel.Accept()
	if err != nil {
		log.Println("could not accept channel.")
		return
	}
	ch.Write([]byte("new channel\r\n"))
	defer ch.Close()

	urlStr := endPoint + "?msg=" + conn.Permissions.Extensions["json"]
	websocketConn, _, err := websocket.DefaultDialer.Dial(urlStr, nil)
	if err != nil {
		ch.Write([]byte(err.Error() + "\r\n"))
		return
	} else {
		ch.Write([]byte("websocket connected\r\n"))
	}

	// Create channels for passing messages to and from the WebSocket goroutine.
	wsInput := make(chan []byte)
	wsErrors := make(chan error)

	// Start another goroutine for handling WebSocket write operations.
	go func() {
		for message := range wsInput {
			err := websocketConn.WriteMessage(websocket.BinaryMessage, message)
			if err != nil {
				wsErrors <- err
			}
		}
	}()

	for req := range reqs {
		log.Println("received:", req.Type, req.Payload)
		ok := false

		if len(req.Type) > 255 {
			log.Println("req.Type is too long")
			return
		}

		typeAndPayload := []byte{0x01, byte(len(req.Type))}
		typeAndPayload = append(typeAndPayload, []byte(req.Type)...)
		typeAndPayload = append(typeAndPayload, req.Payload...)

		switch req.Type {
		case "exec":
			ok = true
			wsInput <- typeAndPayload
			log.Println("received command:", string(req.Payload))
			req.Reply(ok, nil)
		case "pty-req":
			ok = true
			wsInput <- typeAndPayload
			log.Println("received:", req.Type)
			ch.Write([]byte("received pty-req\r\n"))
			req.Reply(true, nil)
		case "shell":
			ok = true
			wsInput <- typeAndPayload
			log.Println("received:", req.Type)
			ch.Write([]byte("received shell\r\n"))
			req.Reply(true, nil)
			handleReq(ch, websocketConn, wsInput)
		case "subsystem":
			ok = true
			wsInput <- typeAndPayload
			log.Println("received:", req.Type)
			ch.Write([]byte("received subsystem\r\n"))
			req.Reply(true, nil)
			handleReq(ch, websocketConn, wsInput)
		case "x11-req":
			log.Println("received:", req.Type)
			ch.Write([]byte("received:" + req.Type + "\r\n"))
			ok = true
			req.Reply(false, nil)

			if !ok {
				message := "declining request:" + req.Type + "\r\n"
				log.Println(message)
				ch.Write([]byte(message))
				req.Reply(ok, nil)
			}
		}
	}
}

func handleReq(ch ssh.Channel, websocketConn *websocket.Conn, wsInput chan []byte) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		buf := make([]byte, 256)
		for {
			n, err := ch.Read(buf)
			if err != nil {
				log.Println("read error from ssh client and ssh client well be closed:", err.Error())
				ch.Close()
				websocketConn.Close()
				cancel()
				return
			}
			if n > 0 {
				message := append([]byte{0x02}, buf[:n]...)
				wsInput <- message
				log.Println("send data:", len(buf[:n]))
			}
		}
	}()
	//不知道为什么必须放在这里
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				messageType, message, err := websocketConn.ReadMessage()
				if err != nil {
					// Check if the error is a CloseError.
					if _, ok := err.(*websocket.CloseError); ok {
						msg := "WebSocket server closed the connection"
						log.Println(msg)
						ch.Close()
						return
					}
					log.Println(err.Error())
					return
				} else {
					log.Println("Received message from websocket server:", messageType, len(message))
					if bytes.Contains(message, EOFBytes) {
						log.Println("Received EOF from websocket server and shell channel will be closed ")
						ch.Close()
					} else {
						ch.Write(message)
					}
				}
			}
		}
	}()
}
