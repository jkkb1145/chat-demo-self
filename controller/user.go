package controller

import (
	"chat-demo-self/model"
	"chat-demo-self/service"
	"github.com/gin-gonic/gin"
	"net/http"
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
	var user model.UserRegister
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":   -1,
			"msg":    "Error Occurred When ShouldBindJSON",
			"detail": err.Error(),
		})
		return
	}
	err := u.UserService.UserRegister(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":   -1,
			"msg":    "Error Occurred During The Execution Of The UserRegister function.",
			"detail": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":   0,
		"msg":    "The Program Ran Successfully Without Error.",
		"detail": nil,
	})
}
