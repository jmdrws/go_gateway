package http_proxy_middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/jmdrws/go_gateway/dao"
	"github.com/jmdrws/go_gateway/middleware"
	"github.com/jmdrws/go_gateway/public"
	"github.com/pkg/errors"
)

func HTTPFlowCountMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serverInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}
		//从上下文中获取后转换
		serviceDetail := serverInterface.(*dao.ServiceDetail)

		//统计项 1 全站 2 服务
		//1、全站
		totalCounter, err := public.FlowCounterHandler.GetCounter(public.FlowTotal)
		if err != nil {
			middleware.ResponseError(c, 4001, err)
			c.Abort()
			return
		}
		totalCounter.Increase()
		//_, _ = totalCounter.GetDayData(time.Now())
		//fmt.Printf("totalCounter qps:%v,dayCount:%v\n", totalCounter.QPS, dayCount)
		//2、服务
		serviceCounter, err := public.FlowCounterHandler.GetCounter(public.FlowServicePrefix + serviceDetail.Info.ServiceName)
		if err != nil {
			middleware.ResponseError(c, 4001, err)
			c.Abort()
			return
		}
		serviceCounter.Increase()
		//_, _ = serviceCounter.GetDayData(time.Now())
		//fmt.Printf("serviceCounter qps:%v,dayCount:%v", serviceCounter.QPS, dayServiceCount)
		c.Next()
	}
}
