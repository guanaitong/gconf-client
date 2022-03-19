package gconf_middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/guanaitong/gconf-go-client"
)

const defaultRedisConfigKey = "redis-config.json"

type RedisType int

const (
	RedisStandalone RedisType = iota
	RedisSentinel
)

// 单机模式配置
type RedisStandaloneConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	NodeHost string `json:"nodeHost"`
	NodePort int    `json:"nodePort"`
}

// 哨兵模式配置
type RedisSentinelConfig struct {
	Master string `json:"master"`
	Nodes  string `json:"nodes"`
}

type RedisConfig struct {
	Type              RedisType             `json:"type"`
	Standalone        RedisStandaloneConfig `json:"standalone"`
	Sentinel          RedisSentinelConfig   `json:"sentinel"`
	Password          string                `json:"password"`
	EncryptedPassword string                `json:"encryptedPassword"`
	Db                int                   `json:"db"`
}

func (redisConfig *RedisConfig) NewClient() *redis.Client {
	var pwd = decrypt(redisConfig.EncryptedPassword)
	if pwd == "" {
		pwd = redisConfig.Password
	}

	if redisConfig.Type == RedisStandalone {
		var (
			host, port = redisConfig.Standalone.Host, redisConfig.Standalone.Port
		)
		opt := &redis.Options{
			Addr:     fmt.Sprintf("%s:%d", host, port),
			Password: pwd,
			DB:       redisConfig.Db,
		}
		return redis.NewClient(opt)
	} else if redisConfig.Type == RedisSentinel {
		fOpt := &redis.FailoverOptions{
			MasterName:    redisConfig.Sentinel.Master,
			SentinelAddrs: strings.Split(redisConfig.Sentinel.Nodes, ","),
			Password:      pwd,
			DB:            redisConfig.Db,
		}
		return redis.NewFailoverClient(fOpt)
	} else {
		panic("unsupported type")
	}
}

func GetRedisConfig(key string) *RedisConfig {
	if key == "" {
		panic(errors.New("redis config is null"))
	}

	redisConfig := new(RedisConfig)
	configValue := gconf.GetCurrentConfigCollection().GetValue(key).Raw()
	err := json.Unmarshal([]byte(configValue), redisConfig)
	if err != nil {
		panic(err.Error())
	}
	return redisConfig
}

func GetDefaultRedisConfig() *RedisConfig {
	return GetRedisConfig(defaultRedisConfigKey)
}
