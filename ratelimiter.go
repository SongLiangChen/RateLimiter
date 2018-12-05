package RateLimiter

type RateLimiter interface {
	InitRules(rules Rules, cnf ...string) error
	TokenAccess(sessionId string, accessKey string) bool
}

var (
	limiters = make(map[string]RateLimiter)
)

func Register(name string, limiter RateLimiter) {
	if _, ok := limiters[name]; ok {
		panic("repeat name of limiter")
	}
	limiters[name] = limiter
}

func NewRateLimiter(name string) (RateLimiter, bool) {
	r, ok := limiters[name]
	return r, ok
}
