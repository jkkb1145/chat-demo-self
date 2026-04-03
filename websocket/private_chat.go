package ws

import (
	"chat-demo-self/conf"
	"chat-demo-self/dao"
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

// 客户端发送的消息的结构体
type SendMsg struct {
	Type    int    `json:"type"`
	Content string `json:"content"`
}

// 服务端回复的消息
type ReplyMsg struct {
	From    string `json:"from"`
	Code    int    `json:"code"`
	Content string `json:"content"`
}

// 连接的用户
type Client struct {
	ID     string          //包含信息:谁发给谁
	SendID string          //包含信息:谁接受谁
	Socket *websocket.Conn //ws连接
	Send   chan []byte     //服务端给接收方发的消息
}

// 广播, 发送信息
type Broadcast struct {
	Client  *Client //发送方的ws连接
	Message []byte  //发送方发送的内容
	Type    int     //类型
}

// 用户管理, 存放建立的连接与各个管道, 由于管道没有设置缓存区, start函数会持续监听直到管道被写入内容
type ClientManager struct {
	Clients    map[string]*Client //所有连接到服务端的客户端的ws连接
	Broadcast  chan *Broadcast    //广播管道, 负责实际的发送消息逻辑
	Reply      chan *Client       //未使用
	Register   chan *Client       //新的用户连接注册
	Unregister chan *Client       //用户断开连接
}

// Message 信息转JSON (包括：发送者、接收者、内容)
type Message struct {
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Content   string `json:"content,omitempty"`
}

// 实例化一个ClientManager管理各个客户端
var Manager = ClientManager{
	Clients:    make(map[string]*Client),
	Broadcast:  make(chan *Broadcast),
	Register:   make(chan *Client),
	Reply:      make(chan *Client),
	Unregister: make(chan *Client),
}

// 将连接的ID设置为发送方->接收方, SendID设为接收方->发送方
func createId(uid, toUid string) string {
	return uid + "->" + toUid
}

func PrivateHandler(c *gin.Context) {
	uid := c.Query("uid")     // 自己的id  1
	toUid := c.Query("toUid") // 对方的id  2
	// 升级成ws协议
	conn, err := (&websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { // CheckOrigin解决跨域问题
			return true
		}}).Upgrade(c.Writer, c.Request, nil) //解决跨域问题
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
	// 将创建好的用户注册到用户管理上, start监听的管道就绪
	Manager.Register <- client
	//开启协程, 与主函数同时运行
	go client.Read(c)
	go client.Write(c)
}

func (c *Client) Read(ctx *gin.Context) {
	defer func() { //在函数运行结束后, 关闭通道与ws连接
		Manager.Unregister <- c //令Unregister管道就绪, start函数将该连接注销
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
			Manager.Broadcast <- &Broadcast{ //start管道中监听的广播管道就绪, 开始发送消息
				Client:  c,
				Message: []byte(sendMsg.Content),
			}
		} else if sendMsg.Type == 2 { //查询历史记录,
			allMsg, err := dao.NewChatDAO().GetMsg(conf.MongoDBName, c.ID)
			if err != nil {
				log.Println("查询历史记录失败: ", err)
				Manager.Unregister <- c
				_ = c.Socket.Close()
				break
			}
			for _, v := range *allMsg {
				msg, _ := json.Marshal(v)
				_ = c.Socket.WriteMessage(websocket.TextMessage, msg)
			}
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
