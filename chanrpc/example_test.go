package chanrpc_test

import (
	"fmt"

	"github.com/minortune/rubbish/chanrpc"
)

type Foo struct {
	value int
}

func (f *Foo) Get(r *int) {
	*r = f.value
}

func (f *Foo) Set(v int) {
	f.value = v
}

func (f *Foo) GetAndSet(v int, r *int) {
	*r = f.value
	f.value = v
}

func (f *Foo) GetNextID(r *int) {
	f.value++
	*r = f.value
}

func Example() {
	s := chanrpc.NewRpc(new(Foo))

	s.Send("Set", 50)

	num1 := new(int)
	e := s.Call("Get", num1)
	if e != nil {
		fmt.Println(e)
	} else {
		fmt.Println(*num1)
	}

	num2 := new(int)
	e = s.Call("GetAndSet", 1, num2)
	if e != nil {
		fmt.Println(e)
	} else {
		fmt.Println(*num2)
	}

	num3 := new(int)
	e = s.Call("GetNextID", num3)
	if e != nil {
		fmt.Println(e)
	} else {
		fmt.Println(*num3)
	}

	s.Send("Set", 80)

	num4 := new(int)
	e = s.Call("GetNextID", num4)
	if e != nil {
		fmt.Println(e)
	} else {
		fmt.Println(*num4)
	}

	num5 := new(int)
	num6 := new(int)
	call1 := s.Go("GetNextID", num5)
	call2 := s.Go("GetNextID", num6)

	e = call2.Done()
	if e != nil {
		fmt.Println(e)
	} else {
		fmt.Println(*num6)
	}

	e = call1.Done()
	if e != nil {
		fmt.Println(e)
	} else {
		fmt.Println(*num5)
	}
	//Output:50
	//50
	//2
	//81
	//83
	//82
}
