package database

import (
	"chat-demo-self/conf"
	"chat-demo-self/global"
	"context"
	"database/sql"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)

func ConnetDb() {
	info := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", conf.DbUser, conf.DbPassWord, conf.DbHost, conf.DbPort, conf.DbName)
	db, err := sql.Open(conf.Db, info)
	if err != nil {
		panic(err)
	}
	//复制给global变量
	global.Db = db
	//连接测试
	if err := global.Db.Ping(); err != nil {
		fmt.Println("MySQL connection error")
		os.Exit(-1)
	}
	fmt.Println("MySQL connection successful")

}

func ConnectRedis() {
	addr := fmt.Sprintf("%s:%s", conf.RedisAddr, conf.RedisPort)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: conf.RedisPassWord,
		DB:       conf.RDb,
	})
	//复制给global变量
	global.RedisClient = client
	// 测试连接
	_, err := global.RedisClient.Ping(context.Background()).Result()
	if err != nil {
		fmt.Println("Redis connection error")
		os.Exit(-1)
	}

	fmt.Println("Redis connection successful")
}

func ConnectMongo() {
	clientOptions := options.Client().ApplyURI("mongodb://" + conf.MongoDBAddr + ":" + conf.MongoDBPort)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	//复制给global变量
	global.MongoDBClient = client
	//连接测试
	err = global.MongoDBClient.Ping(context.Background(), nil)
	if err != nil {
		fmt.Println("MongoDB connection error")
		os.Exit(-1)
	}
	fmt.Println("MongoDB connection successful")
}
