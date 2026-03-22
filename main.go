package main

import (
	"chat-demo-self/conf"
	"chat-demo-self/database"
)

func main() {
	conf.InitConf()
	database.ConnetDb()
	database.ConnectRedis()
	database.ConnectMongo()
}
