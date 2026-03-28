package dao

import (
	"chat-demo-self/global"
	"chat-demo-self/model"
	"context"
	"time"
)

type SendSortMsg struct {
	Content  string `json:"content"`
	Read     uint   `json:"read"`
	CreateAt int64  `json:"create_at"`
}

func InsertMsg(database string, id string, content string, read uint, expire int64) (err error) {
	collection := global.MongoDBClient.Database(database).Collection(id)
	comment := model.Trainer{
		Content:   content,
		StartTime: time.Now().Unix(),
		EndTime:   time.Now().Unix() + expire,
		Read:      read,
	}
	_, err = collection.InsertOne(context.TODO(), comment)
	return
}
