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

func (c *ChatService) CreateNewGroup(member, groupID string) error {
	return c.ChatDAO.CreateNewGroup(member, groupID)
}
