package http_proxy_middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/jmdrws/go_gateway/dao"
	"github.com/jmdrws/go_gateway/middleware"
	"github.com/jmdrws/go_gateway/reverse_proxy"
	"github.com/pkg/errors"
)

// HTTPReverseProxyMiddleware 反向代理中间件
func HTTPReverseProxyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serverInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serverInterface.(*dao.ServiceDetail)

		//创建基于服务的负载均衡器设置，每一个服务中都有一个独立的负载均衡器
		//获取负载均衡的策略
		lb, err := dao.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
		if err != nil {
			middleware.ResponseError(c, 2002, err)
			c.Abort()
			return
		}
		//连接池设置，期望的是每个服务拥有独立的连接池（基本上都是设置一些超时时间之类的参数设置）
		trans, err := dao.TransporterHandler.GetTrans(serviceDetail)
		if err != nil {
			middleware.ResponseError(c, 2003, err)
			c.Abort()
			return
		}
		//middleware.ResponseSuccess(c,"ok")
		//return
		//创建 reverse-proxy
		//使用 reverse-proxy.ServerHTTP(c.Writer, c.Request)	执行实际的下游服务器的数据
		//有了策略和连接池后就可以创建反向代理了
		proxy := reverse_proxy.NewLoadBalanceReverseProxy(c, lb, trans)
		proxy.ServeHTTP(c.Writer, c.Request)
		c.Abort()
		return
	}
}
