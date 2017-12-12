package kaniini

type inMemoryQueue struct {
	name       string
	deliveries chan *Delivery
	done       chan struct{}
}

func NewInMemoryQueue(name string, size int) Queue {
	return &inMemoryQueue{
		name:       name,
		deliveries: make(chan *Delivery, size),
		done:       make(chan struct{}),
	}
}

func (q *inMemoryQueue) Receive() <-chan *Delivery {
	return q.deliveries
}

func (q *inMemoryQueue) Done() chan struct{} {
	return q.done
}

func (q *inMemoryQueue) Send(data []byte) error {
	q.deliveries <- &Delivery{
		Body: data,
	}

	return nil
}
