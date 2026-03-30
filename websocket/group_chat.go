package ws

import (
	"chat-demo-self/pkg/e"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type GroupClient struct {
	ID      string
	GroupID string
	Socket  *websocket.Conn
	Send    chan []byte
}

type GroupBroadcast struct {
	GroupClient *GroupClient
	Message     []byte
	Type        int
}

type GClientManager struct {
	GClients    map[string]*GroupClient
	GBroadcast  chan *GroupBroadcast
	GRegister   chan *GroupClient
	GUnregister chan *GroupClient
}

var GManager = GClientManager{
	GClients:    make(map[string]*GroupClient),
	GBroadcast:  make(chan *GroupBroadcast),
	GRegister:   make(chan *GroupClient),
	GUnregister: make(chan *GroupClient),
}

func GroupHanlder(c *gin.Context) {
	id := c.Query("id")
	groupID := c.Query("group_id")
	conn, err := (&websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { // CheckOrigin解决跨域问题
			return true
		}}).Upgrade(c.Writer, c.Request, nil) // 升级成ws协议
	if err != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}
	client := &GroupClient{
		ID:      id,
		GroupID: groupID,
		Socket:  conn,
		Send:    make(chan []byte),
	}
	GManager.GRegister <- client
	go client.GRead(c)
}

func (g *GroupClient) GRead(ctx *gin.Context) {
	defer func() { // 避免忘记关闭，所以要加上close
		GManager.GUnregister <- g //令Unregister管道就绪
		_ = g.Socket.Close()
	}()
	sendMsg := new(SendMsg)
	for {
		err := g.Socket.ReadJSON(&sendMsg)
		if err != nil {
			log.Println("数据格式不正确", err)
			GManager.GUnregister <- g
			_ = g.Socket.Close()
			break
		}
		if sendMsg.Type == 2 {
			log.Println(g.ID, "发送消息", sendMsg.Content, "到群组", g.GroupID)
			GManager.GBroadcast <- &GroupBroadcast{
				GroupClient: g,
				Message:     []byte(sendMsg.Content),
			}
		}
	}
}
func (g *GroupClient) GWrite() {
	defer func() {
		_ = g.Socket.Close()
	}()
	for {
		select {
		case message, ok := <-g.Send:
			if !ok {
				_ = g.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			log.Println("群组", g.GroupID, "接受消息:", string(message))
			replyMsg := ReplyMsg{
				Code:    e.WebsocketSuccessMessage,
				Content: fmt.Sprintf("%s", string(message)),
			}
			msg, _ := json.Marshal(replyMsg)
			_ = g.Socket.WriteMessage(websocket.TextMessage, msg)
		}
	}
}
