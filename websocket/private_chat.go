package ws

import (
	"chat-demo-self/global"
	"chat-demo-self/pkg/e"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

const month = 60 * 60 * 24 * 30 // 按照30天算一个月

// 发送消息的类型
type SendMsg struct {
	Type    int    `json:"type"`
	Content string `json:"content"`
}

// 回复的消息
type ReplyMsg struct {
	From    string `json:"from"`
	Code    int    `json:"code"`
	Content string `json:"content"`
}

// 用户类
type Client struct {
	ID     string
	SendID string
	Socket *websocket.Conn
	Send   chan []byte
}

// 广播类，包括广播内容和源用户
type Broadcast struct {
	Client  *Client
	Message []byte
	Type    int
}

// 用户管理
type ClientManager struct {
	Clients    map[string]*Client
	Broadcast  chan *Broadcast
	Reply      chan *Client
	Register   chan *Client
	Unregister chan *Client
}

// Message 信息转JSON (包括：发送者、接收者、内容)
type Message struct {
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Content   string `json:"content,omitempty"`
}

var Manager = ClientManager{ //实例化一个ClientManager管理各个客户端
	Clients:    make(map[string]*Client), // 参与连接的用户，出于性能的考虑，需要设置最大连接数
	Broadcast:  make(chan *Broadcast),
	Register:   make(chan *Client),
	Reply:      make(chan *Client),
	Unregister: make(chan *Client),
}

func createId(uid, toUid string) string {
	return uid + "->" + toUid
}

func WsHandler(c *gin.Context) {
	uid := c.Query("uid")     // 自己的id  1
	toUid := c.Query("toUid") // 对方的id  2
	conn, err := (&websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { // CheckOrigin解决跨域问题
			return true
		}}).Upgrade(c.Writer, c.Request, nil) // 升级成ws协议
	if err != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}
	// 创建一个用户实例
	client := &Client{
		ID:     createId(uid, toUid),
		SendID: createId(toUid, uid),
		Socket: conn,
		Send:   make(chan []byte),
	}
	// 用户注册到用户管理上
	Manager.Register <- client
	go client.Read(c)
	go client.Write(c)
}

func (c *Client) Read(ctx *gin.Context) {
	defer func() { // 避免忘记关闭，所以要加上close
		Manager.Unregister <- c //令Unregister管道就绪
		_ = c.Socket.Close()
	}()
	for {
		c.Socket.PongHandler()
		sendMsg := new(SendMsg)
		// _,msg,_:=c.Socket.ReadMessage()//读取String类型
		err := c.Socket.ReadJSON(&sendMsg) // 读取json格式，如果不是json格式，会报错
		if err != nil {
			log.Println("数据格式不正确", err)
			Manager.Unregister <- c
			_ = c.Socket.Close()
			break
		}
		if sendMsg.Type == 1 {
			r1, _ := global.RedisClient.Get(ctx, c.ID).Result()     //由于ID,SendID的格式是1->2,2->1, 所以自带标识作用
			r2, _ := global.RedisClient.Get(ctx, c.SendID).Result() //这里查找的就是Client连接信息中指定的两人间发消息的数量
			if r1 >= "3" && r2 == "" {                              // 防止骚扰消息发送三条无回复则禁止
				replyMsg := ReplyMsg{
					Code:    e.WebsocketLimit,
					Content: "在对方回复前无法发送更多消息",
				}
				msg, _ := json.Marshal(replyMsg)                                      //序列化JSON
				_ = c.Socket.WriteMessage(websocket.TextMessage, msg)                 //把信息回给客户端
				_, _ = global.RedisClient.Expire(ctx, c.ID, time.Hour*24*30).Result() // 防止重复骚扰，未建立连接刷新过期时间一个月
				//redis刷新后,r1的数值会重新变为0,这意味着发送方又可以发送三条消息进行骚扰,所以需要设置较长过期时间
				continue //进行下一个for循环
			} else {
				global.RedisClient.Incr(ctx, c.ID)                                      //将key=c.ID的数值增加1 会使上面r1加1
				_, _ = global.RedisClient.Expire(ctx, c.ID, time.Hour*24*30*3).Result() // 防止过快“分手”，建立连接三个月过期
			}
			log.Println(c.ID, "发送消息", sendMsg.Content)
			Manager.Broadcast <- &Broadcast{
				Client:  c,
				Message: []byte(sendMsg.Content),
			} //令广播管道就绪,Start函数中select可以转入广播管道执行
		}
	}
}

func (c *Client) Write(ctx *gin.Context) {
	defer func() {
		_ = c.Socket.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				_ = c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			log.Println(c.ID, "接受消息:", string(message))
			replyMsg := ReplyMsg{
				Code:    e.WebsocketSuccessMessage,
				Content: fmt.Sprintf("%s", string(message)),
			}
			msg, _ := json.Marshal(replyMsg)
			_ = c.Socket.WriteMessage(websocket.TextMessage, msg)
		}
	}
}
