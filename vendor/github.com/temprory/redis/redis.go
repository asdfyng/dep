package redis

import (
	"crypto/tls"
	"errors"
	"fmt"
	redis "github.com/go-redis/redis"
	"sort"
	"strings"
	"sync"
	"time"
)

type RedisConf struct {
	ID                 string `json:"ID"`
	Addr               string `json:"Addr"`
	Password           string `json:"Password"`
	Database           int    `json:"Database"`
	PoolSize           int    `json:"PoolSize"`
	DialTimeout        int    `json:"DialTimeout"`
	ReadTimeout        int    `json:"ReadTimeout"`
	WriteTimeout       int    `json:"WriteTimeout"`
	PoolTimeout        int    `json:"PoolTimeout"`
	IdleTimeout        int    `json:"IdleTimeout"`
	IdleCheckFrequency int    `json:"IdleCheckFrequency"`
	KeepaliveInterval  int    `json:"KeepaliveInterval"`
	MaxRetries         int    `json:"MaxRetries"`
	// ReadOnly           bool        `json:"ReadOnly"`
	TLSConfig          *tls.Config `json:"TLSConfig"`
	Network            string      `json:"Network"`
	dialTimeout        time.Duration
	readTimeout        time.Duration
	writeTimeout       time.Duration
	keepaliveInterval  time.Duration
	poolTimeout        time.Duration
	idleTimeout        time.Duration
	idleCheckFrequency time.Duration
}

type RedisConfArr []RedisConf

func (arr RedisConfArr) Len() int {
	return len(arr)
}

func (arr RedisConfArr) Swap(i, j int) {
	arr[i], arr[j] = arr[j], arr[i]
}

func (arr RedisConfArr) Less(i, j int) bool {
	return arr[i].ID > arr[j].ID
}

type redisScript struct {
	sync.Mutex
	tag  string
	src  string
	sha1 string
	err  error
}

type Redis struct {
	sync.RWMutex
	client  *redis.Client
	ticker  *time.Ticker
	scripts map[string]*redisScript
}

func (rds *Redis) Client() *redis.Client {
	return rds.client
}

func (rds *Redis) getScriptSha1(tag string) (string, error) {
	rds.RLock()
	defer rds.RUnlock()
	if script, ok := rds.scripts[tag]; ok {
		return script.sha1, script.err
	}
	return "", errors.New(fmt.Sprintf("invalid redis lua script tag: %s", tag))
}

// 应用层应在初始化阶段完成所有LoadScript操作, 初始化后不应再调用此方法
func (rds *Redis) LoadScript(tag string, src string) error {
	rds.Lock()
	defer rds.Unlock()
	client := rds.Client()
	cmd := client.ScriptLoad(src)
	sha1, err := cmd.Result()
	rds.scripts[tag] = &redisScript{
		tag:  tag,
		src:  src,
		sha1: sha1,
		err:  err,
	}
	if err != nil {
		logError("redis ScriptLoad '%s' failed: %v", tag, err)
	}
	return err
}

func (rds *Redis) reLoadScript(tag string) error {
	rds.Lock()
	defer rds.Unlock()

	var (
		err  error = errors.New("no such script")
		sha1       = ""
	)
	if script, ok := rds.scripts[tag]; ok {
		cmd := rds.Client().ScriptLoad(script.src)
		sha1, err = cmd.Result()
		script.sha1, script.err = sha1, err
	}

	if err != nil {
		logError("redis ReLoadScript '%s' failed: %v", tag, err)
	}

	return err
}

func (rds *Redis) EvalSha(tag string, keys []string, args ...interface{}) (interface{}, error) {
	var ret interface{} = nil
	sha1, err := rds.getScriptSha1(tag)
	if err == nil {
		cmd := rds.Client().EvalSha(sha1, keys, args...)
		ret, err = cmd.Result()
		if err != nil && strings.HasPrefix(err.Error(), "NOSCRIPT") {
			if err = rds.reLoadScript(tag); err == nil {
				if sha1, err = rds.getScriptSha1(tag); err == nil {
					cmd = rds.Client().EvalSha(sha1, keys, args...)
					ret, err = cmd.Result()
				}
			}
		}
	}

	return ret, err
}

