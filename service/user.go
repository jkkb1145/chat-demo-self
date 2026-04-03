package service

import (
	"chat-demo-self/dao"
	"chat-demo-self/model"
)

type UserService struct {
	UserDAO *dao.UserDAO
}

func NewUserService() *UserService {
	return &UserService{
		UserDAO: dao.NewUserDAO(),
	}
}
func (u *UserService) UserRegister(user *model.UserRegister) error {
	return u.UserDAO.UserRegister(user)
}
