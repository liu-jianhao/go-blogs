# Go语言的50个“坑”
最近看到了这篇文章，比较适合Go初学者，[英文原文](http://devs.cloudimmunity.com/gotchas-and-common-mistakes-in-go-golang/)，
[中文译文](https://blog.csdn.net/stpeace/article/details/81675028)

下面是我读这篇文章做的笔记，都是一些我之前没注意到的Go语言的特性
## 目录
+ [简短声明的变量只能在函数内部使用](#简短声明的变量只能在函数内部使用)
+ [不能使用简短声明来设置字段的值](#不能使用简短声明来设置字段的值)
+ [不小心覆盖了变量](#不小心覆盖了变量)
+ [显式类型的变量无法使用nil来初始化](#显式类型的变量无法使用nil来初始化)
+ [直接使用值为nil的slice、map](#直接使用值为nil的slice、map)
+ [string类型的变量值不能为nil](#string类型的变量值不能为nil)
+ [range遍历slice和array时混淆了返回值](#range遍历slice和array时混淆了返回值)
+ [访问map中不存在的key](#访问map中不存在的key)
+ [不导出的struct字段无法被encode](#不导出的struct字段无法被encode)
+ [向已关闭的channel发送数据会造成panic](#向已关闭的channel发送数据会造成panic)
+ [使用了值为nil的channel](#使用了值为nil的channel)
+ [将JSON中的数字解码为interface类型](#将JSON中的数字解码为interface类型)
+ [在range迭代slice、array、map时通过更新引用来更新元素](#在range迭代slice、array、map时通过更新引用来更新元素)
+ [类型声明与方法](#类型声明与方法)
+ [跳出for-switch和for-select代码块](#跳出for-switch和for-select代码块)
+ [defer函数的参数值](#defer函数的参数值)
+ [阻塞的gorutinue与资源泄露](#阻塞的gorutinue与资源泄露)
+ [使用指针作为方法的receiver](#使用指针作为方法的receiver)
+ [更新map字段的值](#更新map字段的值)

## 简短声明的变量只能在函数内部使用
```go
// 错误示例
myvar := 1	// syntax error: non-declaration statement outside function body
func main() {
}

// 正确示例
var  myvar = 1
func main() {
}
```

## 不能使用简短声明来设置字段的值
struct 的变量字段不能使用 := 来赋值以使用预定义的变量来避免解决：
```go
// 错误示例
type info struct {
	result int
}
 
func work() (int, error) {
	return 3, nil
}
 
func main() {
	var data info
	data.result, err := work()	// error: non-name data.result on left side of :=
	fmt.Printf("info: %+v\n", data)
}
 
 
// 正确示例
func main() {
	var data info
	var err error	// err 需要预声明
 
	data.result, err = work()
	if err != nil {
		fmt.Println(err)
		return
	}
 
	fmt.Printf("info: %+v\n", data)
}
```

## 不小心覆盖了变量
对从动态语言转过来的开发者来说，简短声明很好用，这可能会让人误会 := 是一个赋值操作符。

如果你在新的代码块中像下边这样误用了 :=，编译不会报错，但是变量不会按你的预期工作：
```go
package main

import "fmt"

func main() {
	x := 1
	fmt.Println(x)		// 1
	{
		fmt.Println(x)	// 1
		x := 2
		fmt.Println(x)	// 2	// 新的 x 变量的作用域只在代码块内部
	}
	fmt.Println(x)		// 1
}
```
可使用 vet 工具来诊断这种变量覆盖，Go 默认不做覆盖检查，添加 -shadow 选项来启用：
```
> lock % go tool vet -shadow main.go
main.go:10: declaration of "x" shadows declaration at main.go:6
```

## 显式类型的变量无法使用nil来初始化
nil 是 interface、function、pointer、map、slice 和 channel 类型变量的默认初始值。但声明时不指定类型，编译器也无法推断出变量的具体类型。
```go
// 错误示例
func main() {
    var x = nil	// error: use of untyped nil
	_ = x
}

// 正确示例
func main() {
	var x interface{} = nil
	_ = x
}
```

## 直接使用值为nil的slice、map
允许对值为 nil 的 slice 添加元素，但对值为 nil 的 map 添加元素则会造成运行时 panic
```go
// map 错误示例
func main() {
    var m map[string]int
    m["one"] = 1		// error: panic: assignment to entry in nil map
    // m := make(map[string]int)// map 的正确声明，分配了实际的内存
}

// slice 正确示例
func main() {
	var s []int
	s = append(s, 1)
}
```

## string类型的变量值不能为nil
对那些喜欢用 nil 初始化字符串的人来说，这就是坑：
```go
// 错误示例
func main() {
	var s string = nil	// cannot use nil as type string in assignment
	if s == nil {	// invalid operation: s == nil (mismatched types string and nil)
		s = "default"
	}
}

// 正确示例
func main() {
	var s string	// 字符串类型的零值是空串 ""
	if s == "" {
		s = "default"
	}
}
```

## range遍历slice和array时混淆了返回值
与其他编程语言中的 for-in 、foreach 遍历语句不同，Go 中的 range 在遍历时会生成 2 个值，第一个是元素索引，第二个是元素的值：
```go
// 错误示例
func main() {
	x := []string{"a", "b", "c"}
	for v := range x {
		fmt.Println(v)	// 1 2 3
	}
}

// 正确示例
func main() {
	x := []string{"a", "b", "c"}
	for _, v := range x {	// 使用 _ 丢弃索引
		fmt.Println(v)
	}
}
```

## 访问map中不存在的key
Go 则会返回元素对应数据类型的零值，比如 nil、'' 、false 和 0，取值操作总有值返回，故不能通过取出来的值来判断 key 是不是在 map 中。

检查 key 是否存在可以用 map 直接访问，检查返回的第二个参数即可：
```go
// 错误的 key 检测方式
func main() {
	x := map[string]string{"one": "2", "two": "", "three": "3"}
	if v := x["two"]; v == "" {
		fmt.Println("key two is no entry")	// 键 two 存不存在都会返回的空字符串
	}
}
 
// 正确示例
func main() {
	x := map[string]string{"one": "2", "two": "", "three": "3"}
	if _, ok := x["two"]; !ok {
		fmt.Println("key two is no entry")
	}
}
```

## 不导出的struct字段无法被encode
以小写字母开头的字段成员是无法被外部直接访问的，所以 struct 在进行 json、xml、gob 等格式的 encode 操作时，这些私有字段会被忽略，导出时得到零值：
```go
func main() {
	in := MyData{1, "two"}
	fmt.Printf("%#v\n", in)	// main.MyData{One:1, two:"two"}
 
	encoded, _ := json.Marshal(in)
	fmt.Println(string(encoded))	// {"One":1}	// 私有字段 two 被忽略了
 
	var out MyData
	json.Unmarshal(encoded, &out)
	fmt.Printf("%#v\n", out) 	// main.MyData{One:1, two:""}
}
```

## 向已关闭的channel发送数据会造成panic
从已关闭的 channel 接收数据是安全的：

接收状态值 ok 是 false 时表明 channel 中已没有数据可以接收了。类似的，从有缓冲的 channel 中接收数据，缓存的数据获取完再没有数据可取时，状态值也是 false

向已关闭的 channel 中发送数据会造成 panic：
```go
func main() {
	ch := make(chan int)
	for i := 0; i < 3; i++ {
		go func(idx int) {
			ch <- idx
		}(i)
	}
 
	fmt.Println(<-ch)		// 输出第一个发送的值
	close(ch)			// 不能关闭，还有其他的 sender
	time.Sleep(2 * time.Second)	// 模拟做其他的操作
}
```
运行结果：
```
> go run main.go
2
panic: send on closed channel

goroutine 19 [running]:
main.main.func1(0xc42007c060, 0x1)
	/home/liu/Desktop/programing/go/src/test/lock/main.go:10 +0x3f
created by main.main
	/home/liu/Desktop/programing/go/src/test/lock/main.go:9 +0x6f
exit status 2
```

## 使用了值为nil的channel
在一个值为 nil 的 channel 上发送和接收数据将永久阻塞：
```go
func main() {
	var ch chan int // 未初始化，值为 nil
	for i := 0; i < 3; i++ {
		go func(i int) {
			ch <- i
		}(i)
	}
 
	fmt.Println("Result: ", <-ch)
	time.Sleep(2 * time.Second)
}
```
runtime 死锁错误：
```
fatal error: all goroutines are asleep - deadlock! goroutine 1 [chan receive (nil chan)]
```
利用这个死锁的特性，可以用在 select 中动态的打开和关闭 case 语句块：
```go
func main() {
	inCh := make(chan int)
	outCh := make(chan int)
 
	go func() {
		var in <-chan int = inCh
		var out chan<- int
		var val int
 
		for {
			select {
			case out <- val:
				println("--------")
				out = nil
				in = inCh
			case val = <-in:
				println("++++++++++")
				out = outCh
				in = nil
			}
		}
	}()
 
	go func() {
		for r := range outCh {
			fmt.Println("Result: ", r)
		}
	}()
 
	time.Sleep(0)
	inCh <- 1
	inCh <- 2
	time.Sleep(3 * time.Second)
}
```
运行结果：
```
> go run main.go
++++++++++
--------
Result:  1
++++++++++
--------
Result:  2
```

## 将JSON中的数字解码为interface类型
在 encode/decode JSON 数据时，Go 默认会将数值当做 float64 处理，比如下边的代码会造成 panic：
```go
func main() {
	var data = []byte(`{"status": 200}`)
	var result map[string]interface{}
 
	if err := json.Unmarshal(data, &result); err != nil {
		log.Fatalln(err)
	}
 
	fmt.Printf("%T\n", result["status"])	// float64
	var status = result["status"].(int)	// 类型断言错误
	fmt.Println("Status value: ", status)
}
```
```
panic: interface conversion: interface {} is float64, not int
```

如果你尝试 decode 的 JSON 字段是整型，你可以：
+ 将 int 值转为 float 统一使用
+ 将 decode 后需要的 float 值转为 int 使用
```go
// 将 decode 的值转为 int 使用
func main() {
    var data = []byte(`{"status": 200}`)
    var result map[string]interface{}
 
    if err := json.Unmarshal(data, &result); err != nil {
        log.Fatalln(err)
    }
 
    var status = uint64(result["status"].(float64))
    fmt.Println("Status value: ", status)
}
```
+ 使用 Decoder 类型来 decode JSON 数据，明确表示字段的值类型
```go
// 指定字段类型
func main() {
	var data = []byte(`{"status": 200}`)
	var result map[string]interface{}
    
	var decoder = json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
 
	if err := decoder.Decode(&result); err != nil {
		log.Fatalln(err)
	}
 
	var status, _ = result["status"].(json.Number).Int64()
	fmt.Println("Status value: ", status)
}
 
 // 你可以使用 string 来存储数值数据，在 decode 时再决定按 int 还是 float 使用
 // 将数据转为 decode 为 string
 func main() {
 	var data = []byte({"status": 200})
  	var result map[string]interface{}
  	var decoder = json.NewDecoder(bytes.NewReader(data))
  	decoder.UseNumber()
  	if err := decoder.Decode(&result); err != nil {
  		log.Fatalln(err)
  	}
    var status uint64
  	err := json.Unmarshal([]byte(result["status"].(json.Number).String()), &status);
	checkError(err)
   	fmt.Println("Status value: ", status)
}
```
+ 使用 struct 类型将你需要的数据映射为数值型
```go
// struct 中指定字段类型
func main() {
  	var data = []byte(`{"status": 200}`)
  	var result struct {
  		Status uint64 `json:"status"`
  	}
 
  	err := json.NewDecoder(bytes.NewReader(data)).Decode(&result)
  	checkError(err)
	fmt.Printf("Result: %+v", result)
}
```
+ 可以使用 struct 将数值类型映射为 json.RawMessage 原生数据类型适用于如果 JSON 数据不着急 decode 或 JSON 某个字段的值类型不固定等情况：
```go
// 状态名称可能是 int 也可能是 string，指定为 json.RawMessage 类型
func main() {
	records := [][]byte{
		[]byte(`{"status":200, "tag":"one"}`),
		[]byte(`{"status":"ok", "tag":"two"}`),
	}
 
	for idx, record := range records {
		var result struct {
			StatusCode uint64
			StatusName string
			Status     json.RawMessage `json:"status"`
			Tag        string          `json:"tag"`
		}
 
		err := json.NewDecoder(bytes.NewReader(record)).Decode(&result)
		checkError(err)
 
		var name string
		err = json.Unmarshal(result.Status, &name)
		if err == nil {
			result.StatusName = name
		}
 
		var code uint64
		err = json.Unmarshal(result.Status, &code)
		if err == nil {
			result.StatusCode = code
		}
 
		fmt.Printf("[%v] result => %+v\n", idx, result)
	}
}
```


## 在range迭代slice、array、map时通过更新引用来更新元素
在 range 迭代中，得到的值其实是元素的一份值拷贝，更新拷贝并不会更改原来的元素，即是拷贝的地址并不是原有元素的地址：
```go
func main() {
	data := []int{1, 2, 3}
	for _, v := range data {
		v *= 10		// data 中原有元素是不会被修改的
	}
	fmt.Println("data: ", data)	// data:  [1 2 3]
}
```
如果要修改原有元素的值，应该使用索引直接访问：
```go
func main() {
	data := []int{1, 2, 3}
	for i, v := range data {
		data[i] = v * 10	
	}
	fmt.Println("data: ", data)	// data:  [10 20 30]
}
```
如果你的集合保存的是指向值的指针，需稍作修改。依旧需要使用索引访问元素，不过可以使用 range 出来的元素直接更新原有值：
```go
func main() {
	data := []*struct{ num int }{{1}, {2}, {3},}
	for _, v := range data {
		v.num *= 10	// 直接使用指针更新
	}
	fmt.Println(data[0], data[1], data[2])	// &{10} &{20} &{30}
}
```

## 类型声明与方法
从一个现有的非 interface 类型创建新类型时，并不会继承原有的方法：
```go
// 定义 Mutex 的自定义类型
type myMutex sync.Mutex
 
func main() {
	var mtx myMutex
	mtx.Lock()
	mtx.UnLock()
}
```
```
mtx.Lock undefined (type myMutex has no field or method Lock)...
```
如果你需要使用原类型的方法，可将原类型以匿名字段的形式嵌到你定义的新 struct 中：
```go
// 类型以字段形式直接嵌入
type myLocker struct {
	sync.Mutex
}
 
func main() {
	var locker myLocker
	locker.Lock()
	locker.Unlock()
}
```
interface 类型声明也保留它的方法集：
```go
type myLocker sync.Locker
 
func main() {
	var locker myLocker
	locker.Lock()
	locker.Unlock()
}
```

## 跳出for-switch和for-select代码块
没有指定标签的 break 只会跳出 switch/select 语句，若不能使用 return 语句跳出的话，可为 break 跳出标签指定的代码块：
```go
// break 配合 label 跳出指定代码块
func main() {
loop:
	for {
		switch {
		case true:
			fmt.Println("breaking out...")
			//break	// 死循环，一直打印 breaking out...
			break loop
		}
	}
	fmt.Println("out...")
}
```

## defer函数的参数值
对 defer 延迟执行的函数，它的参数会在声明时候就会求出具体值，而不是在执行时才求值：
```go
// 在 defer 函数中参数会提前求值
func main() {
	var i = 1
	defer fmt.Println("result: ", func() int { return i * 2 }())
	i++
}
```

## 阻塞的gorutinue与资源泄露
```go
func First(query string, replicas []Search) Result {
	c := make(chan Result)
	replicaSearch := func(i int) { c <- replicas[i](query) }
	for i := range replicas {
		go replicaSearch(i)
	}
	return <-c
}
```
在搜索重复时依旧每次都起一个 goroutine 去处理，每个 goroutine 都把它的搜索结果发送到结果 channel 中，channel 中收到的第一条数据会直接返回。

返回完第一条数据后，其他 goroutine 的搜索结果怎么处理？他们自己的协程如何处理？

在 First() 中的结果 channel 是无缓冲的，这意味着只有第一个 goroutine 能返回，由于没有 receiver，其他的 goroutine 会在发送上一直阻塞。如果你大量调用，则可能造成资源泄露。

为避免泄露，你应该确保所有的 goroutine 都能正确退出，有 2 个解决方法：
+ 使用带缓冲的 channel，确保能接收全部 goroutine 的返回结果：
```go
func First(query string, replicas ...Search) Result {  
    c := make(chan Result,len(replicas))	
    searchReplica := func(i int) { c <- replicas[i](query) }
    for i := range replicas {
        go searchReplica(i)
    }
    return <-c
}
```
+ 使用 select 语句，配合能保存一个缓冲值的 channel default 语句：
default 的缓冲 channel 保证了即使结果 channel 收不到数据，也不会阻塞 goroutine
```go
func First(query string, replicas ...Search) Result {  
    c := make(chan Result,1)
    searchReplica := func(i int) { 
        select {
        case c <- replicas[i](query):
        default:
        }
    }
    for i := range replicas {
        go searchReplica(i)
    }
    return <-c
}
```
+ 使用特殊的废弃（cancellation） channel 来中断剩余 goroutine 的执行：
```go
func First(query string, replicas ...Search) Result {  
    c := make(chan Result)
    done := make(chan struct{})
    defer close(done)
    searchReplica := func(i int) { 
        select {
        case c <- replicas[i](query):
        case <- done:
        }
    }
    for i := range replicas {
        go searchReplica(i)
    }
 
    return <-c
}
```

## 使用指针作为方法的receiver
只要值是可寻址的，就可以在值上直接调用指针方法。即是对一个方法，它的 receiver 是指针就足矣。

但不是所有值都是可寻址的，比如 map 类型的元素、通过 interface 引用的变量：
```go
type data struct {
	name string
}
 
type printer interface {
	print()
}
 
func (p *data) print() {
	fmt.Println("name: ", p.name)
}
 
func main() {
	d1 := data{"one"}
	d1.print()	// d1 变量可寻址，可直接调用指针 receiver 的方法
 
	var in printer = data{"two"}
	in.print()	// 类型不匹配
 
	m := map[string]data{
		"x": data{"three"},
	}
	m["x"].print()	// m["x"] 是不可寻址的	// 变动频繁
}
```
```
cannot use data literal (type data) as type printer in assignment:

data does not implement printer (print method has pointer receiver)

cannot call pointer method on m["x"] cannot take the address of m["x"]
```

## 更新map字段的值
如果 map 一个字段的值是 struct 类型，则无法直接更新该 struct 的单个字段：
```go
// 无法直接更新 struct 的字段值
type data struct {
	name string
}
 
func main() {
	m := map[string]data{
		"x": {"Tom"},
	}
	m["x"].name = "Jerry"
}
```

```go
cannot assign to struct field m["x"].name in map
```
因为 map 中的元素是不可寻址的。需区分开的是，slice 的元素可寻址：
```go
type data struct {
	name string
}
 
func main() {
	s := []data{{"Tom"}}
	s[0].name = "Jerry"
	fmt.Println(s)	// [{Jerry}]
}
```
更新 map 中 struct 元素的字段值，有 2 个方法：
+ 使用局部变量
```go
// 提取整个 struct 到局部变量中，修改字段值后再整个赋值
type data struct {
	name string
}
 
func main() {
	m := map[string]data{
		"x": {"Tom"},
	}
	r := m["x"]
	r.name = "Jerry"
	m["x"] = r
	fmt.Println(m)	// map[x:{Jerry}]
}
```
+ 使用指向元素的 map 指针
```go
func main() {
	m := map[string]*data{
		"x": {"Tom"},
	}
	
	m["x"].name = "Jerry"	// 直接修改 m["x"] 中的字段
	fmt.Println(m["x"])	// &{Jerry}
}
```
但是要注意下边这种误用：
```go
func main() {
	m := map[string]*data{
		"x": {"Tom"},
	}
	m["z"].name = "what???"	 
	fmt.Println(m["x"])
}
```
panic: runtime error: invalid memory address or nil pointer dereference
