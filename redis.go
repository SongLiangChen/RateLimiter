package RateLimiter

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

// 一共有三个key，分别是
// KEYS[1] = sessionId/controller/action组成的字符串
// KEYS[2] = 当前时间戳(微秒)
// KEYS[3] = 参数个数
//
// 参数的设计，奇数下标代表duration,偶数下标代表limit
// 例如
// ARGV[1] = 10,ARGC[2] = 3. 代表10秒内允许访问3次
// 一共有KEYS[3]/2条配置
//
// 返回1表示成功、返回0表示失败
var luaScript = `
local time_now = tonumber(KEYS[1])
local N = tonumber(KEYS[2])
local remains = {}
local lastTimes = {}
local durations = {}
local limits = {}

local j = 1
for i=3, N*3+2, 3 do
	durations[j] = tonumber(KEYS[i+1])
	limits[j] = tonumber(KEYS[i+2])
	j = j+1
end
j = j-1

-- 遍历每一条规则，判断是否还有token剩余，是否满足重新补充的条件
for i=1, N, 1 do
	local ratelimit_info=redis.pcall("HMGET",KEYS[i*3],"remain_token","last_fill_time")
	remains[i] = tonumber(ratelimit_info[1])
	lastTimes[i] = tonumber(ratelimit_info[2])

	-- 之前不存在，创建，并设置过期为一小时
	if (lastTimes[i]==nil) then
    	redis.call("HMSET",KEYS[i*3],"remain_token",limits[i],"last_fill_time",time_now)
    	redis.call("EXPIRE", KEYS[i*3], 3600)
    	lastTimes[i] = time_now
    	remains[i] = limits[i]
	end

	-- 剩余token不足，判断是否需要补充
	if (remains[i] == 0) then
		if (time_now>lastTimes[i]) then
			local a,b
		    a = math.floor((durations[i]*1000000000)/limits[i])
		    remains[i],b = math.modf((time_now - lastTimes[i])/a)
		    if (remains[i]>limits[i]) then
		        remains[i] = limits[i]
		    end
		    lastTimes[i] = time_now
		end
	end

	-- 任意一条规则不满足，则返回失败
	if (remains[i] == 0) then
		return 0
	end
end

-- 对每一条规则减去一个token，并返回成功
for i=1, N, 1 do
	redis.pcall("HMSET", KEYS[i*3],"remain_token",remains[i]-1,"last_fill_time",lastTimes[i])
end

return 1
`

type RedisRateLimiter struct {
	rules     Rules
	scriptSha string
	conn      *redis.Client
}

// InitRules init redis rate limiter
// cnf like redis server addr,password,dbnum,pool size,IdleTimeout second
// e.g. 127.0.0.1:6379,pwd,0,100,30
func (r *RedisRateLimiter) InitRules(rules Rules, cnf ...string) error {
	r.rules = rules

	addr := cnf[0]
	pwd := cnf[1]
	dbNum, err := strconv.Atoi(cnf[2])
	if err != nil {
		return err
	}
	maxConn, err := strconv.Atoi(cnf[3])
	if err != nil {
		return err
	}
	idleTimeout, err := strconv.Atoi(cnf[4])
	if err != nil {
		return err
	}

	r.conn = redis.NewClient(&redis.Options{
		Addr:        addr,
		Password:    pwd,
		DB:          dbNum,
		MaxRetries:  2, // redis命令最多重试三次(<=MaxRetries)
		PoolSize:    maxConn,
		IdleTimeout: time.Duration(idleTimeout) * time.Second,
	})

	if _, err = r.conn.Ping().Result(); err != nil {
		return err
	}

	r.scriptSha, err = r.conn.ScriptLoad(luaScript).Result()
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisRateLimiter) TokenAccess(sessionId string, accessKey string) bool {
	rules := r.rules[accessKey]

	if len(rules) == 0 {
		return true
	}

	key := sessionId + accessKey

	keys := make([]string, 0)
	for _, rule := range rules {
		keys = append(keys, fmt.Sprintf("%v%v%v", key, rule.Duration, rule.Limit))
		keys = append(keys, strconv.Itoa(int(rule.Duration)))
		keys = append(keys, strconv.Itoa(int(rule.Limit)))
	}
	return r.tokenAccess(keys)
}

func (r *RedisRateLimiter) TokenAccessWithRules(sessionId, accessKey string, rules ...Rule) bool {
	key := sessionId + accessKey

	keys := make([]string, 0)
	for _, rule := range rules {
		keys = append(keys, fmt.Sprintf("%v%v%v", key, rule.Duration, rule.Limit))
		keys = append(keys, strconv.Itoa(int(rule.Duration)))
		keys = append(keys, strconv.Itoa(int(rule.Limit)))
	}
	return r.tokenAccess(keys)
}

func (r *RedisRateLimiter) tokenAccess(rules []string) bool {
	keys := []string{strconv.FormatInt(time.Now().UnixNano(), 10), strconv.Itoa(len(rules) / 3)}
	keys = append(keys, rules...)
	val, err := r.conn.EvalSha(r.scriptSha, keys).Int()
	if err != nil {
		return false
	}
	if val == 1 {
		return true
	}
	return false
}

func init() {
	Register("redis", &RedisRateLimiter{
		rules: NewRules(),
	})
}
