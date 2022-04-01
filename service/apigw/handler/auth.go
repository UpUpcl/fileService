package handler

import (
	"apigw/common"
	"apigw/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

func IsTokenVaild(token string) bool {
	if len(token) != 40 {
		return false
	}
	// 判断token的时效性，是否过期
	// 从数据库tbl_user_token查询username对应的token
	// 是否一致
	return true
}

func HTTPInterceptor() gin.HandlerFunc {
	return func (c *gin.Context) {

			username := c.Request.FormValue("username")
			token := c.Request.FormValue("token")

			if len(username) < 3 || !IsTokenVaild(token){
				c.Abort()
				resp := util.NewRespMsg(
					int(common.StatusTokenInvalid),
					"token无效",
					nil)
				c.JSON(http.StatusOK, resp.JsonToBytes())
				return
			}
			c.Next()
		}
}
