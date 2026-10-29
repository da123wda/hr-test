package rdb

import (
	"fmt"
	"hr-dynamic-form-engine/src/cfg"
	"hr-dynamic-form-engine/src/log"
	"time"

	"github.com/gomodule/redigo/redis"
)

var defaultPool *redis.Pool

func Init() {
	rc := cfg.Cfg.Redis
	if rc.Host == "" {
		rc.Host = "127.0.0.1:6379"
	}

	maxIdle := rc.IdleConn
	if maxIdle <= 0 {
		maxIdle = 20
	}
	maxActive := rc.MaxConn
	if maxActive <= 0 {
		maxActive = 100
	}

	pool := &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: time.Minute,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			var opts []redis.DialOption
			opts = append(opts,
				redis.DialConnectTimeout(3*time.Second),
				redis.DialReadTimeout(3*time.Second),
				redis.DialWriteTimeout(3*time.Second),
			)
			if rc.Password != "" {
				opts = append(opts, redis.DialPassword(rc.Password))
			}
			if rc.Database > 0 {
				opts = append(opts, redis.DialDatabase(rc.Database))
			}
			return redis.Dial("tcp", rc.Host, opts...)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	conn := pool.Get()
	defer conn.Close()

	if _, err := conn.Do("PING"); err != nil {
		panic(fmt.Errorf("Redis 数据库连接失败: %w", err))
	}

	defaultPool = pool
	log.Zap.Infoln("redis start success")
}

func DB() redis.Conn {
	if defaultPool == nil {
		panic(fmt.Errorf("redis 连接池未初始化"))
	}
	return defaultPool.Get()
}
