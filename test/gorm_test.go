package test

import (
	"fmt"
	"testing"

	"getcharzp.cn/models"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestGormTest(t *testing.T) {
	dsn := "root:20000620@tcp(127.0.0.1:3306)/gin_gorm_oj?charset=utf8mb4&parseTime=True&loc=Local"
	d, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	data := make([]*models.ProblemBasic, 0)
	err = d.Find(&data).Error
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range data {
		fmt.Printf("problem: %v\n", v)
	}

}
