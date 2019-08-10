package server

import (
	"container/list"
)

type writeQueue struct {
	wr       Writer
	q        *list.List
	addChan  chan interface{}
	stopChan chan bool
}

func newWriteQueue(wr Writer) *writeQueue {
	q := &writeQueue{
		wr:       wr,
		q:        list.New(),
		addChan:  make(chan interface{}),
		stopChan: make(chan bool),
	}
	go q.monitor()
	return q
}

func (q *writeQueue) flush() {
	q.stopChan <- true
	<-q.stopChan
}

func (q *writeQueue) add(x interface{}) {
	q.addChan <- x
}

func (q *writeQueue) monitor() {
	writing := false
	wrChan := make(chan error)
	stopped := false

	write := func() {
		if writing {
			return
		}
		if q.q.Len() == 0 {
			return
		}
		writing = true
		e := q.q.Front()
		q.q.Remove(e)
		go func() {
			err := q.wr.WritePacket(e.Value)
			wrChan <- err
		}()
	}

	for {
		select {
		case x := <-q.addChan:
			q.q.PushBack(x)
			write()
		case <-wrChan:
			writing = false
			if stopped {
				q.stopChan <- false
				return
			}
			write()
		case <-q.stopChan:
			stopped = true
			if !writing {
				q.stopChan <- false
				return
			}
		}
	}
}
