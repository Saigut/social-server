Social Server
-------------
[中文](extra/docs/README_CN.md)

# Introduction
This project implements an instant messaging server with the following main features:

- Supports single chat, group chat, read receipts, friend management, group management, and more
- Simple and reliable message synchronization logic
- Uses a write dissemination message handling approach
- The main logic module is a Stateless module, which facilitates dynamic scaling
- The API uses the gRpc framework. Thanks to the encapsulation of the main logic interface, it is easy to add support for other interface protocols, such as RESTful API, Websocket, etc.

# Preview
- URL： [letstalk.ink](https://letstalk.ink)

# How to Use
## Tool Dependencies
For generating gRpc code, you will need the following tools:

- [Protocol Buffers](https://grpc.io/docs/protoc-installation/)
- [protoc-gen-go](https://grpc.io/docs/languages/go/quickstart/#prerequisites)
- [protoc-gen-go-grpc](https://grpc.io/docs/languages/go/quickstart/#prerequisites)
- protoc-gen-ts

## Database and Cache
- Database: MySQL
  Please refer to `extra/docs/db_tables.sql` to set up the database and tables
- Cache: Redis
  A Redis server needs to be prepared

## Environment Variables
This project uses environment variables for runtime configuration. The required environment variables are as follows:  
MySQL configuration
```
DB_USER=xxxx
DB_PASSWORD=xxxx
DB_HOST=xxxx
DB_NAME=xxxx
```

Redis configuration
```
REDIS_HOST=xxxx
REDIS_PORT=6379
REDIS_PASSWORD=xxxx
REDIS_DB=0
```

You can use a `.env` file to configure environment variables, which are read from the program's working directory by default. You can also configure the `ENV_PATH` environment variable to specify the path to the `.env` file.

## Compilation and Execution
Compile:
```
go build -o social_server ./src/server.go
```

Run:  
This program listens on port 10080. It is recommended to run it in a container.

# Client Development
Please refer to the [api.proto](extra/protos/api.proto) file for integration.

Client example: [Saigut/LumenIM](https://github.com/Saigut/LumenIM)
