文件上传与管理小程序

本项目提供一个基于 Go 的 HTTP 文件上传与下载服务，支持：

文件 MD5 去重（相同文件只保存一次，避免重复上传）

上传前预检（判断文件是否已存在或正在上传）

上传进度显示（客户端）

临时文件管理（保证上传过程安全）

跨客户端共享上传结果（通过固定 URL 下载）

支持 .env 配置

功能特点

服务端（Server）

启动 HTTP 服务监听所有网卡 (0.0.0.0)

文件统一保存至 files/ 目录，命名为 <md5>.<ext>

上传前会检查文件是否存在或正在上传，避免重复传输

上传完成后返回标准 JSON：{"status":"ok","url":"http://HOST:PORT/files/<md5>.<ext>"}

提供 /check 接口支持客户端预检

自动清理过期临时文件（tmp_ 前缀）

客户端（Client）

可通过 -file 指定本地文件上传

可通过 -server 覆盖 .env 配置的服务器地址

上传前自动计算文件 MD5 并进行服务端预检

上传进度实时显示

上传完成后返回可访问 URL

环境要求

Go 1.20+

支持 Windows / Linux / Mac

网络可访问服务端 HTTP 端口

安装与依赖

克隆项目或下载到本地目录

初始化 Go 模块（第一次使用）

go mod init file_update


下载依赖

go get github.com/joho/godotenv


构建 Server 和 Client

go build server.go
go build client.go

配置 .env

在项目根目录创建 .env 文件：

HOST=43.229.212.58   # 服务端外网访问地址
HTTP_PORT=8080       # HTTP 服务端口


如果 .env 不存在或缺少值，则默认 HOST=127.0.0.1，HTTP_PORT=80

使用方法
1. 启动服务端
./server.exe
# 或
go run server.go


启动后会显示类似日志：

[INFO] HTTP 服务启动: 0.0.0.0:8080 (外网访问使用 HOST=43.229.212.58)
[INFO] 初始化已知文件映射，记录 0 个文件


服务端会在 files/ 目录保存上传的文件。

2. 客户端上传文件
./client.exe -file "D:\game\example.mp4"
# 或使用 -server 指定服务器
./client.exe -file "D:\game\example.mp4" -server "43.229.212.58:8080"


输出示例：

[INFO] 使用服务器: http://43.229.212.58:8080
[INFO] 计算本地文件 MD5: D:\game\example.mp4
[INFO] 文件 MD5: 5f6ad40f2d4013ac4aa429839c5207ad
[INFO] 预检 (GET) => http://43.229.212.58:8080/check?md5=5f6ad40f2d4013ac4aa429839c5207ad
[INFO] 开始上传: D:\game\example.mp4 (size=206579487)
[PROGRESS] example.mp4 10.23% (21234567/206579487)
...
[SUCCESS] 上传结果: http://43.229.212.58:8080/files/5f6ad40f2d4013ac4aa429839c5207ad.mp4


说明：

文件如果已上传或正在上传，客户端会提示并停止重复传输

上传完成后 URL 可直接在浏览器访问

HTTP 接口说明

预检接口

GET /check?md5=<file_md5>


返回 JSON：

{
  "status": "exists",  // exists | notfound | uploading | error
  "url": "http://HOST:PORT/files/<md5>.<ext>"
}


上传接口

POST /upload?md5=<file_md5>&filename=<file_name>
Content-Type: multipart/form-data
Form Field: file


返回 JSON：

{
  "status": "ok",  // ok | exists | uploading | error
  "url": "http://HOST:PORT/files/<md5>.<ext>"
}


文件访问

GET /files/<md5>.<ext>


可直接下载或在线播放（根据浏览器类型）

注意事项

文件统一命名为 <md5>.<ext>，避免重复上传

上传过程中临时文件以 tmp_ 前缀保存

如果同时多个客户端上传同一文件，服务器会提示 uploading 并阻止重复传输

客户端上传大文件时不会设置超时

服务器端会定期清理超过 24 小时的临时文件

文件扩展名从客户端上传的原始文件获取

目录结构
file_update/
├─ server.go
├─ client.go
├─ .env
├─ files/         # 上传的文件
└─ README.md

联系与支持

作者：松哥

联系方式：通过 Telegram 或私下沟通

建议：尽量在同一局域网或公网可达环境下运行，避免跨网络带宽瓶颈
