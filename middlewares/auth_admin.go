package middlewares

import (
	"net/http"

	"getcharzp.cn/helper"
	"github.com/gin-gonic/gin"
)

//验证用户是不是管理员

func AuthAdminCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		userClaim, err := helper.AnalyseToken(auth)
		if err != nil {
			c.Abort()
			c.JSON(200, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "Unauthorized",
			})
			return
		}
		if userClaim.IsAdmin != 1 || userClaim == nil {
			c.Abort()
			c.JSON(200, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "Unauthorized",
			})
			return
		}
		c.Next()
	}
}
