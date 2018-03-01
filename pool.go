package gpool

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// Author : kricen : github.com/kricen/gpool

var (
	wg            = sync.RWMutex{}
	ErrMaxSize    = errors.New("init failed,maxsize is not correct")
	ErrJobQuit    = errors.New("exec failed,pool had been realsed")
	ErrJobTimeout = errors.New("timeout")
)

const (
	DefaultQps = 800
)

// GPool : A GROUTINE POOL BASE ON TOKEN BUCKET ALGORITHM
type GPool struct {
	maxSize       int64                         // max goroutine size in channel
	poolChan      chan int                      // a pool to store token
	occupiedCount int64                         // goroutine size in pool
	jobFunc       func(interface{}) interface{} // user's handle function
	quit          bool                          // quit single
	produceSpeed  time.Duration                 // produce the speed of token
}

// New : define a function to init the GPool pool
func New(maxSize int64, jobFunc func(interface{}) interface{}) (*GPool, error) {
	return NewWithQps(maxSize, DefaultQps, jobFunc)
}

// New : define a function to init the GPool pool with param qps
// qps : query peer speed
func NewWithQps(maxSize int64, qps int64, jobFunc func(interface{}) interface{}) (*GPool, error) {
	if maxSize <= 0 {
		return nil, ErrMaxSize
	}
	// init the channel
	poolChan := make(chan int, maxSize)

	pool := &GPool{maxSize: maxSize, poolChan: poolChan, jobFunc: jobFunc}
	// set qps
	pool.SetQps(qps)
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

// PushJobWithTimeout : push a job to pool with a timeout mechanism
func (g *GPool) PushJobWithTimeout(param interface{}, timeout time.Duration) (interface{}, error) {

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	if g.quit {
		return nil, ErrJobQuit
	}
	// consume token
	select {
	case <-g.poolChan:
		g.updateOccupidCount(-1)
	case <-timer.C:
		return nil, ErrJobTimeout
	}

	return g.jobFunc(param), nil
}

// SetQps : set qps optional
func (g *GPool) SetQps(qps int64) {
	// calculate the produceSpeed
	speed := int(float64(1) / float64(qps) * 1 * 1000 * 1000)
	if speed <= 0 {
		speed = DefaultQps
	}
	g.produceSpeed = time.Duration(speed) * time.Microsecond
}

func (g *GPool) start() {
	ticker := time.NewTicker(g.produceSpeed)
	go func() {
		defer ticker.Stop()
		for _ = range ticker.C {
			if !g.quit {
				g.produceToken()
				continue
			}
			return
		}

	}()

}

// produceToken
func (g *GPool) produceToken() {

	// produce may write to a closed channel, so
	// need to recover when panic
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	ot := g.AccessOccupidCount()
	if ot < g.maxSize {
		g.updateOccupidCount(1)
		g.poolChan <- 1
	}
}

// Close : define a exposed function to release the gpool resource
func (g *GPool) Close() {
	g.quit = true
	close(g.poolChan)
}
