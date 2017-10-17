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

func (b *Bucket) resync(c *Config) int32 {
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

// The config of one uri. the same user can just access this uri 'Limit' times in 'Durantion' time
type Config struct {
	Limit    int32
	Duration int32 // Units per second
}

type RateLimiter struct {
	// Filiter store the status of access for every user
	Filiter map[int]map[string]*Bucket
	// Path'key is the uri that need frequency verification, and the Path'value is the config of uri(see Config struct)
	Path map[string]*Config

	sync.RWMutex
}

func (r *RateLimiter) init(p map[string]*Config) {
	for pp, c := range p {
		r.Path[pp] = c
	}
}

func (r *RateLimiter) getBucket(uid int, path string) *Bucket {
	var b map[string]*Bucket
	var ok bool

	r.RLock()
	if b, ok = r.Filiter[uid]; !ok {
		r.RUnlock()

		r.Lock()
		r.Filiter[uid] = make(map[string]*Bucket)
		for p, _ := range r.Path {
			r.Filiter[uid][p] = new(Bucket)
		}
		b = r.Filiter[uid]
		r.Unlock()

	} else {
		r.RUnlock()
	}

	return b[path]
}

func (r *RateLimiter) takeAccess(uid int, path string) bool {
	if _, ok := r.Path[path]; !ok {
		return true
	}

	b := r.getBucket(uid, path)

	ac := b.resync(r.Path[path])
	if ac > 0 {
		return true
	}

	return false
}

func newRateLimiter() *RateLimiter {
	r := new(RateLimiter)
	r.Filiter = make(map[int]map[string]*Bucket)
	r.Path = make(map[string]*Config)
	return r
}

// RateLimit instance
var rlt *RateLimiter

func init() {
	rlt = newRateLimiter()
}

func InitRateLimiter(p map[string]*Config) {
	rlt.init(p)
}

func TakeAccess(uid int, path string) bool {
	return rlt.takeAccess(uid, path)
}
