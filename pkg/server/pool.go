package server

import (
	"time"

	"github.com/prometheus/common/log"
	"github.com/spf13/viper"
)

type TaskFunc func()

type Pool struct {
	taskQueue     chan TaskFunc
	workerSize    int
	goroutineChan chan struct{}
}

func NewPool() *Pool {
	workerSize := viper.GetInt("app.worker")
	queueSize := viper.GetInt("app.queue")
	if queueSize < 1 {
		queueSize = 1
	}
	if workerSize < 1 {
		workerSize = 1
	}
	p := &Pool{taskQueue: make(chan TaskFunc, queueSize), workerSize: workerSize, goroutineChan: make(chan struct{}, workerSize)}
	return p
}

func (p *Pool) Commit(taskFunc TaskFunc) {
	log.Infof("receive a task")
	log.Infof("current worker size: %d", p.workerSize)
	log.Infof("task queue size: %d", p.Len())
	log.Infof("current goroutine size: %d", len(p.goroutineChan))
	p.taskQueue <- taskFunc
	go p.setUp()
}

func (p *Pool) Len() int {
	return len(p.taskQueue)
}

func (p *Pool) IsEmpty() bool {
	return len(p.taskQueue) == 0
}

func (p *Pool) setUp() {
	p.goroutineChan <- struct{}{}
	go p.run()
}

func (p *Pool) run() {
	defer func() {
		<-p.goroutineChan
	}()
	select {
	case task := <-p.taskQueue:
		task()
	case <-time.After(time.Hour * 1):
		break
	}
}
