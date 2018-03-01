package gpool

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	// define jobFunc
	f := func(param interface{}) interface{} {
		time.Sleep(1 * time.Millisecond)
		fmt.Println(param)
		return param
	}

	// acquire a pool
	gp, err := New(100, f)
	//gp, err := NewWithQps(100,1000,f)
	if err != nil {
		t.Log(err.Error())
		return
	}
	// you can set qps with this function or when gp init
	gp.SetQps(1000)

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			// push the job to pool
			gp.PushJob(index)
			//gp.PushJobWithTimeout(index, time.Second)
		}(i)
	}

	wg.Wait()
	// realse pool when not to use it
	gp.Close()

}

func TestHello(t *testing.T) {
	fmt.Printf("%20s%20s\n", "ddddddd", "ddddddsssd")
	fmt.Printf("%20s%20s\n", "dddddddsssssss", "ddddddsssd")
	for i := 0; i < 1000; i++ {
		fmt.Printf("\r%d==================================================%d", i-1, i)
		time.Sleep(1 * time.Second)

	}
}
