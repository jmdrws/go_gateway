package load_balance

import (
	"fmt"
	"net"
	"reflect"
	"sort"
	"time"
)

const (
	//default check setting
	DefaultCheckMethod    = 0
	DefaultCheckTimeout   = 5
	DefaultCheckMaxErrNum = 2
	DefaultCheckInterval  = 5
)

type LoadBalanceCheckConf struct {
	observers    []Observer
	confIpWeight map[string]string
	activeList   []string
	format       string
}

func (s *LoadBalanceCheckConf) Attach(o Observer) {
	s.observers = append(s.observers, o)
}

func (s *LoadBalanceCheckConf) NotifyAllObservers() {
	for _, obs := range s.observers {
		obs.Update()
	}
}

func (s *LoadBalanceCheckConf) GetConf() []string {
	confList := []string{}
	for _, ip := range s.activeList {
		weight, ok := s.confIpWeight[ip]
		if !ok {
			weight = "50" //默认weight
		}
		confList = append(confList, fmt.Sprintf(s.format, ip)+","+weight)
	}
	return confList
}

func (s *LoadBalanceCheckConf) WatchConf() {
	//fmt.Println("watchConf")
	go func() {
		confIpErrNum := map[string]int{}
		for {
			changedList := []string{}
			for item, _ := range s.confIpWeight {
				//创建连接，设置超时时间
				conn, err := net.DialTimeout("tcp", item, time.Duration(DefaultCheckTimeout)*time.Second)
				if err == nil {
					//关闭连接，map中对应节点的err次数为空
					conn.Close()
					if _, ok := confIpErrNum[item]; ok {
						confIpErrNum[item] = 0
					}
				}
				if err != nil {
					//map中对应节点的err次数加一
					if _, ok := confIpErrNum[item]; ok {
						confIpErrNum[item] += 1
					} else {
						confIpErrNum[item] = 1
					}
				}
				//若err次数小于设置的最大失败数，就添加节点
				//否则表示节点挂了，不添加进节点切片中
				if confIpErrNum[item] < DefaultCheckMaxErrNum {
					changedList = append(changedList, item)
				}
			}
			sort.Strings(changedList)
			sort.Strings(s.activeList)
			if !reflect.DeepEqual(changedList, s.activeList) {
				s.UpdateConf(changedList)
			}
			time.Sleep(time.Duration(DefaultCheckInterval) * time.Second)
		}
	}()
}

func (s *LoadBalanceCheckConf) UpdateConf(conf []string) {
	//fmt.Println("UpdateConf", conf)
	s.activeList = conf
	for _, obs := range s.observers {
		obs.Update()
	}
}

// NewLoadBalanceCheckConf check逻辑：循环检测超过一定次数就会移除这个节点
func NewLoadBalanceCheckConf(format string, conf map[string]string) (*LoadBalanceCheckConf, error) {
	var aList []string
	//默认初始化
	for item, _ := range conf {
		aList = append(aList, item)
	}
	mConf := &LoadBalanceCheckConf{format: format, activeList: aList, confIpWeight: conf}
	//探活检测
	mConf.WatchConf()
	return mConf, nil
}
