package cars

import "sync"

type Camaro struct {
	Name          string
	Features      map[string]int
	Callback      func(string, int) (int64, error)
	Fuel          interface{}
	ChNoPos       chan string
	ChRecv        <-chan int32
	ChSend        chan<- int32
	Struct        struct{ XXX int }
	One, Two      string
	Ellipsis      func(x ...string)
	ExampleMutex  func(sync.Mutex)
	Three, Four   sync.Mutex
	AnotherStruct struct{ sync.Mutex }
	// Any           any

	sync.Mutex
}
