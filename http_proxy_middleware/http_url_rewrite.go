package http_proxy_middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/jmdrws/go_gateway/dao"
	"github.com/jmdrws/go_gateway/middleware"
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

//重写url
func HTTPUrlRewriteMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serverInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serverInterface.(*dao.ServiceDetail)
		//url重写功能，每行一个
		//items[0] 重写为 items[1]
		//正则表达式 ： ^/test_http_service/abbb/(.*) /test_http_service/bbba/$1
		//将/test_http_service/abbb 重写为/test_http_service/baaa
		for _, item := range strings.Split(serviceDetail.HTTPRule.UrlRewrite, ",") {
			//fmt.Println("item rewrite", item)
			items := strings.Split(item, " ")
			if len(items) != 2 {
				continue
			}
			regexp, err := regexp.Compile(items[0])
			if err != nil {
				continue
			}
			//fmt.Println("before rewrite", c.Request.URL.Path)
			replacePath := regexp.ReplaceAll([]byte(c.Request.URL.Path), []byte(items[1]))
			c.Request.URL.Path = string(replacePath)
			//fmt.Println("after rewrite", c.Request.URL.Path)
		}
		c.Next()
	}
}
