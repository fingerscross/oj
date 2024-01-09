package models

import (
	"log"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// init用于初始化数据库指针
var DB = Init() //数据库的gorm指针

var RDB = InitRedisDB() //redis 的指针

func Init() *gorm.DB {
	dsn := "root:20000620@tcp(127.0.0.1:3306)/gin_gorm_oj?charset=utf8mb4&parseTime=True&loc=Local"
	d, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("gorm init err: \n", err)
	}
	return d
}

func InitRedisDB() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

}
