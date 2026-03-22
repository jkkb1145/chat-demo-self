package conf

import (
	"gopkg.in/ini.v1"
	"strconv"
)

var (
	AppMode  string
	HttpPort string

	Db         string
	DbHost     string
	DbPort     string
	DbUser     string
	DbPassWord string
	DbName     string

	RedisDb       string
	RedisAddr     string
	RedisPort     string
	RedisPassWord string
	RedisDbName   string
	RDb           int

	MongoDBName     string
	MongoDBAddr     string
	MongoDBPassWord string
	MongoDBPort     string
)

func InitConf() {
	file, err := ini.Load("D:\\GoDemo\\chat-demo-self\\conf\\config.ini")
	if err != nil {
		panic(err)
	}
	LoadServer(file)
	LoadMySQL(file)
	LoadRedis(file)
	LoadMongoDB(file)
}

func LoadServer(file *ini.File) {
	AppMode = file.Section("service").Key("AppMode").String()
	HttpPort = file.Section("service").Key("HttpPort").String()
}

func LoadMySQL(file *ini.File) {
	Db = file.Section("mysql").Key("Db").String()
	DbHost = file.Section("mysql").Key("DbHost").String()
	DbPort = file.Section("mysql").Key("DbPort").String()
	DbUser = file.Section("mysql").Key("DbUser").String()
	DbPassWord = file.Section("mysql").Key("DbPassWord").String()
	DbName = file.Section("mysql").Key("DbName").String()
}

func LoadRedis(file *ini.File) {
	RedisDb = file.Section("redis").Key("RedisDb").String()
	RedisAddr = file.Section("redis").Key("RedisAddr").String()
	RedisPort = file.Section("redis").Key("RedisPort").String()
	RedisPassWord = file.Section("redis").Key("RedisPassWord").String()
	RedisDbName = file.Section("redis").Key("RedisDbName").String()
	RDb, _ = strconv.Atoi(RedisDbName)
}

func LoadMongoDB(file *ini.File) {
	MongoDBName = file.Section("mongodb").Key("MongoDBName").String()
	MongoDBAddr = file.Section("mongodb").Key("MongoDBAddr").String()
	MongoDBPassWord = file.Section("mongodb").Key("MongoDBPassWord").String()
	MongoDBPort = file.Section("mongodb").Key("MongoDBPort").String()
}
