package main

import (
	"fmt"
	"net/http"
)

func example1() {
	defer fmt.Println("later")

	fmt.Println("first")
}

func example2() {
	resp, err := http.Get("http://baidu.com")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	fmt.Println(resp.Body)
	return
}

func example3() {
	defer func() {
		if err := recover(); err != nil {
			// panic prevented stack trace is intact
			fmt.Println(err)
		}
	}()

	panic("oops!")
}

func example4() {
	num := 20
	defer func() {
		fmt.Println(num)
	}()

	num = 8
}

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

func example6() {
	defer fmt.Println("last")
	defer fmt.Println("first")
}

type car struct {
	model string
}

func (c car) PrintModel() {
	fmt.Println(c.model)
}

/*
func (c *car) PrintModel() {
	fmt.Println(c.model)
}
*/

func example7() {
	c := car{model: "DeLorean DMC-12"}
	defer c.PrintModel()
	c.model = "Chevrolet Impala"
}

func main() {
	fmt.Println("-----------Example1------------")
	example1()
	fmt.Println()

	fmt.Println("-----------Example2------------")
	example2()
	fmt.Println()

	fmt.Println("-----------Example3------------")
	example3()
	fmt.Println()

	fmt.Println("-----------Example4------------")
	example4()
	fmt.Println()

	fmt.Println("-----------Example5------------")
	example5()
	fmt.Println()

	fmt.Println("-----------Example6------------")
	example6()
	fmt.Println()

	fmt.Println("-----------Example7------------")
	example7()
	fmt.Println()
}
