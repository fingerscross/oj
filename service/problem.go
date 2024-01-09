package service

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"getcharzp.cn/define"
	"getcharzp.cn/helper"
	"getcharzp.cn/models"
	"github.com/gin-gonic/gin"

	"gorm.io/gorm"
)

// GetProblemList
// @Tags 公共方法
// @Summary 问题列表
// @Param page query int false "page"
// @Param size query int false "size"
// @Param keyword query string false "keyword"
// @Param category_identity query string false "category_identity"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /problem-list [get]
func GetProblemList(c *gin.Context) {
	size, _ := strconv.Atoi(c.DefaultQuery("size", define.DefaultSize))

	page, err := strconv.Atoi(c.DefaultQuery("page", define.DefaultPage))
	if err != nil {
		log.Println("strconv error", err)
	}

	page = (page - 1) * size //数据库的page起始位置

	var count int64

	keyword := c.Query("keyword")
	categoryIdentity := c.Query("category_identity")

	list := make([]*models.ProblemBasic, 0)

	tx := models.GetProblemList(keyword, categoryIdentity) //拿到db指针
	err = tx.Count(&count).Omit("content").Offset(page).Limit(size).Find(&list).Error
	//limit即分页大小
	// offset指定开始返回记录前要跳过的记录数。
	//count指获取模型的记录数
	//Omit 不显示某些信息
	if err != nil {
		log.Println("get problem list error")
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": map[string]interface{}{
		"list":  list,
		"count": count,
	}})

}

// GetProblemDetail
// @Tags 公共方法
// @Summary 问题详情
// @Param identity query string false "problem_identity"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /problem-detail [get]
func GetProblemDetail(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "问题identity不能为空",
		})
		return
	}
	pb := new(models.ProblemBasic)
	err := models.DB.Where("identity = ?", identity).Preload("ProblemCategories").Preload("ProblemCategories.CategoryBasic").
		First(&pb).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(200, gin.H{

				"code": -1,
				"msg":  "问题不存在",
			})
			return
		}

		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "problemdetailerror" + err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"data": pb,
	})
}

// ProblemCreate
// @Tags 管理员私有方法
// @Summary 问题创建
// @Param authorization header string true "authorization"
// @Param title formData string true "title"
// @Param content formData string true "content"
// @Param max_runtime formData int false "max_runtime"
// @Param max_mem formData int false "max_mem"
// @Param category_ids formData []string false "category_ids" collectionFormat(multi)
// @Param test_cases formData []string true "test_cases" collectionFormat(multi)
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /admin/problem-create [post]
func ProblemCreate(c *gin.Context) {
	title := c.PostForm("title")
	content := c.PostForm("content")
	maxRuntime, _ := strconv.Atoi(c.PostForm("max_runtime"))
	maxMem, _ := strconv.Atoi(c.PostForm("max_mem"))
	categoryIds := c.PostFormArray("category_ids")
	testCases := c.PostFormArray("test_cases")
	if title == "" || content == "" || len(testCases) == 0 || len(categoryIds) == 0 || maxRuntime == 0 || maxMem == 0 {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "参数不能为空",
		})
		return
	}

	identity := helper.GetUUID()
	data := &models.ProblemBasic{
		Identity:   identity,
		Title:      title,
		Content:    content,
		MaxRuntime: maxRuntime,
		MaxMem:     maxMem,
	}

	//处理分类
	categoryBasics := make([]*models.ProblemCategory, 0)
	for _, id := range categoryIds {
		categoryId, _ := strconv.Atoi(id)
		categoryBasics = append(categoryBasics, &models.ProblemCategory{
			ProblemId:  data.ID,
			CategoryId: uint(categoryId),
		})
	}

	data.ProblemCategories = categoryBasics //因为problembasic里有problemcategories的关联

	//处理test
	testCaseBasics := make([]*models.TestCase, 0)
	for _, testCase := range testCases {
		caseMap := make(map[string]string)
		err := json.Unmarshal([]byte(testCase), &caseMap)

		if err != nil {
			c.JSON(200, gin.H{
				"code": -1,
				"msg":  "测试用例格式错误",
			})
			return
		}
		if _, ok := caseMap["input"]; !ok {
			c.JSON(200, gin.H{
				"code": -1,
				"msg":  "测试用例格式错误",
			})
			return
		}
		if _, ok := caseMap["output"]; !ok {
			c.JSON(200, gin.H{
				"code": -1,
				"msg":  "测试用例格式错误",
			})
			return
		}

		testCaseBasic := &models.TestCase{
			Identity:        helper.GetUUID(),
			ProblemIdentity: identity,
			Input:           caseMap["input"],
			Output:          caseMap["output"],
		}
		testCaseBasics = append(testCaseBasics, testCaseBasic)
	}
	data.TestCases = testCaseBasics

	err := models.DB.Create(data).Error
	if err != nil {

		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "创建失败" + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"identity": data.Identity,
		},
	})

}

