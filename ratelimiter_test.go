package RateLimiter

import (
	"testing"
	"time"
)

func Test_RateLimiter(t *testing.T) {
	r, ok := NewRateLimiter("redis")
	if !ok {
		t.Fatalf("不存在")
		return
	}
	rules := NewRules()
	rules.AddRule("abc", &Rule{
		1,
		1,
	})
	if err := r.InitRules(rules, "127.0.0.1:6379", "", "0", "100", "30"); err != nil {
		t.Fatalf("err: %v", err.Error())
		return
	}

	t.Log(r.TokenAccess("a", "abc"))
	t.Log(r.TokenAccess("a", "abc"))
	t.Log(r.TokenAccess("a", "abc"))

	time.Sleep(time.Second * 2)
	t.Log(r.TokenAccess("a", "abc"))
}
