package controller

import (
	"chat-demo-self/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ChatController struct {
	ChatService *service.ChatService
}

func NewChatController() *ChatController {
	return &ChatController{
		ChatService: service.NewChatService(),
	}
}

func (c *ChatController) CreateNewGroup(ctx *gin.Context) {
	member := ctx.Query("member")
	groupID := ctx.Query("group_id")
	if err := c.ChatService.CreateNewGroup(member, groupID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":   -1,
			"msg":    "创建新群聊失败",
			"detail": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code":   0,
		"msg":    "群组创建成功",
		"detail": nil,
	})
}
