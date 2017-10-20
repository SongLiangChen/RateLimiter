package RateLimiter

import (
	"sync/atomic"
	"time"
)

const (
	ONE_SECOND = 1000000000
)

type Bucket struct {
	TokenRemain  int32
	LastSyncTime int64
}

// TokenBucker algorithm
// Refer: https://zhuanlan.zhihu.com/p/20872901?hmsr=toutiao.io&utm_medium=toutiao.io&utm_source=toutiao.io
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
