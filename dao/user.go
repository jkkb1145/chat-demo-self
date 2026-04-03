package dao

import (
	"chat-demo-self/global"
	"chat-demo-self/model"
	"database/sql"
	"fmt"
)

type UserDAO struct {
	UserDB *sql.DB
}

func NewUserDAO() *UserDAO {
	return &UserDAO{
		UserDB: global.GetDb(),
	}
}

func (u *UserDAO) UserRegister(user *model.UserRegister) error {
	query := "INSERT INTO user(UserName,Password) VALUE (?,?)"
	_, err := global.Db.Exec(query, user.UserName, user.PassWord)
	if err != nil {
		fmt.Printf("创建新用户失败 \n %v", err)
		return err
	}
	return nil
}
