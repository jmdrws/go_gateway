package public

import (
	"golang.org/x/time/rate"
	"sync"
)

var FlowLimiterHandler *FlowLimiter

type FlowLimiter struct {
	FlowLimiterMap   map[string]*FlowLimiterItem
	FlowLimiterSlice []*FlowLimiterItem
	Locker           sync.RWMutex
}

type FlowLimiterItem struct {
	ServiceName string
	Limiter     *rate.Limiter
}

func NewFlowLimiter() *FlowLimiter {
	return &FlowLimiter{
		FlowLimiterMap:   map[string]*FlowLimiterItem{},
		FlowLimiterSlice: []*FlowLimiterItem{},
		Locker:           sync.RWMutex{},
	}
}
func init() {
	FlowLimiterHandler = NewFlowLimiter()
}

func (counter *FlowLimiter) GetLimiter(serviceName string, qps float64) (*rate.Limiter, error) {
	for _, item := range counter.FlowLimiterSlice {
		if item.ServiceName == serviceName {
			return item.Limiter, nil
		}
	}

	newLimiter := rate.NewLimiter(rate.Limit(qps), int(qps*3))
	item := &FlowLimiterItem{
		ServiceName: serviceName,
		Limiter:     newLimiter,
	}
	counter.FlowLimiterSlice = append(counter.FlowLimiterSlice, item)
	counter.Locker.Lock()
	defer counter.Locker.Unlock()
	counter.FlowLimiterMap[serviceName] = item
	return newLimiter, nil
}
