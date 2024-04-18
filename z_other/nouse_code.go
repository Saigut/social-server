package z_other

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"golang.org/x/net/http2/h2c"
	"log"
	"net/http"
	"social_server/src/app/service/user_mgmt"
	"strings"
)

type Access struct {
}

func (p *Access) ProcessWebsocket() {

}

func (p *Access) SessUserLogin(param*SessUserLoginParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Access) SessUserLogout(param *SessUserLogoutParam) (err error) {
	return errors.New("method not implemented")
}

func (p *Access) GetUserSessList(param *GetUserSessListParam) (sessList []SessInfo, err error) {
	return nil, errors.New("method not implemented")
}

func (p *Access) FindUserSess(param *SessUserLogoutParam) (sess *SessInfo, err error) {
	return nil, errors.New("method not implemented")
}


var upgrader = websocket.Upgrader{
	// 允许所有CORS请求
	CheckOrigin: func(r *http.Request) bool { return true },
}

// 假设有一个 Protocol Buffer 消息结构
type MyProtoMessage struct {
	// Protocol Buffer 消息字段
}

func validateClient(r *http.Request) bool {
	// 示例：检查一个特定的 HTTP 头或者查询参数
	token := r.Header.Get("Authorization")
	// 这里应该有一些逻辑来验证 token 的有效性
	return token == "expected-token"
}

func sendMessages(conn *websocket.Conn, sendChan <-chan *MyProtoMessage) {
	for msg := range sendChan {
		responseBytes, err := proto.Marshal(msg)
		if err != nil {
			log.Println("Marshal error:", err)
			continue
		}
		if err := conn.WriteMessage(websocket.BinaryMessage, responseBytes); err != nil {
			log.Println("Write error:", err)
			// 可以选择在这里结束循环，如果写入失败
			break
		}
	}
}

// 处理 WebSocket 连接
func handleWebSocket(w http.ResponseWriter, r *http.Request) {

	// 在升级之前进行客户端验证
	if !validateClient(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	//defer conn.Close()

	// 创建一个通道用于发送消息
	sendChan := make(chan *MyJsonMessage)
	defer close(sendChan)

	// 启动发送消息的协程
	go sendMessages(conn, sendChan)

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		if messageType != websocket.BinaryMessage {
			// 向 conn 发送 error 回复
			log.Println("Not BinaryMessage")
			break
		}

		// 假设消息是 Protocol Buffer 格式
		var protoMsg MyProtoMessage
		err = proto.Unmarshal(message, &protoMsg)
		if err != nil {
			log.Println("Unmarshal error:", err)
			continue
		}

		// 处理接收到的消息
		// TODO: 将消息转发给其他模块

		// 响应客户端（可选）
		// 假设 response 是要发送回客户端的 Protocol Buffer 消息
		response := &MyProtoMessage{}
		responseBytes, err := proto.Marshal(response)
		if err != nil {
			log.Println("Marshal error:", err)
			continue
		}
		if err := conn.WriteMessage(messageType, responseBytes); err != nil {
			log.Println("Write error:", err)
			continue
		}
	}
}

func main_old() {
	http.HandleFunc("/ws", handleWebSocket)
	log.Println("Starting server on: 10080")
	err := http.ListenAndServe(":10080", nil)
	if err != nil {
		log.Fatal("ListenAndServe error:", err)
	}

	// 创建一个 HTTP 服务器，使用 h2c 来支持非 TLS 的 HTTP/2
	httpServer := &http.Server{
		Handler: h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
				grpcApiServer.ServeHTTP(w, r)
			} else {
				// 处理非 gRPC 请求
			}
		}), http2Server),
	}
}


type WebSocketConn struct {
	Conn    chan *websocket.Conn
	IsAuthed bool
}

type Api struct {
	UserMgmt user_mgmt.UsermgmtT
}


func (p *Api) startRpcServer(connChan chan *WebSocketConn) {
	for {
		conn := <-connChan
		go p.processWsConn(conn)
	}
}


type SociMsgJsonrpc struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      int             `json:"id"`
}

type JsonMsgCtx struct {
	recvChan chan SociMsgJsonrpc
	sendChan chan *string
}

