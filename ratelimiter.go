package RateLimiter

import (
	"sync"
)

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

// **************************
// Public interface
// **************************
func InitRateLimiter(rules Rules) {
	rlt.init(rules)
}

func TakeAccess(uid string, path string) bool {
	return rlt.takeAccess(uid, path)
}
