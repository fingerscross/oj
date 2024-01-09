package models

import (
	"gorm.io/gorm"
)

type ProblemBasic struct {
	gorm.Model
	Identity          string             `gorm:"column:identity;type:varchar(36);"			json:"identity"`
	ProblemCategories []*ProblemCategory `gorm:"foreignKey:problem_id;references:id"` //关联问题分类表  即 需要problembasic的id列作为 引用表（关联表）的problem_id列 （外键）
	Title             string             `gorm:"column:title;type:varchar(255);" 			json:"title"`
	Content           string             `gorm:"column:content;type:text;"  				json:"content"`
	MaxMem            int                `gorm:"column:max_mem;type:int(11);"  			json:"max_mem"`
	MaxRuntime        int                `gorm:"column:max_runtime;type:int(11);" 			json:"max_runtime"`
	TestCases         []*TestCase        `gorm:"foreignKey:problem_identity;references:identity" json:"test_cases"`
	PassNum           int64              `gorm:"column:pass_num;type:int(11);" json:"pass_num"`
	SubmitNum         int64              `gorm:"column:submit_num;type:int(11);" json:"submit_num"`
}

// 建成表  model里一般是对数据库的查找 简单的查找就放在service
func (table *ProblemBasic) TableName() string {

	return "problem_basic"
}

func GetProblemList(keyword, categoryIdentity string) *gorm.DB {
	tx := DB.Model(new(ProblemBasic)).Preload("ProblemCategories").Preload("ProblemCategories.CategoryBasic").
		Where("title like ? OR content like ?", "%"+keyword+"%", "%"+keyword+"%") //查找表  模糊查询

	if categoryIdentity != "" { //关联了两个表
		tx.Joins("RIGHT JOIN problem_category pc on pc.problem_id = problem_basic.id").
			Where("pc.category_id = (SELECT cb.id FROM category_basic cb WHERE cb.identity = ?)", categoryIdentity)
	}

	return tx
}