func (p *Api) processWsConn(conn *WebSocketConn) {
	// 创建一个通道用于发送消息
	sendChan := make(chan *string)
	// 启动发送消息的协程
	go sendMessages(conn, sendChan)

	jsonMsgIdMap := make(map[int]JsonMsgCtx)

	for {
		messageType, message, err := conn.Conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			sendChan <- nil
			break
		}

		if messageType != websocket.BinaryMessage {
			// 向 conn 发送 error 回复
			log.Println("Not BinaryMessage")
			sendChan <- nil
			break
		}

		// 读取消息类型和长度
		msgType := binary.LittleEndian.Uint32(message[:4])
		msgLength := binary.LittleEndian.Uint32(message[4:8])

		// 根据消息类型处理消息
		if msgType == 100 { // 100 代表 JSON 类型
			if int(msgLength) != len(message[8:]) {
				log.Println("Invalid message length")
				continue
			}

			if conn.IsAuthed {
				p.processJsonMsgs(message[8:], sendChan, &jsonMsgIdMap)
			} else {
				p.processUnauthJsonMsg(message[8:])
			}


		} else {
			log.Println("Unsupported message type " + fmt.Sprintf("%d", msgType))
			// 发送个 error 回复，msgType 1： 不支持的消息类型。
			// TODO  sendChan <- xxx
			break
		}
	}
}

func (p *Api) processSessMsg() {

}

func (p *Api) processUnauthJsonMsg(jsonStr *[]byte) (err error) {
	var yourMessage YourMessage
	err = json.Unmarshal(*jsonStr, &yourMessage) // 解引用 jsonStr 指针
	if err != nil {
		errStr := "Error unmarshalling JSON: " + err.Error()
		log.Println(errStr)
		return errors.New(errStr)
	}

	// 处理 yourMessage
	fmt.Printf("Received message: %+v\n", yourMessage)
	// 回复消息
	// ...

	return nil
}

func (p *Api) processJsonMsgs(jsonStr *[]byte, sendChan chan *string, idMap *map[int]JsonMsgCtx) (err error) {
	var jsonMsg SociMsgJsonrpc
	err = json.Unmarshal(*jsonStr, &jsonMsg) // 解引用 jsonStr 指针
	if err != nil {
		errStr := "Error unmarshalling JSON: " + err.Error()
		log.Println(errStr)
		return errors.New(errStr)
	}

	value, exists := (*idMap)[jsonMsg.ID]
	if exists {
		value.recvChan <- jsonMsg
	} else {
		value = JsonMsgCtx{}
		value.sendChan = sendChan
		go p.processJsonMsg(value)
		(*idMap)[jsonMsg.ID] = value
		value.recvChan <- jsonMsg
	}

	return nil
}

func (p *Api) processJsonMsg(idCtx JsonMsgCtx) (err error) {

	for {
		recvMsg := <- idCtx.recvChan

		switch recvMsg.Method {
		case "logout":
			return p.processUmLogout(recvMsg.Params, idCtx)
		case "unregister":
			return p.processUmUnregister(recvMsg.Params, idCtx)
		case "add_friends":
			return p.processUmAddFriends(recvMsg.Params, idCtx)
		case "del_friends":
			return p.processUmDelFriends(recvMsg.Params, idCtx)
		case "list_friends":
			return p.processUmListFriends(recvMsg.Params, idCtx)
		case "put_post":
			return p.processPostPutPost(recvMsg.Params, idCtx)
		case "get_video_hls":
			return p.processPostGetVideoHls(recvMsg.Params, idCtx)
		case "get_post_list":
			return p.processPostGetPostList(recvMsg.Params, idCtx)
		case "get_post_metadata":
			return p.processPostGetPostMetadata(recvMsg.Params, idCtx)
		case "get_explorer_post_list":
			return p.processPostGetExplorerPostList(recvMsg.Params, idCtx)
		case "get_likes":
			return p.processPostGetLikes(recvMsg.Params, idCtx)
		case "do_like":
			return p.processPostDoLike(recvMsg.Params, idCtx)
		case "undo_like":
			return p.processPostUndoLike(recvMsg.Params, idCtx)
		case "get_comments":
			return p.processPostGetComments(recvMsg.Params, idCtx)
		case "add_comment":
			return p.processPostAddComment(recvMsg.Params, idCtx)
		case "del_comment":
			return p.processPostDelComment(recvMsg.Params, idCtx)
		default:
			errStr := "Unknown method: " + recvMsg.Method
			fmt.Println(errStr)
			return errors.New(errStr)
		}
	}
}

