package riak

import "sync"

type queue struct {
	queueSize uint16
	queueChan chan interface{}
	sync.RWMutex
}

func newQueue(queueSize uint16) *queue {
	if queueSize == 0 {
		panic("[queue] size must be greater than zero!")
	}
	return &queue{
		queueSize: queueSize,
		queueChan: make(chan interface{}, queueSize),
	}
}

func (q *queue) enqueue(v interface{}) error {
	if v == nil {
		panic("attempt to enqueue nil value")
	}
	q.Lock()
	defer q.Unlock()
	return q._do_enqueue(v)
}

func (q *queue) _do_enqueue(v interface{}) error {
	if len(q.queueChan) == int(q.queueSize) {
		return newClientError("attempt to enqueue when queue is full", nil)
	}
	q.queueChan <- v
	return nil
}

func (q *queue) dequeue() (interface{}, error) {
	q.Lock()
	defer q.Unlock()
	return q._do_dequeue()
}

func (q *queue) _do_dequeue() (interface{}, error) {
	select {
	case v, ok := <-q.queueChan:
		if !ok {
			return nil, newClientError("attempt to dequeue from closed queue", nil)
		}
		return v, nil
	default:
		return nil, nil
	}
}

func (q *queue) iterate(f func(interface{}) (bool, bool)) error {
	q.Lock()
	defer q.Unlock()
	count := uint16(len(q.queueChan))
	if count == 0 {
		return nil
	}
	c := uint16(0)
	for {
		c++
		v, err := q._do_dequeue()
		if err != nil {
			return err
		}
		// NB: v may be nil if queue is currently empty
		brk, re_queue := f(v)
		if re_queue && v != nil {
			err = q._do_enqueue(v)
			if err != nil {
				return err
			}
		}
		if brk {
			break
		}
		if c == count {
			break
		}
	}
	return nil
}

func (q *queue) isEmpty() bool {
	q.RLock()
	defer q.RUnlock()
	return len(q.queueChan) == 0
}

func (q *queue) count() uint16 {
	q.RLock()
	defer q.RUnlock()
	return uint16(len(q.queueChan))
}

func (q *queue) destroy() {
	close(q.queueChan)
}
