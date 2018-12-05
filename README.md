# RateLimiter
## 基于Token Bucket算法实现的api限流模块

## 功能：
1、可以针对某个uri自定义各种限流规则，例如：规定 10 秒内，用户的请求次数不能超过 200 次；而且，1 小时内，用户的请求次数不能超过 5000 次；并且，1 天内， 用户的请求次数不能超过 20000 次


示例：
```
package main

import (
	"net/http"

	"github.com/SongLiangChen/RateLimiter"
)

func main() {
	rules := RateLimiter.NewRules()
	// 规定任何用户1s内只允许访问5次
	rules.AddRule("/test", &RateLimiter.Rule{
		Duration: 1,
		Limit:    5,
	})
	// 同时规定任何用户10s内只能访问10次
	rules.AddRule("/test", &RateLimiter.Rule{
		Duration: 10,
		Limit:    5,
	})

	r, _ := RateLimiter.NewRateLimiter("momory")
    	r.InitRules(rules)
    
    	// redis
    	// r, _ := RateLimiter.NewRateLimiter("redis")
    	// r.InitRules(rules, "127.0.0.1:6379", "", "0", "10", "20")
    
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if !r.TokenAccess(r.FormValue("uid"), "/test") {
			w.Write([]byte("请求太频繁"))
			return
		}

		// ...do your work
	})

	http.ListenAndServe(":8080", nil)
}

```


2、可以即针对某个uri设置限流规则，同时可以为全局uri设置限流规则

示例
```
package main

import (
	"net/http"

	"github.com/SongLiangChen/RateLimiter"
)

func main() {
	rules := RateLimiter.NewRules()
	// 规定任何用户1s内只允许访问5次/test1
	rules.AddRule("/test1", &RateLimiter.Rule{
		Duration: 1,
		Limit:    5,
	})
	// 同时规定任何用户10s内只能访问10次/test2
	rules.AddRule("/test2", &RateLimiter.Rule{
		Duration: 10,
		Limit:    5,
	})
	// 并且规定对任何uri的访问，60s内只能访问10次
	rules.AddRule("", &RateLimiter.Rule{
		Duration: 60,
		Limit:    10,
	})

	r, _ := RateLimiter.NewRateLimiter("momory")
    	r.InitRules(rules)
    
    	// redis
    	// r, _ := RateLimiter.NewRateLimiter("redis")
    	// r.InitRules(rules, "127.0.0.1:6379", "", "0", "10", "20")

	http.HandleFunc("/test1", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		uid := r.FormValue("uid")
		if !RateLimiter.TokenAccess(uid, "/test1") || !RateLimiter.TokenAccess(uid, "") {
			w.Write([]byte("请求太频繁"))
			return
		}

		// ...do your work
	})

	http.HandleFunc("/test2", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		uid := r.FormValue("uid")
		if !RateLimiter.TokenAccess(uid, "/test2") || !RateLimiter.TokenAccess(uid, "") {
			w.Write([]byte("请求太频繁"))
			return
		}

		// ...do your work
	})

	http.ListenAndServe(":8080", nil)
}

```
