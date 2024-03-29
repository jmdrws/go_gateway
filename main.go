package main

import (
	"flag"
	"github.com/jmdrws/go_gateway/dao"
	"github.com/jmdrws/go_gateway/golang_common/lib"
	"github.com/jmdrws/go_gateway/grpc_proxy_router"
	"github.com/jmdrws/go_gateway/http_proxy_router"
	"github.com/jmdrws/go_gateway/router"
	"github.com/jmdrws/go_gateway/tcp_proxy_router"
	"os"
	"os/signal"
	"syscall"
)

//endpoint dashboard后台管理  server代理服务器
//config ./conf/prod/ 对应配置文件夹
var (
	endpoint = flag.String("endpoint", "", "input endpoint dashboard or server")
	config   = flag.String("my_config", "", "input config file like ./conf/dev/")
)

func main() {
	//解析参数
	flag.Parse()
	if *endpoint == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *config == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *endpoint == "dashboard" {
		lib.InitModule(*config)
		defer lib.Destroy()
		router.HttpServerRun()

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		router.HttpServerStop()
	} else {
		lib.InitModule(*config)
		defer lib.Destroy()
		dao.ServiceManagerHandler.LoadOnce()
		dao.AppManagerHandler.LoadOnce()
		go func() {
			http_proxy_router.HttpServerRun()
		}()
		go func() {
			http_proxy_router.HttpsServerRun()
		}()
		go func() {
			tcp_proxy_router.TcpServerRun()
		}()
		go func() {
			grpc_proxy_router.GrpcServerRun()
		}()
		//quit持续监听信号（syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM）
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		grpc_proxy_router.GrpcServerStop()
		tcp_proxy_router.TcpServerStop()
		http_proxy_router.HttpServerStop()
		http_proxy_router.HttpsServerStop()
	}
}
