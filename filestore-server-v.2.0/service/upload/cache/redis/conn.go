package rPool

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"time"
)

var (
	pool *redis.Pool
	redisHost = "127.0.0.1:6379"
)

func newRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle: 1000,
		MaxActive: 1000,
		IdleTimeout: 300*time.Second,
		Dial: func() (redis.Conn, error) {
			// 1.打开链接
			c, err := redis.Dial("tcp", redisHost)
			if err != nil {
				fmt.Println(err.Error())
				return nil, err
			}
			// 2.访问凭证
			return  c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute{
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

func init()  {
	pool = newRedisPool()
}

func RedisPool() *redis.Pool {
	return pool
}