func (p *Api) processUnauthPbMsg() {

}

func (p *Api) processPbMsg() {

}

// User Management (UM) APIs
func (p *Api) processUmRegister(params json.RawMessage, idCtx JsonMsgCtx) (err error) {
	return errors.New("method not implemented")
}

func (p *Api) processUmUnregister(params json.RawMessage, idCtx JsonMsgCtx) (err error) {
	var jsonParams user_mgmt.UmUnregisterParam
	err = json.Unmarshal(params, &jsonParams)
	if err != nil {
		errStr := "Error unmarshalling JSON: " + err.Error()
		log.Println(errStr)
		return errors.New(errStr)
	}
	err = p.UserMgmt.Unregister(&jsonParams)
	if err != nil {
		/// todo: error response
		return err
	}
	/// todo: response
	return nil
}

func (p *Api) processUmLogin(params json.RawMessage, idCtx JsonMsgCtx) (err error) {
	return errors.New("method not implemented")
}

func (p *Api) processUmLogout(params json.RawMessage, idCtx JsonMsgCtx) (err error) {
	var jsonParams SociMsgJsonrpc
	err = json.Unmarshal(params, &jsonParams)
	if err != nil {
		errStr := "Error unmarshalling JSON: " + err.Error()
		log.Println(errStr)
		return errors.New(errStr)
	}
}

func (p *Api) processUmAddFriends(params json.RawMessage, idCtx JsonMsgCtx) (err error) {

}

func (p *Api) processUmDelFriends(params json.RawMessage, idCtx JsonMsgCtx) (err error) {
	return errors.New("method not implemented")
}

func (p *Api) processUmListFriends(params json.RawMessage, idCtx JsonMsgCtx) (err error) {
	return errors.New("method not implemented")
}

// Message (Msg) APIs
func (p *Api) processChatGetChatMsg(params json.RawMessage, idCtx JsonMsgCtx) (err error) {
	return errors.New("method not implemented")
}

func (p *Api) processChatGetChatMsgHistWith(params json.RawMessage, idCtx JsonMsgCtx) (err error) {
	return errors.New("method not implemented")
}

func (p *Api) processChatSendChatMsgTo(params json.RawMessage, idCtx JsonMsgCtx) (err error) {
	return errors.New("method not implemented")
}

// Video APIs
func (p *Api) processPostPutPost(params json.RawMessage, idCtx JsonMsgCtx) (err error) {
	return errors.New("method not implemented")
}

func (p *Api) processPostGetVideoHls(params json.RawMessage, idCtx JsonMsgCtx) (err error) {
	return errors.New("method not implemented")
}

func (p *Api) processPostGetPostList(params json.RawMessage, idCtx JsonMsgCtx) (err error) {
	return errors.New("method not implemented")
}

func (p *Api) processPostGetPostMetadata(params json.RawMessage, idCtx JsonMsgCtx) (err error) {
	return errors.New("method not implemented")
}

func (p *Api) processPostGetExplorerPostList(params json.RawMessage, idCtx JsonMsgCtx) (err error) {
	return errors.New("method not implemented")
}

func (p *Api) processPostGetLikes(params json.RawMessage, idCtx JsonMsgCtx) (err error) {
	return errors.New("method not implemented")
}

func (p *Api) processPostDoLike(params json.RawMessage, idCtx JsonMsgCtx) (err error) {
	return errors.New("method not implemented")
}

func (p *Api) processPostUndoLike(params json.RawMessage, idCtx JsonMsgCtx) (err error) {
	return errors.New("method not implemented")
}

func (p *Api) processPostGetComments(params json.RawMessage, idCtx JsonMsgCtx) (err error) {
	return errors.New("method not implemented")
}

func (p *Api) processPostAddComment(params json.RawMessage, idCtx JsonMsgCtx) (err error) {
	return errors.New("method not implemented")
}

func (p *Api) processPostDelComment(params json.RawMessage, idCtx JsonMsgCtx) (err error) {
	return errors.New("method not implemented")
}
