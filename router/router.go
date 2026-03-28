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
		ctl := controller.NewUserController()
		user.POST("/register", ctl.UserRegister)
		user.GET("/ws", ws.WsHandler)
	}
	return r
}
