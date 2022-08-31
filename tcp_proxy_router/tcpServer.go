package tcp_proxy_router

import (
	"github.com/e421083458/golang_common/lib"
	"github.com/jmdrws/go_gateway/dao"
	"log"
)

func TcpServerRun() {
	serviceList := dao.ServiceManagerHandler.GetTcpServiceList()
	for _, serviceItem := range serviceList {
		tempItem := serviceItem
		log.Printf("[INFO] tcp_proxy_run %s\n", lib.GetStringConf("proxy.tcp.addr"))
		go func(serviceInfo *dao.ServiceDetail) {

		}(tempItem)
	}

}

func TcpServerStop() {
	log.Printf("[INFO] tcp_proxy_stop stopped\n")
}
