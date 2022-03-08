package server

import (
	"github.com/prometheus/common/log"
	"github.com/spf13/viper"
)

type TaskFunc func()

type Pool struct {
	taskQueue  chan TaskFunc
	stopChan   chan struct{}
	workerSize int
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
	p := &Pool{taskQueue: make(chan TaskFunc, queueSize), stopChan: make(chan struct{}), workerSize: workerSize}
	for i := 0; i < p.workerSize; i++ {
		go p.run()
	}
	return p
}

func (p *Pool) Commit(taskFunc TaskFunc) {
	log.Infof("receive a task")
	log.Infof("current worker size: %d", p.workerSize)
	log.Infof("task queue size: %d", p.Len())
	p.taskQueue <- taskFunc
}

func (p *Pool) Len() int {
	return len(p.taskQueue)
}

func (p *Pool) IsEmpty() bool {
	return len(p.taskQueue) == 0
}

func (p *Pool) run() {
	for {
		select {
		case task := <-p.taskQueue:
			task()
		case <-p.stopChan:
			break
		}
	}
}
