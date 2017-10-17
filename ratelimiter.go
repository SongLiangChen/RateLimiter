package RateLimiter

import (
	"sync"
	"sync/atomic"
	"time"
)

// Refer: https://zhuanlan.zhihu.com/p/20872901?hmsr=toutiao.io&utm_medium=toutiao.io&utm_source=toutiao.io

const (
	ONE_SECOND = 1000000000
)

type Bucket struct {
	TokenRemain  int32
	LastSyncTime int64
}

func (b *Bucket) resync(c *Rule) int32 {
	now := time.Now().UnixNano()
	remain := atomic.LoadInt32(&b.TokenRemain)
	lasttime := atomic.LoadInt64(&b.LastSyncTime)

	if lasttime == 0 {
		atomic.StoreInt32(&b.TokenRemain, c.Limit-1)
		atomic.StoreInt64(&b.LastSyncTime, now)
		return c.Limit
	}

	if now > lasttime {
		a := int64(c.Duration) * ONE_SECOND / int64(c.Limit)
		tmp := (now - lasttime) / a

		if tmp > 0 {
			remain = remain + int32(tmp)
			if remain > c.Limit {
				remain = c.Limit
			}
			atomic.StoreInt32(&b.TokenRemain, remain)
			atomic.StoreInt64(&b.LastSyncTime, now)
		}
	}

	ret := remain
	if remain > 0 {
		atomic.AddInt32(&b.TokenRemain, -1)
	}
	return ret
}

// The Rule of one uri. the same user can just access this uri 'Limit' times in 'Durantion' time
type Rule struct {
	Limit    int32
	Duration int32 // Units per second
}

type Rules map[string][]*Rule

func (rules Rules) AddRule(path string, rule *Rule) {
	if _, ok := rules[path]; !ok {
		rules[path] = make([]*Rule, 0)
	}

	rules[path] = append(rules[path], rule)
}

func NewRules() Rules {
	return make(Rules)
}

type RateLimiter struct {
	// filiter store the status of access for every user
	filiter map[string]map[string][]*Bucket
	// rules'key is the uri that need frequency verification, and the rules'value is the Rules of uri(see Rule struct)
	rules Rules

	sync.RWMutex
}

func (r *RateLimiter) init(rules Rules) {
	r.rules = rules
}

func (r *RateLimiter) getBuckets(uid string, path string) []*Bucket {
	var b map[string][]*Bucket
	var ok bool

	r.RLock()
	if b, ok = r.filiter[uid]; !ok {
		r.RUnlock()

		r.Lock()
		r.filiter[uid] = make(map[string][]*Bucket)

		for path, rs := range r.rules {
			r.filiter[uid][path] = make([]*Bucket, 0)
			for i := 0; i < len(rs); i++ {
				r.filiter[uid][path] = append(r.filiter[uid][path], new(Bucket))
			}
		}

		b = r.filiter[uid]
		r.Unlock()

	} else {
		r.RUnlock()
	}

	return b[path]
}

func (r *RateLimiter) takeAccess(uid string, path string) bool {
	if _, ok := r.rules[path]; !ok {
		return true
	}

	bs := r.getBuckets(uid, path)

	for i, b := range bs {
		if b.resync(r.rules[path][i]) <= 0 {
			return false
		}
	}

	return true
}

func newRateLimiter() *RateLimiter {
	r := new(RateLimiter)
	r.filiter = make(map[string]map[string][]*Bucket)
	r.rules = make(map[string][]*Rule)
	return r
}

// RateLimit instance
var rlt *RateLimiter

func init() {
	rlt = newRateLimiter()
}

func InitRateLimiter(rules Rules) {
	rlt.init(rules)
}

func TakeAccess(uid string, path string) bool {
	return rlt.takeAccess(uid, path)
}
