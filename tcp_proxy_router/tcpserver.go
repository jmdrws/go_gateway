package tcp_proxy_router

import (
	"context"
	"fmt"
	"github.com/jmdrws/go_gateway/dao"
	"github.com/jmdrws/go_gateway/reverse_proxy"
	"github.com/jmdrws/go_gateway/tcp_proxy_middleware"
	"github.com/jmdrws/go_gateway/tcp_server"
	"log"
	"net"
)

var tcpServerList []*tcp_server.TcpServer

type tcpHandler struct {
}

func (t *tcpHandler) ServeTCP(ctx context.Context, conn net.Conn) {
	conn.Write([]byte("tcpHandler\n"))
}

func TcpServerRun() {
	// 获取TCP服务列表，将tcp的端口全部打开
	serviceList := dao.ServiceManagerHandler.GetTcpServiceList()
	for _, serviceItem := range serviceList {
		tempItem := serviceItem
		//通过将tempItem传入协程 开启所有的tcp服务
		go func(serviceDetail *dao.ServiceDetail) {
			//设置tcp服务器
			//获取端口
			addr := fmt.Sprintf(":%d", serviceDetail.TCPRule.Port)
			//获取负载均衡
			rb, err := dao.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
			if err != nil {
				log.Fatalf(" [INFO] GetTcpLoadBalancer %v err:%v\n", addr, err)
				return
			}
			//构建路由及设置中间件
			router := tcp_proxy_middleware.NewTcpSliceRouter()
			router.Group("/").Use(
				tcp_proxy_middleware.TCPFlowCountMiddleware(),
				tcp_proxy_middleware.TCPFlowLimitMiddleware(),
				tcp_proxy_middleware.TCPWhiteListMiddleware(),
				tcp_proxy_middleware.TCPBlackListMiddleware(),
			)

			//构建回调handler
			routerHandler := tcp_proxy_middleware.NewTcpSliceRouterHandler(
				//传入负载均衡策略，路由，设置反向代理
				func(c *tcp_proxy_middleware.TcpSliceRouterContext) tcp_server.TCPHandler {
					return reverse_proxy.NewTcpLoadBalanceReverseProxy(c, rb)
				}, router)
			baseCtx := context.WithValue(context.Background(), "service", serviceDetail)

			// 配置TCPServer
			tcpServer := &tcp_server.TcpServer{
				Addr:    addr,
				Handler: routerHandler,
				BaseCtx: baseCtx,
			}

			//放入切片中
			tcpServerList = append(tcpServerList, tcpServer)
			log.Printf(" [INFO] tcp_proxy_run %v\n", addr)

			//开启监听
			if err := tcpServer.ListenAndServe(); err != nil && err != tcp_server.ErrServerClosed {
				log.Fatalf(" [INFO] tcp_proxy_run %v err:%v\n", addr, err)
			}
		}(tempItem)
	}
}

func TcpServerStop() {
	for _, tcpServer := range tcpServerList {
		tcpServer.Close()
		log.Printf(" [INFO] tcp_proxy_stop %v stopped\n", tcpServer.Addr)
	}
}
