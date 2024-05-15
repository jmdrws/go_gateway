package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var count1 = 0
var count2 = 0

func main() {
	rs1 := &RealServer{Addr: "127.0.0.1:8001"}
	rs1.Run()
	rs2 := &RealServer{Addr: "127.0.0.1:8002"}
	rs2.Run()
	//rs3 := &RealServer{Addr: "127.0.0.1:80"}
	//rs3.Run()
	//监听关闭信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

type RealServer struct {
	Addr string
}

func (r *RealServer) Run() {
	log.Println("Starting httpserver at " + r.Addr)
	mux := http.NewServeMux()
	mux.HandleFunc("/", r.HelloHandler)
	mux.HandleFunc("/base/error", r.ErrorHandler)
	mux.HandleFunc("/test_http_string/test_http_string/aaa", r.TimeoutHandler)
	server := &http.Server{
		Addr:         r.Addr,
		WriteTimeout: time.Second * 3,
		Handler:      mux,
	}
	go func() {
		log.Fatal(server.ListenAndServe())
	}()
}

func (r *RealServer) HelloHandler(w http.ResponseWriter, req *http.Request) {
	//127.0.0.1:8008/abc?sdsdsa=11
	//r.Addr=127.0.0.1:8008
	//req.URL.Path=/abc
	//fmt.Println(req.Host)
	upath := fmt.Sprintf("http://%s%s", r.Addr, req.URL.Path)
	//realIP := fmt.Sprintf("RemoteAddr=%s,X-Forwarded-For=%v,X-Real-Ip=%v\n", req.RemoteAddr, req.Header.Get("X-Forwarded-For"), req.Header.Get("X-Real-Ip"))
	//header := fmt.Sprintf("%v", req.Header)
	//io.WriteString(w, upath)
	//io.WriteString(w, realIP)
	//io.WriteString(w, header)

	//var node1 = ""
	//var node2 = ""
	//if upath == "http://127.0.0.1:8001/" {
	//	count1++
	//} else {
	//	count2++
	//}
	//node1 = node1 + "127.0.0.1:8001 请求总数: " + fmt.Sprintf("%d", count1)
	//node2 = node2 + "127.0.0.1:8002 请求总数: " + fmt.Sprintf("%d", count2)
	// 创建一个Response实例
	response := Response{
		Status: 200,
		//Message: upath + "\n" + realIP + "\n" + header,
		Path: upath,
		Msg:  "127.0.0.1:8080/http_url_write/first/check ---重写为---> 127.0.0.1:8001/http_backend/second/check",
		//Node1: node1,
		//Node2: node2,
		//Header: header,
	}

	// 设置响应头类型为JSON
	w.Header().Set("Content-Type", "application/json")

	// 将数据编码为JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 写入响应
	w.Write(jsonResponse)
}

// 定义一个结构体，用于生成JSON数据
type Response struct {
	Status int    `json:"status"`
	Path   string `json:"path"`
	//Header string `json:"header"`
	Msg string `json:"msg"`
	//Node1  string `json:"node1"`
	//Node2  string `json:"node2"`
}

func (r *RealServer) ErrorHandler(w http.ResponseWriter, req *http.Request) {
	upath := "error handler"
	w.WriteHeader(500)
	io.WriteString(w, upath)
}

func (r *RealServer) TimeoutHandler(w http.ResponseWriter, req *http.Request) {
	time.Sleep(6 * time.Second)
	upath := "timeout handler"
	w.WriteHeader(200)
	io.WriteString(w, upath)
}
