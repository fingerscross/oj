package helper

import (
	"crypto/md5"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/smtp"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jordan-wright/email"
	uuid "github.com/satori/go.uuid"
)

type UserClaims struct {
	Identity string `json:"identity"`
	Name     string `json:"name"`
	IsAdmin  int    `json:"is_admin"`
	jwt.StandardClaims
}

// 生成md5
func GetMd5(s string) string {

	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// md5需要byte类型 再用16进制输出
var myKey = []byte("gin-gorm-oj-key")

//别人只拿到token,没拿到密钥,也不能进行解密

// 生成token//参数对应用户的ientity和name
func GenerateToken(identity, name string, isAdmin int) (string, error) {
	UserClaim := &UserClaims{
		Identity:       identity,
		Name:           name,
		IsAdmin:        isAdmin,
		StandardClaims: jwt.StandardClaims{},
		//还可以增加一个过期时间
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaim)
	tokenString, err := token.SignedString(myKey) //tokenString就是真正的token
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// 解析token
func AnalyseToken(tokenString string) (*UserClaims, error) {
	userClaim := new(UserClaims) //将解析的token放进来
	claims, err := jwt.ParseWithClaims(tokenString, userClaim, func(token *jwt.Token) (interface{}, error) {
		return myKey, nil
	})
	if err != nil {
		return nil, err
	}
	//claims验证为假
	if !claims.Valid {
		return nil, fmt.Errorf("analyse token error :%v", err)
	}
	return userClaim, nil
}

// SendCode
// 发送验证码
func SendCode(toUserEmail, code string) error {
	e := email.NewEmail()
	e.From = "GET <gu18915702696@163.com>"
	e.To = []string{toUserEmail}
	e.Subject = "验证码已发送，请查收"
	e.HTML = []byte("您的验证码：<b>" + code + "</b>")
	return e.SendWithTLS("smtp.163.com:465", smtp.PlainAuth(
		"", "gu18915702696@163.com", "NDHBNEVUQPDIWUPH", "smtp.163.com"),
		&tls.Config{InsecureSkipVerify: true, ServerName: "smtp.163.com"})
}

// GetUUID
// uuid 用于生成唯一的identity
func GetUUID() string {
	return uuid.NewV4().String()
}

// G生成随机验证码
func GetRand() string {
	rand.Seed(time.Now().UnixNano()) //种随机数种子
	s := ""
	for i := 0; i < 6; i++ {
		s += strconv.Itoa(rand.Intn(10))
	}
	return s

}

// CodeSave
// 保存代码
func CodeSave(code []byte) (string, error) {
	dirName := "code/" + GetUUID()
	path := dirName + "/main.go"
	err := os.Mkdir(dirName, 0777)
	if err != nil {
		return "", err
	}
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	f.Write(code)
	defer f.Close()
	return path, nil
}
