package ws

import (
	"chat-demo-self/conf"
	"chat-demo-self/dao"
	"chat-demo-self/pkg/e"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

func (manager *ClientManager) Start() {
	for {
		log.Println("<---监听管道通信--->") //定义管道时没有分配缓存区,select可以让准备好了的管道执行
		select {
		case conn := <-Manager.Register: // 建立连接
			log.Printf("建立新连接: %v", conn.ID)
			Manager.Clients[conn.ID] = conn //将新的连接存入存放Client的map中
			replyMsg := &ReplyMsg{
				Code:    e.WebsocketSuccess,
				Content: "已连接至服务器",
			}
			msg, _ := json.Marshal(replyMsg)
			_ = conn.Socket.WriteMessage(websocket.TextMessage, msg)
		case conn := <-Manager.Unregister: // 断开连接
			log.Printf("连接失败:%v", conn.ID)
			if _, ok := Manager.Clients[conn.ID]; ok {
				replyMsg := &ReplyMsg{
					Code:    e.WebsocketEnd,
					Content: "连接已断开",
				}
				msg, _ := json.Marshal(replyMsg)
				_ = conn.Socket.WriteMessage(websocket.TextMessage, msg)
				close(conn.Send)
				delete(Manager.Clients, conn.ID)
			}
		//广播信息
		case broadcast := <-Manager.Broadcast:
			message := broadcast.Message
			sendId := broadcast.Client.SendID
			flag := false                           // 默认对方不在线
			for id, conn := range Manager.Clients { //在用户管理结构体的所有Client连接中找是否有反向连接,有则表示双方连接在线
				if id != sendId {
					continue
				}
				select {
				case conn.Send <- message:
					flag = true
				default:
					close(conn.Send)
					delete(Manager.Clients, conn.ID)
				}
			}
			id := broadcast.Client.ID
			if flag {
				log.Println("对方在线应答")
				replyMsg := &ReplyMsg{
					Code:    e.WebsocketOnlineReply,
					Content: "对方在线应答",
				}
				msg, err := json.Marshal(replyMsg)
				_ = broadcast.Client.Socket.WriteMessage(websocket.TextMessage, msg)
				err = dao.InsertMsg(conf.MongoDBName, id, string(message), 1, int64(3*month)) //写入mongodb保存
				if err != nil {
					fmt.Println("InsertOneMsg Err", err)
				}
			} else {
				log.Println("对方不在线")
				replyMsg := ReplyMsg{
					Code:    e.WebsocketOfflineReply,
					Content: "对方不在线应答",
				}
				msg, err := json.Marshal(replyMsg)
				_ = broadcast.Client.Socket.WriteMessage(websocket.TextMessage, msg)
				err = dao.InsertMsg(conf.MongoDBName, id, string(message), 0, int64(3*month)) //0表示未读
				if err != nil {
					fmt.Println("InsertOneMsg Err", err)
				}
			}
		}
	}
}
