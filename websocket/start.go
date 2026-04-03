package ws

import (
	"chat-demo-self/conf"
	"chat-demo-self/dao"
	"chat-demo-self/pkg/e"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"strings"
)

func (manager *ClientManager) Start() {
	for {
		log.Println("<---监听管道通信--->") //定义管道时没有分配缓存区,select可以让准备好了的管道执行
		select {
		//监听单聊管道
		//新连接加入
		case conn := <-Manager.Register:
			log.Printf("建立新连接: %v", conn.ID)
			Manager.Clients[conn.ID] = conn //将新的连接存入存放Client的map中
			replyMsg := &ReplyMsg{
				Code:    e.WebsocketSuccess,
				Content: "已连接至服务器",
			}
			msg, _ := json.Marshal(replyMsg)
			_ = conn.Socket.WriteMessage(websocket.TextMessage, msg)
		//移除连接
		case conn := <-Manager.Unregister:
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
		//广播管道激活, 服务端向接收方发送信息
		case broadcast := <-Manager.Broadcast:
			message := broadcast.Message
			sendId := broadcast.Client.SendID
			flag := false                           // 接收方在线状况, 默认对方不在线
			for id, conn := range Manager.Clients { //在用户管理结构体的所有Client连接中找是否有反向连接,有则表示双方连接在线
				if id != sendId {
					continue
				}
				select {
				case conn.Send <- message: //此时的conn是接收方的ws连接, 激活了write函数中监听的Send管道, 将会把信息返回给接收方
					flag = true
				default: //Send管道被占用, 一般是发送消息过多服务端无法应对, 强制关闭
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
				err = dao.NewChatDAO().InsertMsg(conf.MongoDBName, id, string(message), 1, int64(3*month)) //写入mongodb保存聊天记录
				if err != nil {
					fmt.Println("写入MongoDB时错误: ", err)
				}
			} else {
				log.Println("对方不在线")
				replyMsg := ReplyMsg{
					Code:    e.WebsocketOfflineReply,
					Content: "对方不在线应答",
				}
				msg, err := json.Marshal(replyMsg)
				_ = broadcast.Client.Socket.WriteMessage(websocket.TextMessage, msg)
				err = dao.NewChatDAO().InsertMsg(conf.MongoDBName, id, string(message), 0, int64(3*month)) //0表示未读
				if err != nil {
					fmt.Println("写入MongoDB时错误: ", err)
				}
			}
		//监听群聊管道
		case conn := <-GManager.GRegister: // 建立连接
			log.Printf("%v加入群聊", conn.ID)
			GManager.GClients[conn.ID] = conn
			replyMsg := &ReplyMsg{
				Code:    e.WebsocketSuccess,
				Content: "已连接至服务器",
			}
			msg, _ := json.Marshal(replyMsg)
			_ = conn.Socket.WriteMessage(websocket.TextMessage, msg)
		case conn := <-GManager.GUnregister: // 断开连接
			log.Printf("%v断开连接", conn.ID)
			key := createId(conn.ID, conn.GroupID)
			if _, ok := GManager.GClients[key]; ok {
				replyMsg := &ReplyMsg{
					Code:    e.WebsocketEnd,
					Content: "连接已断开",
				}
				msg, _ := json.Marshal(replyMsg)
				_ = conn.Socket.WriteMessage(websocket.TextMessage, msg)
				close(conn.Send)
				delete(GManager.GClients, conn.ID)
			}
		case gbroadcast := <-GManager.GBroadcast:
			message := gbroadcast.Message
			id := gbroadcast.GroupClient.ID
			groupID := gbroadcast.GroupClient.GroupID
			member, err := dao.NewChatDAO().GetMemberByGroupID(groupID)
			if err != nil {
				fmt.Println("获取群聊成员错误: ", err)
			}
			strSlice := strings.Split(member, ",") //切分字符串提取成员
			for _, v := range strSlice {           //遍历群聊成员
				if v != id { //排除发送方自己
					for _, conn := range GManager.GClients { //向已经连接到服务端的群组成员转发消息
						if conn.ID != v || conn.GroupID != groupID {
							continue
						}
						select {
						case conn.Send <- message: //循环激活Send管道, 向群聊其他成员转发消息
						default:
							close(conn.Send)
							delete(Manager.Clients, conn.ID)
						}
					}
				}
			}
			//
			replyMsg := &ReplyMsg{
				Code:    e.WebsocketOnlineReply,
				Content: "已发送至群组",
			}
			msg, err := json.Marshal(replyMsg)
			_ = gbroadcast.GroupClient.Socket.WriteMessage(websocket.TextMessage, msg)
			err = dao.NewChatDAO().InsertMsg(conf.MongoDBName, groupID, string(message), 1, int64(3*month))
			if err != nil {
				fmt.Println("写入MongoDB时错误: ", err)
			}
		}
	}
}
