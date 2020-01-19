## What is singleflight?
缓存大家都一定不陌生，缓存又一个过期时间，未命中缓存就从数据源获取结果更新新的缓存，
这是开发中的常规操作。

比如有个帖子突然暴火，这个贴子的转发、评论等等这些与这个帖子相关的操作的请求就会突然暴增，
要保证这个服务以及下游服务挺过这个高峰时期，缓存就必不可少了，问题就来了，
怎么知道哪些请求的内容需要缓存下来，缓存的过期时间和更新怎么确定？

这时候就轮到singleflight登场了。

### 初探singleflight
我们实现一个简单的HTTP请求的程序：
```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/baidu", func(w http.ResponseWriter, r *http.Request) {
		status, err := baiduStatus()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Baidu Status: %v", status)
	})

	http.ListenAndServe("127.0.0.1:8080", nil)
}

func baiduStatus() (int, error) {
	log.Println("Making request to Baidu API")
	defer log.Println("Request to Baidu API Complete")

	time.Sleep(1 * time.Second)

    resp, err := http.Get("https://www.baidu.com")
	if err != nil {
        log.Println("get baidu.com error")
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, err
}
```

```
> go run main.go
```
打开另一个终端请求（用到了工具vegeta）：
```
> echo "GET http://localhost:8080/baidu" | vegeta attack -duration=1s -rate=10 | vegeta report
Requests      [total, rate, throughput]  10, 11.07, 7.98
Duration      [total, attack, wait]      1.25344005s, 903.330824ms, 350.109226ms
Latencies     [mean, 50, 95, 99, max]    801.031033ms, 799.323498ms, 1.25329694s, 1.25329694s, 1.25329694s
Bytes In      [total, mean]              180, 18.00
Bytes Out     [total, mean]              0, 0.00
Success       [ratio]                    100.00%
Status Codes  [code:count]               200:10
Error Set:
```

输出结果：
```go
2020/01/19 16:35:19 Making request to Baidu API
2020/01/19 16:35:19 Making request to Baidu API
2020/01/19 16:35:19 Making request to Baidu API
2020/01/19 16:35:20 Making request to Baidu API
2020/01/19 16:35:20 Making request to Baidu API
2020/01/19 16:35:20 Making request to Baidu API
2020/01/19 16:35:20 Making request to Baidu API
2020/01/19 16:35:20 Making request to Baidu API
2020/01/19 16:35:20 Making request to Baidu API
2020/01/19 16:35:20 Making request to Baidu API
2020/01/19 16:35:20 Request to Baidu API Complete
2020/01/19 16:35:20 Request to Baidu API Complete
2020/01/19 16:35:20 Request to Baidu API Complete
2020/01/19 16:35:21 Request to Baidu API Complete
2020/01/19 16:35:21 Request to Baidu API Complete
2020/01/19 16:35:21 Request to Baidu API Complete
2020/01/19 16:35:21 Request to Baidu API Complete
2020/01/19 16:35:21 Request to Baidu API Complete
2020/01/19 16:35:21 Request to Baidu API Complete
2020/01/19 16:35:21 Request to Baidu API Complete
```
可以看到请求了10次http请求

现在我们用上singleflight：
```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

    "golang.org/x/sync/singleflight"
)

func main() {
	var requestGroup singleflight.Group

	http.HandleFunc("/baidu", func(w http.ResponseWriter, r *http.Request) {
        // 传入一个key做标记
		v, err, shared := requestGroup.Do("baidu", func() (interface{}, error) {
			return baiduStatus()
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		status := v.(int)

		log.Printf("/baidu handler requst: status %v, shared result %t", status, shared)

		fmt.Fprintf(w, "Baidu Status: %q", status)
	})

    http.ListenAndServe("127.0.0.1:8080", nil)
}

func baiduStatus() (int, error) {
	log.Println("Making request to Baidu API")
	defer log.Println("Request to Baidu API Complete")

	time.Sleep(1 * time.Second)

    resp, err := http.Get("https://www.baidu.com")
	if err != nil {
        log.Println("get baidu.com error")
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, err
}
```

用同样的请求得到：
```
2020/01/19 16:37:33 Making request to Baidu API
2020/01/19 16:37:34 Request to Baidu API Complete
2020/01/19 16:37:34 /baidu handler requst: status 200, shared result true
2020/01/19 16:37:34 /baidu handler requst: status 200, shared result true
2020/01/19 16:37:34 /baidu handler requst: status 200, shared result true
2020/01/19 16:37:34 /baidu handler requst: status 200, shared result true
2020/01/19 16:37:34 /baidu handler requst: status 200, shared result true
2020/01/19 16:37:34 /baidu handler requst: status 200, shared result true
2020/01/19 16:37:34 /baidu handler requst: status 200, shared result true
2020/01/19 16:37:34 /baidu handler requst: status 200, shared result true
2020/01/19 16:37:34 /baidu handler requst: status 200, shared result true
```
看到只请求了一次http请求。


### 剖析singleflight
```go
// call is an in-flight or completed singleflight.Do call
type call struct {
	wg sync.WaitGroup

	// These fields are written once before the WaitGroup is done
	// and are only read after the WaitGroup is done.
	val interface{}
	err error

	// These fields are read and written with the singleflight
	// mutex held before the WaitGroup is done, and are read but
	// not written after the WaitGroup is done.
	dups  int
	chans []chan<- Result
}
```
+ call用来表示一个正在执行或已完成的函数调用

```go
// Group represents a class of work and forms a namespace in
// which units of work can be executed with duplicate suppression.
type Group struct {
	mu sync.Mutex       // protects m
	m  map[string]*call // lazily initialized
}
```
+ Group是对任务的分类

```go
// Do executes and returns the results of the given function, making
// sure that only one execution is in-flight for a given key at a
// time. If a duplicate comes in, the duplicate caller waits for the
// original to complete and receives the same results.
// The return value shared indicates whether v was given to multiple callers.
func (g *Group) Do(key string, fn func() (interface{}, error)) (v interface{}, err error, shared bool) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		c.dups++
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err, true
	}
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	g.doCall(c, key, fn)
	return c.val, c.err, c.dups > 0
}
```
+ Do函数中，fn函数的返回结果存储在call.val和call.err中

因此，对于缓存的过期和更新问题都放在fn函数中实现。
