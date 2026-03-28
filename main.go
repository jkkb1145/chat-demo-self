package main

import (
	"chat-demo-self/conf"
	"chat-demo-self/database"
	"chat-demo-self/router"
	ws "chat-demo-self/websocket"
)

func main() {
	conf.InitConf()
	database.ConnetDb()
	database.ConnectRedis()
	database.ConnectMongo()
	r := router.InitRouter()
	go ws.Manager.Start()
	_ = r.Run(conf.HttpPort)
}
