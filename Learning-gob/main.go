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
