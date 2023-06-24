package redis

import (
	"time"

	redisLib "github.com/go-redis/redis"
	"github.com/rohanchauhan02/common/logs"
	redisTraceLib "gopkg.in/DataDog/dd-trace-go.v1/contrib/go-redis/redis"
)

var (
	logger = logs.NewCommonLog()
)

type Redis interface {
	InitClient() error
	SetRedisValue(key string, payload string, ttl time.Duration)
	GetRedisValue(key string) string
	DeleteRedisValue(key string) int64
	GetClient() *redisTraceLib.Client
}

type RedisConfig struct {
	Host        string
	Password    string
	DB          int
	PoolSize    int
	ReadTimeout time.Duration
}

type redis struct {
	host        string
	password    string
	db          int
	poolSize    int
	readTimeout time.Duration
	client      *redisTraceLib.Client
}

// NewRedis is a factory that return interface of its implementation
func NewRedis(config RedisConfig) Redis {
	return &redis{
		host:        config.Host,
		password:    config.Password,
		db:          config.DB,
		poolSize:    config.PoolSize,
		readTimeout: config.ReadTimeout,
	}
}

func (r *redis) InitClient() error {

	logger.Info("Start open redis connection...")

	redisOpt := &redisLib.Options{
		Addr:     r.host,
		Password: r.password,
	}

	if r.db != 0 {
		redisOpt.DB = r.db
	} else {
		redisOpt.DB = 0
	}

	if r.poolSize != 0 {
		redisOpt.PoolSize = r.poolSize
	} else {
		redisOpt.PoolSize = 64
	}

	if r.readTimeout != 0 {
		redisOpt.ReadTimeout = r.readTimeout
	} else {
		redisOpt.ReadTimeout = 10 * time.Second
	}

	cl := redisTraceLib.NewClient(redisOpt)

	_, err := cl.Ping().Result()
	if err != nil {
		return err
	}
	r.client = cl

	return nil
}

func (r *redis) SetRedisValue(key string, payload string, ttl time.Duration) {
	r.client.Set(key, payload, ttl)
}

func (r *redis) GetRedisValue(key string) string {
	val, err := r.client.Get(key).Result()
	if err != nil {
		return ""
	}
	return val
}

func (r *redis) DeleteRedisValue(key string) int64 {
	val, err := r.client.Del(key).Result()
	if err != nil {
		return 0
	}
	return val
}

func (r *redis) GetClient() *redisTraceLib.Client {
	return r.client
}
