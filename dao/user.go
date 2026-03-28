package dao

import (
	"chat-demo-self/global"
	"database/sql"
)

type UserDAO struct {
	UserDB *sql.DB
}

func NewUserDAO() *UserDAO {
	return &UserDAO{
		UserDB: global.GetDb(),
	}
}
