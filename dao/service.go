package dao

import (
	"errors"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/jmdrws/go_gateway/dto"
	"github.com/jmdrws/go_gateway/public"
	"net/http/httptest"
	"strings"
	"sync"
)

type ServiceDetail struct {
	Info          *ServiceInfo   `json:"info" description:"基本信息"`
	HTTPRule      *HttpRule      `json:"http_rule" description:"http_rule"`
	TCPRule       *TcpRule       `json:"tcp_rule" description:"tcp_rule"`
	GRPCRule      *GrpcRule      `json:"grpc_rule" description:"grpc_rule"`
	LoadBalance   *LoadBalance   `json:"load_balance" description:"load_balance"`
	AccessControl *AccessControl `json:"access_control" description:"access_control"`
}

// ServiceManagerHandler 将ServiceManager以Handler的形式暴露出去
var ServiceManagerHandler *ServiceManager

//在加载dao层的时候，就会直接执行init方法，即直接进行初始化
func init() {
	ServiceManagerHandler = NewServiceManager()
}

type ServiceManager struct {
	ServiceMap   map[string]*ServiceDetail //通过key去取
	ServiceSlice []*ServiceDetail          //通过遍历去取
	Locker       sync.RWMutex
	init         sync.Once //初始化设置，使其值初始化一次
	err          error
}

// NewServiceManager new初始化，设置为空的即可
func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		ServiceMap:   map[string]*ServiceDetail{},
		ServiceSlice: []*ServiceDetail{},
		Locker:       sync.RWMutex{},
		init:         sync.Once{},
	}
}
func (s *ServiceManager) GetTcpServiceList() []*ServiceDetail {
	var list []*ServiceDetail
	for _, serviceItem := range s.ServiceSlice {
		tempItem := serviceItem
		if tempItem.Info.LoadType == public.LoadTypeTCP {
			list = append(list, tempItem)
		}
	}
	return list
}

func (s *ServiceManager) GetGrpcServiceList() []*ServiceDetail {
	var list []*ServiceDetail
	for _, serviceItem := range s.ServiceSlice {
		tempItem := serviceItem
		if tempItem.Info.LoadType == public.LoadTypeGRPC {
			list = append(list, tempItem)
		}
	}
	return list
}

// HTTPAccessMode 接入匹配方法
func (s *ServiceManager) HTTPAccessMode(c *gin.Context) (*ServiceDetail, error) {
	//http有两种匹配方式
	//1、前缀匹配 /abc ==> serviceSlice.rule
	//2、域名匹配 www.test.com ==> serviceSlice.rule
	//域名：host c.Request.Host
	//		path c.Request.URL.Path

	host := c.Request.Host                  //eg: www.test.com:8080
	host = host[0:strings.Index(host, ":")] //截取前面的域名进行匹配
	//fmt.Println("host:", host)
	path := c.Request.URL.Path //eg: 请求/abc/get?id=111	path=/abc/get

	//遍历选择slice，因为它没有锁，相比起来更高效一些
	for _, serviceItem := range s.ServiceSlice {
		//判断是否为http的服务
		if serviceItem.Info.LoadType != public.LoadTypeHTTP {
			continue
		}
		//开始匹配
		//域名匹配方式
		if serviceItem.HTTPRule.RuleType == public.HTTPRuleTypeDomain {
			if serviceItem.HTTPRule.Rule == host {
				return serviceItem, nil
			}
		}
		//前缀匹配
		if serviceItem.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL {
			//func HasPrefix(s, prefix string) bool 判断字符串s是否以传入的prefix开头
			if strings.HasPrefix(path, serviceItem.HTTPRule.Rule) {
				return serviceItem, nil
			}
		}
	}
	return nil, errors.New("not matched service")
}

// LoadOnce 一次性加载所有服务信息到内存的方法（直接在main.go中触发）
func (s *ServiceManager) LoadOnce() error {
	//DO：只会执行一次，它在内部有一把锁的机制
	s.init.Do(func() {
		serviceInfo := &ServiceInfo{}

		//通过httptest.NewRecorder()模拟了http.ResponseWriter这个参数
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		tx, err := lib.GetGormPool("default")
		if err != nil {
			s.err = err
			return
		}
		params := &dto.ServiceListInput{
			PageNo:   1,
			PageSize: 99999,
		}
		list, _, err := serviceInfo.PageList(c, tx, params)
		if err != nil {
			s.err = err
			return
		}
		//涉及到map的设置，如果我们在设置的时候出现了读的情况，那么就会出现无法找到，内存溢出的情况
		s.Locker.Lock()
		defer s.Locker.Unlock()
		for _, listItem := range list {
			//使用tempItem的原因：list for循环的listItem是一个复用的变量，
			//当循环过后，所用指向listItem的变量都会是listItem最后遍历的值
			tempItem := listItem
			serviceDetail, err := tempItem.ServiceDetail(c, tx, &tempItem)
			if err != nil {
				s.err = err
				return
			}
			s.ServiceMap[tempItem.ServiceName] = serviceDetail
			s.ServiceSlice = append(s.ServiceSlice, serviceDetail)
		}
	})
	return s.err
}
