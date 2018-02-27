package pool

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// Author : kricen : github.com/kricen/gpool

// GPool : A GROUTINE POOL BASE ON TOKEN BUCKET ALGORITHM

// maxSize : max groutien size (default 10000)
// poolChan : in order to acquire token
// jobFunc : defined by user to complete the requirement themselves!

var (
	wg         = sync.RWMutex{}
	ErrMaxSize = errors.New("init failed,maxsize is not correct")
	ErrJobQuit = errors.New("exec failed,pool had been realsed")
)

type GPool struct {
	maxSize       int64
	poolChan      chan int
	occupiedCount int64
	jobFunc       func(interface{}) interface{}
	quit          bool
	initCount     int64
}

// New : define a function to init the GPool pool
func New(maxSize int64, jobFunc func(interface{}) interface{}) (*GPool, error) {
	if maxSize <= 0 {
		return nil, ErrMaxSize
	}
	// init the channel
	poolChan := make(chan int, maxSize)
	pool := &GPool{maxSize: maxSize, poolChan: poolChan, jobFunc: jobFunc}
	// start the pool
	pool.start()

	return pool, nil
}

//AccessOccupidCount : get the occupiedCount safety
func (g *GPool) AccessOccupidCount() int64 {
	wg.Lock()
	defer wg.Unlock()
	return g.occupiedCount
}

// change the value of occupiedCount,plus or minus
// param is a enum value(-1,1)
func (g *GPool) updateOccupidCount(num int64) {
	wg.Lock()
	defer wg.Unlock()
	atomic.AddInt64(&g.occupiedCount, num)
}

// PushJob : push a job to pool
func (g *GPool) PushJob(param interface{}) (interface{}, error) {

	if g.quit {
		return nil, ErrJobQuit
	}
	// consume token
	select {
	case <-g.poolChan:
		g.updateOccupidCount(-1)
	}

	return g.jobFunc(param), nil
}

func (g *GPool) start() {
	go func() {
		for {
			if !g.quit {
				g.produceToken()
				time.Sleep(1 * time.Millisecond)
				continue
			}
			return
		}

	}()

}

// produceToken
func (g *GPool) produceToken() {
	ot := g.AccessOccupidCount()
	if ot < g.maxSize {
		g.initCount++
		g.updateOccupidCount(1)
		g.poolChan <- 1
	}
}

// Close : define a exposed function to release the gpool resource
func (g *GPool) Close() {
	g.quit = true
	close(g.poolChan)
}
