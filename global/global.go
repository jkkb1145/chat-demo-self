package global

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

var Db *sql.DB
var MongoDBClient *mongo.Client
var RedisClient *redis.Client

func GetDb() *sql.DB {
	return Db
}
