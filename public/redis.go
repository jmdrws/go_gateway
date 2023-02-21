package public

import (
	"github.com/e421083458/golang_common/lib"
	"github.com/gomodule/redigo/redis"
)

// RedisConfPipline redis连接的设置方法，例如在流量统计中间件中设置数据和超时时间
func RedisConfPipline(pip ...func(c redis.Conn)) error {
	//redis读取方式，创建连接
	c, err := lib.RedisConnFactory("default")
	if err != nil {
		return err
	}
	defer c.Close()
	for _, f := range pip {
		f(c)
	}
	c.Flush()
	return nil
}

// RedisConfDo redis执行操作的方法，例如在流量统计中间件中使用get方法获取redis中储存的流量数据
func RedisConfDo(commandName string, args ...interface{}) (interface{}, error) {
	c, err := lib.RedisConnFactory("default")
	if err != nil {
		return nil, err
	}
	defer c.Close()
	return c.Do(commandName, args...)
}
