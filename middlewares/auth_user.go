package middlewares

import (
	"net/http"

	"getcharzp.cn/helper"
	"github.com/gin-gonic/gin"
)

// 验证用户
func AuthUserCheck() gin.HandlerFunc {
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
		if userClaim == nil {
			c.Abort()
			c.JSON(200, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "Unauthorized",
			})
			return
		}
		c.Set("user_claims", userClaim)
		c.Next()

	}
}
