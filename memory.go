package RateLimiter

import (
	"sync"
)

type MemRateLimiter struct {
	// filiter store the status of access for every user
	filiter map[string]map[string][]*Bucket
	// rules'key is the uri that need frequency verification, and the rules'value is the Rules of uri(see Rule struct)
	rules Rules

	sync.RWMutex
}

func (r *MemRateLimiter) InitRules(rules Rules, cnf ...string) error {
	r.rules = rules
	return nil
}

func (r *MemRateLimiter) TokenAccess(sessionId string, accessKey string) bool {
	if _, ok := r.rules[accessKey]; !ok {
		return true
	}

	bs := r.getBuckets(sessionId, accessKey)

	for i, b := range bs {
		if b.resync(r.rules[accessKey][i]) <= 0 {
			return false
		}
	}

	return true
}

func (r *MemRateLimiter) getBuckets(sessionId string, accessKey string) []*Bucket {
	var b map[string][]*Bucket
	var ok bool

	r.RLock()
	if b, ok = r.filiter[sessionId]; !ok {
		r.RUnlock()

		r.Lock()
		r.filiter[sessionId] = make(map[string][]*Bucket)

		for path, rs := range r.rules {
			r.filiter[sessionId][path] = make([]*Bucket, 0)
			for i := 0; i < len(rs); i++ {
				r.filiter[sessionId][path] = append(r.filiter[sessionId][path], new(Bucket))
			}
		}

		b = r.filiter[sessionId]
		r.Unlock()

	} else {
		r.RUnlock()
	}

	return b[accessKey]
}

func newMemRateLimiter() *MemRateLimiter {
	r := new(MemRateLimiter)
	r.filiter = make(map[string]map[string][]*Bucket)
	r.rules = NewRules()
	return r
}

func init() {
	Register("memory", newMemRateLimiter())
}
