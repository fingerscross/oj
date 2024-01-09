package test

import (
	"fmt"
	"testing"

	"github.com/dgrijalva/jwt-go"
)

type UserClaims struct {
	Identity string `json:"identity"`
	Name     string `json:"name"`

	jwt.StandardClaims
}

var myKey = []byte("gin-gorm-oj-key")

//别人只拿到token,没拿到密钥,也不能进行解密

// 生成token
func TestGenerateToken(t *testing.T) {
	UserClaim := &UserClaims{
		Identity: "user_1",
		Name:     "Get",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, UserClaim)
	tokenString, err := token.SignedString(myKey) //tokenString就是真正的token
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("tokenString: %v\n", tokenString)

}

// 解析token
func TestAnalyseToken(t *testing.T) {
	tokenString := "hdfuehdf"
	userClaim := new(UserClaims) //将解析的token放进来
	claims, err := jwt.ParseWithClaims(tokenString, userClaim, func(t *jwt.Token) (interface{}, error) {
		return myKey, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if claims.Valid {
		fmt.Printf("userClaim: %v\n", userClaim)
	}

}
