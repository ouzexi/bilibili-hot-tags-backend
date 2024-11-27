package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// OriginMiddleware 创建一个中间件，用于检查请求来源
func OriginMiddleware(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求的接口路径
		/* requestPath := c.Request.URL.Path
		if requestPath == variable.IgnoreApi {
			c.Next()
			return
		} */

		origin := c.Request.Header.Get("Origin")
		referer := c.Request.Header.Get("Referer")

		// 检查请求的 Origin 或 Referer 是否符合要求
		if origin != allowedOrigin && referer != allowedOrigin {
			c.AbortWithStatus(http.StatusBadGateway) // 返回 502 状态码
			return
		}

		c.Next() // 继续处理请求
	}
}
