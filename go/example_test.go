package g_test

import (
	"fmt"
	"time"

	g "github.com/shanyux/leaf/go"
)

func Example() {
	d := g.New(10)

	// go 1
	var res int
	d.GoToExec(func() {
		fmt.Println("1 + 1 = ?")
		res = 1 + 1
	}, func() {
		fmt.Println(res)
	})

	d.Cb(<-d.ChanCb)

	// go 2
	d.GoToExec(func() {
		fmt.Print("My name is ")
	}, func() {
		fmt.Println("Leaf")
	})

	d.Close()

	// Output:
	// 1 + 1 = ?
	// 2
	// My name is Leaf
}

func ExampleLinearContext() {
	d := g.New(10)

	// parallel
	d.GoToExec(func() {
		time.Sleep(time.Second / 2)
		fmt.Println("1")
	}, nil)
	d.GoToExec(func() {
		fmt.Println("2")
	}, nil)

	d.Cb(<-d.ChanCb)
	d.Cb(<-d.ChanCb)

	// linear
	c := d.NewLinearContext()
	c.GoToExec(func() {
		time.Sleep(time.Second / 2)
		fmt.Println("1")
	}, nil)
	c.GoToExec(func() {
		fmt.Println("2")
	}, nil)

	d.Close()

	// Output:
	// 2
	// 1
	// 1
	// 2
}
