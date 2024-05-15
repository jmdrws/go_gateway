package public

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/jmdrws/go_gateway/golang_common/lib"
	"sync/atomic"
	"time"
)

// RedisFlowCountService 流量统计器结构体
type RedisFlowCountService struct {
	AppID       string
	Interval    time.Duration
	QPS         int64
	Unix        int64
	TickerCount int64
	TotalCount  int64
}

// NewRedisFlowCountService 参数：设置APPID和统计结果的刷新时间频率
func NewRedisFlowCountService(appID string, interval time.Duration) *RedisFlowCountService {
	reqCounter := &RedisFlowCountService{
		AppID:    appID,
		Interval: interval,
		QPS:      0,
		Unix:     0,
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()
		ticker := time.NewTicker(interval)
		for {
			<-ticker.C
			//获取请求次数reqCounter.TickerCount
			tickerCount := atomic.LoadInt64(&reqCounter.TickerCount)
			//重置请求次数reqCounter.TickerCount
			atomic.StoreInt64(&reqCounter.TickerCount, 0)

			currentTime := time.Now()
			dayKey := reqCounter.GetDayKey(currentTime)
			hourKey := reqCounter.GetHourKey(currentTime)
			if err := RedisConfPipline(
				func(c redis.Conn) {
					//数据的增加
					c.Send("INCRBY", dayKey, tickerCount)
					//超时时间设置
					c.Send("EXPIRE", dayKey, 86400*2)

					c.Send("INCRBY", hourKey, tickerCount)
					c.Send("EXPIRE", hourKey, 86400*2)
				}); err != nil {
				fmt.Println("RedisConfPipline err", err)
				continue
			}

			totalCount, err := reqCounter.GetDayData(currentTime)
			if err != nil {
				fmt.Println("reqCounter.GetDayData err", err)
				continue
			}
			nowUnix := time.Now().Unix()
			if reqCounter.Unix == 0 {
				reqCounter.Unix = time.Now().Unix()
				continue
			}
			tickerCount = totalCount - reqCounter.TotalCount
			if nowUnix > reqCounter.Unix {
				reqCounter.TotalCount = totalCount
				//QPS
				reqCounter.QPS = tickerCount / (nowUnix - reqCounter.Unix)
				reqCounter.Unix = time.Now().Unix()
				//reqCounter.Unix = nowUnix
			}
		}
	}()
	return reqCounter
}

// GetDayKey 统计力度（天），组装到RedisKey
func (o *RedisFlowCountService) GetDayKey(t time.Time) string {
	dayStr := t.In(lib.TimeLocation).Format("20060102")
	//设置到redis的key中
	return fmt.Sprintf("%s_%s_%s", RedisFlowDayKey, dayStr, o.AppID)
}

// GetHourKey 统计力度（小时），组装到RedisKey
func (o *RedisFlowCountService) GetHourKey(t time.Time) string {
	hourStr := t.In(lib.TimeLocation).Format("2006010215")
	return fmt.Sprintf("%s_%s_%s", RedisFlowHourKey, hourStr, o.AppID)
}

// GetHourData 封装获取方法 redis的get获取数据
func (o *RedisFlowCountService) GetHourData(t time.Time) (int64, error) {
	return redis.Int64(RedisConfDo("GET", o.GetHourKey(t)))
}

func (o *RedisFlowCountService) GetDayData(t time.Time) (int64, error) {
	return redis.Int64(RedisConfDo("GET", o.GetDayKey(t)))
}

// Increase 原子增加
func (o *RedisFlowCountService) Increase() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()
		atomic.AddInt64(&o.TickerCount, 1)
		//fmt.Println("TickerCount---------->", o.TickerCount)
	}()
}
