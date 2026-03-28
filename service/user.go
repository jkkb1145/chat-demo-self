package service

import "chat-demo-self/dao"

type UserService struct {
	UserDAO *dao.UserDAO
}

func NewUserService() *UserService {
	return &UserService{
		UserDAO: dao.NewUserDAO(),
	}
}
