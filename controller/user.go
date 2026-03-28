package controller

import (
	"chat-demo-self/service"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	UserService *service.UserService
}

func NewUserController() *UserController {
	return &UserController{
		UserService: service.NewUserService(),
	}
}

func (u *UserController) UserRegister(c *gin.Context) {

}
