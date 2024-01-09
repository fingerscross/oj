package service

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"time"

	"getcharzp.cn/define"
	"getcharzp.cn/helper"
	"getcharzp.cn/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetSubmitList
// @Tags 公共方法
// @Summary 问题列表
// @Param page query int false "page"
// @Param size query int false "size"
// @Param problem_identity query string false "problem_identity"
// @Param user_identity query string false "user_identity"
// @Param status query int false "status"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /submit-list [get]
func GetSubmitList(c *gin.Context) {
	size, _ := strconv.Atoi(c.DefaultQuery("size", define.DefaultSize))

	page, err := strconv.Atoi(c.DefaultQuery("page", define.DefaultPage))
	if err != nil {
		log.Println("strconv error", err)
	}
	page = (page - 1) * size //数据库的page起始位置

	var count int64
	list := make([]models.SubmitBasic, 0) //看看和 new有什么区别？  new只能拿一条数据  make可以全拿出来
	//list := new(models.SubmitBasic)
	problemIdentity := c.Query("problem_identity")
	userIdentity := c.Query("user_identity")     //从前端请求拿到参数
	status, _ := strconv.Atoi(c.Query("status")) //atoi string 转化成int

	tx := models.GetSubmitList(problemIdentity, userIdentity, status) //拿到的数据
	err = tx.Count(&count).Offset(page).Limit(size).Find(&list).Error
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "get submit error" + err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"List":  list,
			"count": count,
		},
	})

}

// Submit
// @Tags 用户私有方法
// @Summary 代码提交
// @Param authorization header string true "authorization"
// @Param problem_identity query string true "problem_identity"
// @Param code body string true "code"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /user/submit [post]
func Submit(c *gin.Context) {
	problemIdentity := c.Query("problem_identity")
	code, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "Read Code error" + err.Error(),
		})
		return
	}

	//代码保存
	path, err := helper.CodeSave(code)
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "Code save error" + err.Error(),
		})
		return
	}

	//auth_user里的user得到user_id
	u, _ := c.Get("user_claims")
	userClaim := u.(*helper.UserClaims)

	sb := &models.SubmitBasic{
		Identity:        helper.GetUUID(),
		ProblemIdentity: problemIdentity,
		UserIdentity:    userClaim.Identity,
		Path:            path,
	}

	//代码判断
	pb := new(models.ProblemBasic)
	err = models.DB.Where("identity = ?", problemIdentity).Preload("TestCases").First(pb).Error
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "Get Problem error" + err.Error(),
		})
		return
	}

	//定义channel
	//答案错误
	WA := make(chan int)
	//运行超内存
	OOM := make(chan int)
	//编译错误
	CE := make(chan int)
	//答案正确
	AC := make(chan int)
	//非法代码
	//EC :=make(chan struct{})

	passCount := 0
	//互斥锁
	var lock sync.Mutex
	var msg string

	for _, testCase := range pb.TestCases {
		testCase := testCase
		go func() {
			cmd := exec.Command("go", "run", path)
			var out, stderr bytes.Buffer

			cmd.Stderr = &stderr
			cmd.Stdout = &out
			stdinPipe, err := cmd.StdinPipe()
			//根据测试的输入案例输出结果 和标准输出的结果比对
			if err != nil {
				log.Fatalln(err)
			}
			io.WriteString(stdinPipe, testCase.Input+"\n")

			var bm runtime.MemStats
			runtime.ReadMemStats(&bm)

			if err := cmd.Run(); err != nil {
				log.Println(err, stderr.String())
				if err.Error() == "exit status 2" {
					msg = stderr.String()
					CE <- 1
					return
				}

			}
			var em runtime.MemStats
			runtime.ReadMemStats(&em)
			//答案错误
			if testCase.Output != out.String() {

				WA <- 1
				return
			}

			//运行超内存 默认kb
			if em.Alloc/1024-(bm.Alloc/1024) > uint64(pb.MaxMem) {

				OOM <- 1
				return

			}
			lock.Lock()
			passCount++
			if passCount == len(pb.TestCases) {
				AC <- 1
			}
			lock.Unlock()
		}()
	}

	select {
	case <-WA:
		msg = "答案错误"
		sb.Status = 2
	case <-OOM:
		msg = "运行超内存"
		sb.Status = 4
	case <-CE:
		sb.Status = 5
	case <-AC:
		msg = "答案正确"
		sb.Status = 1
	case <-time.After(time.Millisecond * time.Duration(pb.MaxRuntime)):
		if passCount == len(pb.TestCases) {
			sb.Status = 1
			msg = "答案正确"
		} else {
			sb.Status = 3
			msg = "运行超时"
		}

	}

	if err = models.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Create(sb).Error
		if err != nil {
			return errors.New("SubmitBasic Save Error:" + err.Error())
		}

		m := make(map[string]interface{})
		m["submit_num"] = gorm.Expr("submit_num + ?", 1)
		if sb.Status == 1 {
			m["pass_num"] = gorm.Expr("pass_num + ?", 1)
		}
		//更新user_basic
		err = tx.Model(new(models.UserBasic)).Where("identity = ?", userClaim.Identity).Updates(m).Error
		if err != nil {
			return errors.New("UserBasic Modify Error:" + err.Error())
		}

		//更新 problem_basic
		err = tx.Model(new(models.ProblemBasic)).Where("identity = ?", problemIdentity).Updates(m).Error
		if err != nil {
			return errors.New("problemBasic Modify Error:" + err.Error())
		}

		return nil
	}); err != nil {

		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "Submit error" + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"status": sb.Status,
			"msg":    msg,
		},
	})
}
