package http_proxy_middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/jmdrws/go_gateway/dao"
	"github.com/jmdrws/go_gateway/middleware"
)

// HTTPAccessModeMiddleware 匹配服务的接入方式  基于请求信息
func HTTPAccessModeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		service, err := dao.ServiceManagerHandler.HTTPAccessMode(c)
		if err != nil {
			middleware.ResponseError(c, 1001, err)
			c.Abort()
			return
		}
		//fmt.Println("matched service:", public.Obj2Json(service))
		//设置上下文信息，方便之后的服务取得这个服务的信息
		c.Set("service", service)
		c.Next()
	}
}
