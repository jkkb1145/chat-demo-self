package controller

import "chat-demo-self/service"

type ChatController struct {
	ChatService *service.ChatService
}

func NewChatController() *ChatController {
	return &ChatController{
		ChatService: service.NewChatService(),
	}
}