// ProblemModify
// @Tags 管理员私有方法
// @Summary 问题修改
// @Param authorization header string true "authorization"
// @Param identity formData string true "identity"
// @Param title formData string true "title"
// @Param content formData string true "content"
// @Param max_runtime formData int false "max_runtime"
// @Param max_mem formData int false "max_mem"
// @Param category_ids formData []string false "category_ids" collectionFormat(multi)
// @Param test_cases formData []string false "test_cases" collectionFormat(multi)
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /admin/problem-modify [put]
func ProblemModify(c *gin.Context) {
	identity := c.PostForm("identity")
	title := c.PostForm("title")
	content := c.PostForm("content")
	maxRuntime, _ := strconv.Atoi(c.PostForm("max_runtime"))
	maxMem, _ := strconv.Atoi(c.PostForm("max_mem"))
	categoryIds := c.PostFormArray("category_ids")
	testCases := c.PostFormArray("test_cases")

	if identity == "" || title == "" || content == "" || len(testCases) == 0 || len(categoryIds) == 0 || maxRuntime == 0 || maxMem == 0 {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "参数不能为空",
		})
		return
	}

	//事务
	if err := models.DB.Transaction(func(tx *gorm.DB) error {
		//问题基础信息保存 porblem_basic
		problemBasic := &models.ProblemBasic{
			Identity:   identity,
			Title:      title,
			Content:    content,
			MaxRuntime: maxRuntime,
			MaxMem:     maxMem,
		}
		err := tx.Where("identity = ?", identity).Updates(problemBasic).Error
		if err != nil {
			return err
		}
		//查询问题详情 因为要查id
		err = tx.Where("identity = ?", identity).Find(problemBasic).Error
		if err != nil {
			return err
		}

		//关联问题分类的更新
		//1、删除已存在的关联关系
		err = tx.Where("problem_id = ? ", problemBasic.ID).Delete(new(models.ProblemCategory)).Error
		if err != nil {
			return err
		}
		//2、新增新的关联关系
		pcs := make([]*models.ProblemCategory, 0)
		for _, id := range categoryIds {
			intID, _ := strconv.Atoi(id)
			pcs = append(pcs, &models.ProblemCategory{
				ProblemId:  problemBasic.ID,
				CategoryId: uint(intID),
			})
		}

		err = tx.Create(&pcs).Error
		if err != nil {
			return err
		}
		//关联测试案例的更新
		//1、删除已存在的测试案例
		err = tx.Where("problem_identity = ?", identity).Delete(new(models.TestCase)).Error
		if err != nil {
			return err
		}
		//2、增加新的关联关系
		tc := make([]*models.TestCase, 0)
		for _, testcase := range testCases {
			casemap := make(map[string]string)
			err = json.Unmarshal([]byte(testcase), &casemap)
			if err != nil {
				return err
			}

			if _, ok := casemap["input"]; !ok {
				return errors.New("测试案例input格式错误")
			}
			if _, ok := casemap["output"]; !ok {
				return errors.New("测试案例output格式错误")
			}

			tc = append(tc, &models.TestCase{
				Identity:        helper.GetUUID(),
				ProblemIdentity: identity,
				Input:           casemap["input"],
				Output:          casemap["output"],
			})
		}
		err = tx.Create(tc).Error
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "Problem modify wrong" + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "更新成功",
	})

}
