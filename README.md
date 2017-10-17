# RateLimiter
# 基于Token Bucket算法实现的api限流模块

目前该版本对同一条uri，只能设置一种策略：某个时间段D内，允许某个用户访问N次。

使用：

go get github.com/SongLiangChen/RateLimiter


示例：
```
package main

import (
	"net/http"
	"strconv"

	"github.com/SongLiangChen/RateLimiter"
)

func main() {
	access_config := make(map[string]*RateLimiter.Config)

	// 对test 配置60秒内只允许访问100次
	access_config["test"] = &RateLimiter.Config{
		Duration: 60,
		Limit:    100,
	}

	// 对test/abc 配置一小时内允许访问10次
	access_config["test/abc"] = &RateLimiter.Config{
		Duration: 3600,
		Limit:    10,
	}

	RateLimiter.InitRateLimiter(access_config)

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		uid, _ := strconv.Atoi(r.FormValue("uid"))
		if !RateLimiter.TakeAccess(uid, "/test") {
			w.Write([]byte("请求太频繁"))
		}
	})

	http.HandleFunc("/test/abc", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		uid, _ := strconv.Atoi(r.FormValue("uid"))
		if !RateLimiter.TakeAccess(uid, "/test/abc") {
			w.Write([]byte("请求太频繁"))
		}
	})

	http.ListenAndServe(":8080", nil)
}
```