func (rds *Redis) Eval(src string, keys []string, args ...interface{}) (interface{}, error) {
	cmd := rds.Client().Eval(src, keys, args...)
	ret, err := cmd.Result()
	return ret, err
}

func (rds *Redis) Close() error {
	return rds.client.Close()
}

func NewRedis(conf RedisConf) *Redis {
	if conf.DialTimeout > 0 {
		conf.dialTimeout = time.Second * time.Duration(conf.DialTimeout)
	} else {
		conf.dialTimeout = time.Second * 3
	}
	if conf.ReadTimeout > 0 {
		conf.readTimeout = time.Second * time.Duration(conf.ReadTimeout)
	} else {
		conf.readTimeout = time.Second * 2
	}
	if conf.WriteTimeout > 0 {
		conf.writeTimeout = time.Second * time.Duration(conf.WriteTimeout)
	} else {
		conf.writeTimeout = time.Second * 2
	}
	if conf.KeepaliveInterval > 0 {
		conf.keepaliveInterval = time.Second * time.Duration(conf.KeepaliveInterval)
	} else {
		conf.keepaliveInterval = time.Second * 300
	}
	if conf.IdleTimeout > 0 {
		conf.idleTimeout = time.Second * time.Duration(conf.IdleTimeout)
	} else {
		// conf.idleTimeout  = time.Second * 2
	}
	if conf.IdleCheckFrequency > 0 {
		conf.idleCheckFrequency = time.Second * time.Duration(conf.IdleCheckFrequency)
	} else {
		// conf.idleCheckFrequency = time.Second * 300
	}
	if conf.KeepaliveInterval > 0 {
		conf.keepaliveInterval = time.Second * time.Duration(conf.KeepaliveInterval)
	} else {
		conf.keepaliveInterval = time.Second * 300
	}

	logInfo("NewRedis Connect To Redis ...")

	client := redis.NewClient(&redis.Options{
		Network:            conf.Network,
		Addr:               conf.Addr,
		Password:           conf.Password,
		DB:                 conf.Database,
		DialTimeout:        conf.dialTimeout,
		ReadTimeout:        conf.readTimeout,
		WriteTimeout:       conf.writeTimeout,
		PoolSize:           conf.PoolSize,
		PoolTimeout:        conf.poolTimeout,
		IdleTimeout:        conf.idleTimeout,
		IdleCheckFrequency: conf.idleCheckFrequency,
		MaxRetries:         conf.MaxRetries,
		// ReadOnly:           conf.ReadOnly,
		TLSConfig: conf.TLSConfig,
	})
	cmd := client.Ping()
	if cmd.Err() != nil {
		logFatal("NewRedis client.Ping() Failed: %v", cmd.Err())
	}

	ticker := time.NewTicker(conf.keepaliveInterval)
	safeGo(func() {
		for {
			if _, ok := <-ticker.C; !ok {
				break
			}
			if err := client.Ping().Err(); err != nil {
				logDebug("Redis Ping: %v", err)
			}
		}
	})

	logInfo("NewRedis Connect To Redis Success")

	return &Redis{
		client:  client,
		scripts: map[string]*redisScript{},
		ticker:  ticker,
	}
}

type RedisMgrConf struct {
	Platforms map[string][]RedisConf
}

type RedisMgr struct {
	instances map[string][]*Redis
}

func (mgr *RedisMgr) GetRedis(platform string, idx int) *Redis {
	pool, ok := mgr.instances[platform]
	if !ok {
		return nil
	}
	return pool[uint32(idx)%uint32(len(pool))]
}

func (mgr *RedisMgr) ForEach(cb func(*Redis)) {
	for _, pool := range mgr.instances {
		for _, rds := range pool {
			cb(rds)
		}
	}
}

func NewRedisMgr(mgrConf RedisMgrConf) *RedisMgr {
	mgr := &RedisMgr{
		instances: map[string][]*Redis{},
	}

	total := 0
	for platform, confs := range mgrConf.Platforms {
		sort.Sort(RedisConfArr(confs))
		for _, conf := range confs {
			mgr.instances[platform] = append(mgr.instances[platform], NewRedis(conf))
			total++
		}

	}

	if total == 0 {
		panic("invalid RedisMgrConf, 0 config")
	}

	return mgr
}
