## gob
gob是Go语言自带的一个数据结构序列化的编解码工具，
与json和protobuf类似，相比之下特点是简单，但是只能用于Go语言，
即只能用Go语言来编码和解码，最常用的场景就是RPC。

### example
```go
package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type animalI interface {
	say()
}

type cat struct {
}

func (c *cat) say() {
	fmt.Println("Meow")
}

type dog struct {
}

func (d *dog) say() {
	fmt.Println("Woo")
}

func init() {
	// 当编码中有字段是interface时，需要对interface的可能产生的类型进行注册
	gob.Register(&cat{})
	gob.Register(&dog{})
}

func main() {
	network := new(bytes.Buffer)
	enc := gob.NewEncoder(network)

	var animal animalI
	animal = new(cat)
	if err := enc.Encode(&animal); err != nil {
		panic(err)
	}

	animal = new(dog)
	if err := enc.Encode(&animal); err != nil {
		panic(err)
	}

	dec := gob.NewDecoder(network)

	var getAnimal animalI
	if err := dec.Decode(&getAnimal); err != nil {
		panic(err)
	}
	getAnimal.say()

	if err := dec.Decode(&getAnimal); err != nil {
		panic(err)
	}
	getAnimal.say()
}
```

输出：
```
Meow
Woo
```
