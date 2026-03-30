package dao

import (
	"chat-demo-self/global"
	"chat-demo-self/model"
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ChatDAO struct {
	ChatDB *sql.DB
}

func NewChatDAO() *ChatDAO {
	return &ChatDAO{
		ChatDB: global.GetDb(),
	}
}

type SendSortMsg struct {
	Content  string `json:"content"`
	Read     uint   `json:"read"`
	CreateAt int64  `json:"create_at"`
}

func (c *ChatDAO) InsertMsg(database string, id string, content string, read uint, expire int64) (err error) {
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
func (c *ChatDAO) GetMemberByGroupID(groupID string) (string, error) {
	sqlStr := "SELECT member FROM group WHERE id = ?"

	var resultStr string

	err := global.Db.QueryRow(sqlStr, groupID).Scan(&resultStr)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("未找到 ID = %d 的数据\n", groupID)
		} else {
			fmt.Printf("查询失败: %v\n", err)
		}
		return "", err
	}
	return resultStr, nil
}
