package http_proxy_middleware

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmdrws/go_gateway/dao"
	"github.com/jmdrws/go_gateway/middleware"
	"github.com/jmdrws/go_gateway/public"
	"strings"
)

// HTTPJwtAuthTokenMiddleware jwt auth token
func HTTPJwtAuthTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)

		// decode jwt token
		// app_list 取得 appInfo   app_id 与 claims.Issuer 对应
		// appInfo 放到 gin.context
		token := strings.ReplaceAll(c.GetHeader("Authorization"), "Bearer ", "")
		fmt.Println(token)
		appMatched := false
		if token != "" {
			claims, err := public.JwtDecode(token)
			if err != nil {
				middleware.ResponseError(c, 2002, err)
				c.Abort()
				return
			}
			//fmt.Println("claims.Issuer",claims.Issuer)
			appList := dao.AppManagerHandler.GetAppList()
			for _, appInfo := range appList {
				if appInfo.AppID == claims.Issuer {
					c.Set("app", appInfo)
					appMatched = true
					break
				}
			}
		}
		if serviceDetail.AccessControl.OpenAuth == 1 && !appMatched {
			middleware.ResponseError(c, 2003, errors.New("not match valid app"))
			c.Abort()
			return
		}
		c.Next()
	}
}
