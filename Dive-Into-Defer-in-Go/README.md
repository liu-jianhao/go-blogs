# 深入理解defer

## 认识defer
先看一个简单的例子：
```go
func example1() {
	defer fmt.Println("later")

	fmt.Println("first")
}
```
输出结果：
```
first
later
```
可以看到，在defer之后的语句是在函数结束前执行的

## defer可以干什么
### 释放资源
defer经常用于需要释放的资源
```go
func example2() {
	resp, err := http.Get("http://baidu.com")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	fmt.Println(resp.Body)
	return
}
```

### panic恢复
```go
func example3() {
	defer func() {
		if err := recover(); err != nil {
			// panic prevented stack trace is intact
		}
	}()

	panic("oops!")
}
```
`recover()`返回提供给`panic()`的值，你可以根据这个值决定要执行操作，你还可以用其他`panic`类型，程序panic的时候根据这个值就能知道是由什么错误引起的

### defer闭包
```go
func example4() {
	num := 20
	defer func() {
		fmt.Println(num)
	}()

	num = 8
}
```
输出结果：
```
8
```

### 参数传递
```go
func example5() {
	var n int
	i := 10

	defer func(i int) {
		n = n + i // i = 10  n = 20
		fmt.Println(n)
	}(i) // i = 10 n = 0

	i = i * 2 // i = 20
	n = i     // n = 20
}
```
输出结果：
```go
30
```

### 多个defer
多个defer会保存为一个栈，所以最后一个defer语句会最先执行

```go
func example6() {
	defer fmt.Println("last")
	defer fmt.Println("first")
}
```
输出结果：
```
first
last
```

### 方法使用defer
```go
type car struct {
	model string
}

func (c car) PrintModel() {
	fmt.Println(c.model)
}

func example7() {
	c := car{model: "DeLorean DMC-12"}
	defer c.PrintModel()
	c.model = "Chevrolet Impala"
}
```
输出结果：
```
DeLorean DMC-12
```
如果改为指针接受者：
```go
func (c *Car) PrintModel() {
  fmt.Println(c.model)
}
```
输出结果：
```
Chevrolet Impala
``