package models

import "gorm.io/gorm"

type UserBasic struct {
	gorm.Model
	Identity  string `gorm:"column:identity;type:varchar(36);" json:"identity"`
	Name      string `gorm:"column:name;type:varchar(100);" json:"name"`
	Password  string `gorm:"column:password;type:varchar(32);"  json:"-"`
	Phone     string `gorm:"column:phone;type:varchar(20);" json:"phone"`
	Mail      string `gorm:"column:mail;type:varchar(100);"  json:"mail"`
	PassNum   int64  `gorm:"column:finish_problem_num;type:int(11);" json:"pass_num"`
	SubmitNum int64  `gorm:"column:submit_num;type:int(11);" json:"submit_num"`
	IsAdmin   int    `gorm:"column:is_admin;type:tinyint(1);"  json:"is_admin"`
}

func (table *UserBasic) TableName() string {
	return "user_basic"
}
