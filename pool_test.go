package pool

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
	gp, err := New(100, 1000, f)
	if err != nil {
		t.Log(err.Error())
		return
	}

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
