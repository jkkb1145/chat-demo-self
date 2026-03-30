package service

import "chat-demo-self/dao"

type ChatService struct {
	ChatDAO *dao.ChatDAO
}

func NewChatService() *ChatService {
	return &ChatService{
		ChatDAO: dao.NewChatDAO(),
	}
}
