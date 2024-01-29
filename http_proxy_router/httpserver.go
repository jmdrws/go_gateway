package http_proxy_router

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jmdrws/go_gateway/cert_file"
	"github.com/jmdrws/go_gateway/golang_common/lib"
	"github.com/jmdrws/go_gateway/middleware"
	"log"
	"net/http"
	"time"
)

var (
	HttpSrvHandler  *http.Server
	HttpsSrvHandler *http.Server
)

func HttpServerRun() {
	gin.SetMode(lib.GetStringConf("proxy.base.debug_mode"))
	r := InitRouter(
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
	)
	HttpSrvHandler = &http.Server{
		Addr:           lib.GetStringConf("proxy.http.addr"),
		Handler:        r,
		ReadTimeout:    time.Duration(lib.GetIntConf("proxy.http.read_timeout")) * time.Second,
		WriteTimeout:   time.Duration(lib.GetIntConf("proxy.http.write_timeout")) * time.Second,
		MaxHeaderBytes: 1 << uint(lib.GetIntConf("proxy.http.max_header_bytes")),
	}
	log.Printf(" [INFO] HttpProxyRun:%s\n", lib.GetStringConf("proxy.http.addr"))
	if err := HttpSrvHandler.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf(" [ERROR] HttpProxyRun %s err:%v\n", lib.GetStringConf("proxy.http.addr"), err)
	}
}

func HttpsServerRun() {
	gin.SetMode(lib.GetStringConf("proxy.base.debug_mode"))
	r := InitRouter(
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
	)
	HttpsSrvHandler = &http.Server{
		Addr:           lib.GetStringConf("proxy.https.addr"),
		Handler:        r,
		ReadTimeout:    time.Duration(lib.GetIntConf("proxy.https.read_timeout")) * time.Second,
		WriteTimeout:   time.Duration(lib.GetIntConf("proxy.https.write_timeout")) * time.Second,
		MaxHeaderBytes: 1 << uint(lib.GetIntConf("proxy.https.max_header_bytes")),
	}
	log.Printf(" [INFO] HttpsProxyRun:%s\n", lib.GetStringConf("proxy.https.addr"))
	if err := HttpsSrvHandler.ListenAndServeTLS(cert_file.Path("server.crt"), cert_file.Path("server.key")); err != nil && err != http.ErrServerClosed {
		log.Fatalf(" [ERROR] HttpsProxyRun:%s err:%v\n", lib.GetStringConf("proxy.https.addr"), err)
	}
}

func HttpServerStop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := HttpSrvHandler.Shutdown(ctx); err != nil {
		log.Fatalf(" [ERROR] HttpProxyRun err %v\n", err)
	}
	log.Printf(" [INFO] HttpProxyRun stopped\n")
}

func HttpsServerStop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := HttpsSrvHandler.Shutdown(ctx); err != nil {
		log.Fatalf(" [ERROR] HttpsProxyRun err:%v\n", err)
	}
	log.Printf(" [INFO] HttpsProxyRun stopped\n")
}
