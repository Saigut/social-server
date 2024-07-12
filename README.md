Social Server
-------------
[English](extra/docs/README_en.md)

# 简介
此项目实现了一个即时聊天服务器，主要特点如下：

- 支持单聊、群聊、已读回执、好友管理、群组管理等功能
- 消息同步逻辑简洁、可靠
- 采用写扩散的消息处理方式
- 主逻辑模块为 Stateless 模块，便于支持动态伸缩
- Api 使用了 gRpc 框架。得益于主逻辑的接口封装，可以方便地添加对其它接口协议，如 RESTful API、Websocket 等的支持

# 项目预览
- 地址： [letstalk.ink](https://letstalk.ink)

# 如何使用
## 工具依赖
对于 gRpc 代码的生成，你将会用到以下工具：

- [Protocol Buffers](https://grpc.io/docs/protoc-installation/)
- [protoc-gen-go](https://grpc.io/docs/languages/go/quickstart/#prerequisites)
- [protoc-gen-go-grpc](https://grpc.io/docs/languages/go/quickstart/#prerequisites)
- protoc-gen-ts

## 数据库与缓存
- 数据库：MySQL  
  请参见 `extra/docs/db_tables.sql` 建立数据库及表格
- 缓存：Redis  
  需要准备好 Redis 服务器

## 环境变量
本项目使用环境变量来配置运行时，需要配置的环境变量如下：  
MySQL 配置
```
DB_USER=xxxx
DB_PASSWORD=xxxx
DB_HOST=xxxx
DB_NAME=xxxx
```

Redis 配置
```
REDIS_HOST=xxxx
REDIS_PORT=6379
REDIS_PASSWORD=xxxx
REDIS_DB=0
```

你可以使用 `.env` 文件来配置环境变量，默认从程序工作目录读取。你也可以配置 `ENV_PATH` 环境变量来指定 `.env` 文件的路径。

## 编译运行
编译：
```
go build -o social_server ./src/server.go
```

运行：  
此程序监听的端口为 10080。建议以容器的方式运行。

# 客户端开发
请参考 [api.proto](../protos/api.proto) 文件进行对接。

客户端示例：[Saigut/LumenIM](https://github.com/Saigut/LumenIM)