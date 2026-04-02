package router

import (
	"chat-demo-self/controller"
	ws "chat-demo-self/websocket"
	"fmt"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", func(c *gin.Context) {
			fmt.Println("success")
		})
	}
	user := v1.Group("/user")
	{
		user.POST("/user_register", controller.NewUserController().UserRegister)
		user.POST("/create_group", controller.NewChatController().CreateNewGroup)
	}
	chat := v1.Group("/chat")
	{
		chat.GET("/private_chat", ws.PrivateHandler)
		chat.GET("/group_chat", ws.GroupHanlder)
	}
	return r
}
