# websocket-ssh-client

## 1.安装websocket-ssh-server
websocket-ssh-server 简称wsss
1.1.安装
install
```
mkdir /opt/package/wsss -p && cd /opt/package/wsss/
## put to websocket-sshd-server-linux-amd64-v1.0.0.tar.gz to here
## curl -O http://192.168.3.8:3000/websocket-sshd-server/websocket-sshd-server-linux-amd64-v1.0.0.tar.gz
tar -xf websocket-sshd-server-linux-amd64-v1.0.0.tar.gz -C /usr/local/
```

start
```
/usr/local/websocket-sshd-server/websocket-ssh-server -c /usr/local/websocket-sshd-server/config.yml
```
启动之后默认监听5001端口


1.2.配置nginx(可选)
配置nignx代理
```
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

```
nginx -s reload
```

http://192.168.3.9/wsss/

1.3.配置开机启动(可选)
```
vi /etc/systemd/system/wsss.service
```

```
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

start wsss
```
systemctl daemon-reload
systemctl enable wsss
systemctl start wsss
systemctl status wsss
systemctl stop wsss
```
2.使用docker启动
```
docker run -dit --name=wsss -p 5001:5001 litongjava/wsss:1.0.0
```

## 2.安装websocket-ssh-client

config.yml
```yml
app:
  host: 127.0.0.1
  port: 22
  endPoint: ws://192.168.3.9:5001/wsss/socket
```
run
```shell

```