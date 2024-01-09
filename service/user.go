package service

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"getcharzp.cn/define"
	"getcharzp.cn/helper"
	"getcharzp.cn/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetUserDetail
// @Tags 公共方法
// @Summary 问题详情
// @Param identity query string false "user_identity"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /user-detail [get]
func GetUserDetail(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "用户唯一identity不能为空",
		})
		return
	}

	data := new(models.UserBasic)
	err := models.DB.Omit("password").Where("identity = ?", identity).Find(&data).Error
	if err != nil {

		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "get user detail by identity" + identity + "Error!" + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code": 200,
		"data": data,
	})
}

// Login
// @Tags 公共方法
// @Summary 用户登录
// @Param username formData string false "username"
// @Param password formData string false "password"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /login [post]
func Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	if username == "" || password == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "必填信息为空",
		})
		return
	}
	password = helper.GetMd5(password)

	ub := new(models.UserBasic)
	err2 := models.DB.Where("name = ? AND password = ? ", username, password).First(&ub).Error
	if err2 != nil {
		if err2 == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "账户信息错误",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Get UserBasic Error" + err2.Error(),
		})
		return
	}
	//安装jwt生成token

	token, err := helper.GenerateToken(ub.Identity, ub.Name, ub.IsAdmin)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "generatetoken error:" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{ //登录成功就拿一个token
		"code": 200,
		"data": map[string]interface{}{
			"token": token,
		},
	})
}

// SendCode
// @Tags 公共方法
// @Summary 发送验证码
// @Param email formData string true "email"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /send-code [post]
func SendCode(c *gin.Context) {
	email := c.PostForm("email")
	if email == "" {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}
	code := helper.GetRand()
	models.RDB.Set(c, email, code, time.Second*300) //在redis中写入code 并且设置过期时间
	err := helper.SendCode(email, code)
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "send code error" + err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "验证码发送成功",
	})

}

// Register
// @Tags 公共方法
// @Summary 用户注册
// @Param username formData string true "username"
// @Param password formData string true "password"
// @Param mail formData string true "mail"
// @Param code formData string true "code"
// @Param phone formData string false "phone"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /register [post]
func Register(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	mail := c.PostForm("mail")
	usercode := c.PostForm("code")
	phone := c.PostForm("phone")
	if mail == "" || usercode == "" || username == "" || password == "" {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}
	//验证验证码是否正确
	sysCode, err := models.RDB.Get(c, mail).Result()
	if err != nil {
		log.Printf("Get code Error:%v", err)
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "请重新获取验证码",
		})
		return
	}
	if sysCode != usercode {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "验证码不正确",
		})
		return
	}

	//判断邮箱是否已存在
	var cnt int64
	err = models.DB.Where("mail = ? ", mail).Model(new((models.UserBasic))).Count(&cnt).Error
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "Get User Error" + err.Error(),
		})
		return
	}

	if cnt > 0 {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "该邮箱已被注册",
		})
		return
	}

	//新建用户插入
	//拿到这个struct
	userIdentity := helper.GetUUID()
	data := &models.UserBasic{
		Identity: userIdentity,
		Name:     username,
		Password: helper.GetMd5(password),
		Phone:    phone,
		Mail:     mail,
	}
	//正式插入数据库
	err = models.DB.Create(data).Error
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "Create User Error:" + err.Error(),
		})
		return
	}

	//生成token
	token, err := helper.GenerateToken(userIdentity, username, data.IsAdmin)
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "Generate Token Error:" + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code": 200,
		"msg": map[string]interface{}{
			"token": token,
		},
	})

}

// GetRankList
// @Tags 公共方法
// @Summary 问题列表
// @Param page query int false "page"
// @Param size query int false "size"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /rank-list [get]
func GetRankList(c *gin.Context) {
	size, _ := strconv.Atoi(c.DefaultQuery("size", define.DefaultSize))

	page, err := strconv.Atoi(c.DefaultQuery("page", define.DefaultPage))
	if err != nil {
		log.Println("strconv error", err)
	}

	page = (page - 1) * size //数据库的page起始位置

	var count int64

	list := make([]*models.UserBasic, 0)
	err = models.DB.Model(new(models.UserBasic)).Count(&count).Order("finish_problem_num DESC, submit_num ASC").Offset(page).Limit(size).Find(&list).Error
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "Get Rand List Error" + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"list":  list,
			"count": count,
		},
	})
}
