package reverse_proxy

import (
	"github.com/gin-gonic/gin"
	"github.com/jmdrws/go_gateway/middleware"
	"github.com/jmdrws/go_gateway/reverse_proxy/load_balance"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

//var count1 = 0
//var count2 = 0

// NewLoadBalanceReverseProxy 创建反向代理的方法
func NewLoadBalanceReverseProxy(c *gin.Context, lb load_balance.LoadBalance, trans *http.Transport) *httputil.ReverseProxy {
	//修改控制器内容，说白了就是拼接
	director := func(req *http.Request) {

		//取得负载均衡的值nextAddr
		nextAddr, err := lb.Get(req.URL.String())
		//if nextAddr ==  "http://127.0.0.1:8001"{
		//	count1++
		//	fmt.Println("负载均衡分配IP地址: ", nextAddr)
		//	fmt.Print(nextAddr,"请求总数: ", count1)
		//}else {
		//	count2++
		//	fmt.Println("负载均衡分配IP地址: ", nextAddr)
		//	fmt.Print(nextAddr,"请求总数: ", count2)
		//}
		//fmt.Print("nextAddr ", nextAddr)
		if err != nil || nextAddr == "" {
			panic("get next addr fail")
		}
		target, err := url.Parse(nextAddr)
		if err != nil {
			panic(err)
		}
		//http://127.0.0.1:2002/dir?name=123
		//targetQuery: name=123
		//Scheme: http
		//Host: 127.0.0.1:2002
		targetQuery := target.RawQuery
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		//singleJoiningSlash("/base","/dir")
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		req.Host = target.Host
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "user-agent")
		}
	}

	//更改内容
	modifyFunc := func(resp *http.Response) error {
		if strings.Contains(resp.Header.Get("Connection"), "Upgrade") {
			return nil
		}
		//
		//var payload []byte
		//var readErr error
		//
		//if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		//	gr, err := gzip.NewReader(resp.Body)
		//	if err != nil {
		//		return err
		//	}
		//	payload, readErr = ioutil.ReadAll(gr)
		//	resp.Header.Del("Content-Encoding")
		//} else {
		//	payload, readErr = ioutil.ReadAll(resp.Body)
		//}
		//if readErr != nil {
		//	return readErr
		//}
		//
		//c.Set("status_code", resp.StatusCode)
		//c.Set("payload", payload)
		//resp.Body = ioutil.NopCloser(bytes.NewBuffer(payload))
		//resp.ContentLength = int64(len(payload))
		//resp.Header.Set("Content-Length", strconv.FormatInt(int64(len(payload)), 10))
		return nil
	}

	//错误回调 ：关闭real_server时测试，错误回调
	//范围：transport.RoundTrip发生的错误、以及ModifyResponse发生的错误
	errFunc := func(w http.ResponseWriter, r *http.Request, err error) {
		middleware.ResponseError(c, 999, err)
	}
	return &httputil.ReverseProxy{Director: director, Transport: trans, ModifyResponse: modifyFunc, ErrorHandler: errFunc}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
