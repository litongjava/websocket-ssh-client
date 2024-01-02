# Websocket-SSH-Client
## 1. 简介
Websocket-SSH-Client是一个基于Go语言开发的服务，它允许SSH客户端通过Websocket连接到远程ssh服务器。以下是该服务的工作流程：

1. **SSH客户端**：这是流程的起点，SSH客户端尝试建立一个连接。

2. **Websocket-SSH-Client**：SSH客户端连接到运行在22端口的Websocket-SSH-Client。这个服务的作用是将SSH协议转换为WSSS协议，这是一个基于HTTP和WSS的协议。

3. **Websocket-SSH-Server**：Websocket-SSH-Client通过WSSS协议连接到运行在5501端口的Websocket-SSH-Server。Websocket-SSH-Server接收到WSSS协议的数据，并将其转换回SSH协议。

4. **SSH服务器**：最后，Websocket-SSH-Server作为SSH客户端，连接到SSH服务器。

以下是这个流程的图示：

```markdown
1. SSH Client
    |
    | SSH
    |
2. Websocket-SSH-Client (Port 22)
    |
    | WSSS (HTTP & WSS)
    |
3. Websocket-SSH-Server (Port 5501)
    |
    | SSH
    |
4. SSH Server
```

## 2. 安装Websocket-SSH-Server
Websocket-SSH-Server（简称WSSS）是一个基于Go语言开发的服务，它允许Websocket-SSH-Client进行连接。以下是安装和配置WSSS的步骤：

### 2.1 安装
首先，你需要下载并解压Websocket-SSH-Server的安装包。你可以使用以下命令来完成这个步骤：

```bash
mkdir /opt/package/wsss -p && cd /opt/package/wsss/
## put to websocket-sshd-server-linux-amd64-v1.0.0.tar.gz to here
## curl -O http://192.168.3.8:3000/websocket-sshd-server/websocket-sshd-server-linux-amd64-v1.0.0.tar.gz
tar -xf websocket-sshd-server-linux-amd64-v1.0.0.tar.gz -C /usr/local/
```

然后，你可以使用以下命令来启动Websocket-SSH-Server：

```bash
/usr/local/websocket-sshd-server/websocket-ssh-server -c /usr/local/websocket-sshd-server/config.yml
```

启动后，Websocket-SSH-Server将默认监听5001端口。

### 2.2 配置Nginx（可选）
如果你想通过Nginx代理Websocket-SSH-Server，你可以添加以下配置到你的Nginx配置文件中：

```nginx
location /wsss { ## 后端项目 - 用户 wsss
    proxy_pass http://localhost:5001;
    proxy_set_header Host $http_host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header REMOTE-HOST $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
}
```

然后，你可以使用以下命令来重载Nginx配置：

```bash
nginx -s reload
```

现在，你可以通过`http://192.168.3.9/wsss/`来访问Websocket-SSH-Server。

### 2.3 配置开机启动（可选）
如果你希望Websocket-SSH-Server在系统启动时自动运行，你可以创建一个systemd服务。首先，使用以下命令创建一个服务文件：

```bash
vi /etc/systemd/system/wsss.service
```

然后，将以下内容复制到服务文件中：

```ini
[Unit]
Description=wsss
After=network.target

[Service]
Type=simple
User=root
Restart=on-failure
RestartSec=5s
WorkingDirectory = /usr/local/websocket-sshd-server
ExecStart=/usr/local/websocket-sshd-server/websocket-ssh-server -c /usr/local/websocket-sshd-server/config.yml

[Install]
WantedBy=multi-user.target
```

最后，你可以使用以下命令来启动和管理你的服务：

```bash
systemctl daemon-reload
systemctl enable wsss
systemctl start wsss
systemctl status wsss
systemctl stop wsss
```
### docker启动
```shell
docker run -dit --name=wsss -p 5001:5001 litongjava/wsss:1.0.0
```
## 3. 安装Websocket-SSH-Client
Websocket-SSH-Client是一个基于Go语言开发的SSH客户端，它允许你通过Websocket连接到SSH服务器。以下是安装和使用Websocket-SSH-Client的步骤：

### 3.1 下载
你可以从以下地址下载Websocket-SSH-Client的最新版本：

[https://github.com/litongjava/websocket-ssh-client/releases](https://github.com/litongjava/websocket-ssh-client/releases)

### 3.2 运行
首先，你需要创建一个配置文件`config.yml`，并填入以下内容：

```yaml
app:
  host: 127.0.0.1
  port: 22
  endPoint: ws://192.168.3.9:5001/wsss/socket
```

然后，你可以使用以下命令来运行Websocket-SSH-Client：

```bash
ssh {username}@{target-host}:{target-host-sshd-port}:{websocket-ssh-client-ip}
```

其中：

- `{username}`：这是你要以其身份登录到远程主机的用户的用户名。
- `{target-host}`：这是你要连接的远程主机的IP地址或主机名。
- `{target-host-sshd-port}`：这是远程主机上运行的SSH服务器（sshd）的端口号。
- `{websocket-ssh-client-ip}`：这是运行Websocket-SSH-Client的主机的IP地址。

例如，如果你想以`root`用户的身份连接到本地主机（IP地址为`127.0.0.1`）的22端口，并且Websocket-SSH-Client也运行在本地主机上，你可以使用以下命令：

```bash
ssh root@127.0.0.1:22@127.0.0.1
```

这是一份为初学者编写的文档，希望你能从中获得所需的信息。如果你有任何问题或需要进一步的帮助，请随时向我提问！