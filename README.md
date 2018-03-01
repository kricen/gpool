
GPool is a simple and fast goroutine pool base on token bucket algorithm

## Install

```
go get github.com/kricen/gpool

```


## How to use

```go

// define jobFunc
f := func(param interface{}) interface{} {
	time.Sleep(1 * time.Millisecond)
	fmt.Println(param)
	return param
}

// acquire a pool: maxSize and job function defined by user
gp, err := New(100, f)
// need 3 params (maxSize int64, qps int64, jobFunc func(interface{}) interface{})
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
		//gp.PushJobWithTimeout(index, time.Second)s
	}(i)
}

wg.Wait()
// realse pool when not to use it
gp.Close()

```
