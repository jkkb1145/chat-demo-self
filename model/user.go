package model

type User struct {
	UserName string `json:"user_name"`
	PassWord string `json:"pass_word"`
}

type UserRegister struct {
	UserName string `json:"user_name"`
	PassWord string `json:"pass_word"`
}
