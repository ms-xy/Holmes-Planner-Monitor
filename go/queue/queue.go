package queue

import (
	// "fmt"
	"github.com/ms-xy/Holmes-Planner-Monitor/go/message"
	"sync"
)

// Create a new queue. The resulting queue is properly initialized.
// Only use this method to create a new queue, otherwise the internal locks
// are not initialized. (Dequeue would result in a nil pointer access and
// subsequent panic if the parameter blocking is true)
func New() *Queue {
	q := &Queue{}
	q.wg = sync.NewCond(q)
	// q.wg: &sync.WaitGroup{},
	return q
}

// Do not instantiate a Queue directly.
// Use New() instead.
type Queue struct {
	// resource lock
	sync.Mutex
	// simple not-ready block (for dequeue)
	wg *sync.Cond

	// regular queue fields
	head   *Node
	tail   *Node
	length uint64
}

// This data structure is the content carrier of a queue.
type Node struct {
	prev *Node
	next *Node
	data message.Message
}

// Append one element at the end of the queue.
// This operation will signal dequeue listeners that new data is available.
// However, only one listener may be awoken by this function.
// Which listener gets awoken is not predictable.
func (q *Queue) Enqueue(msg message.Message) {
	q.Lock()
	defer q.Unlock()

	// create new node for the data and enqueue it depending on queue state
	n := &Node{data: msg}
	if q.length == 0 {
		q.head, q.tail = n, n
	} else {
		q.tail.next = n
		n.prev = q.tail
		q.tail = n
	}
	q.length++

	// notify listeners that we got new data
	q.wg.Signal()
}

// Retrieve one element from the head of the queue.
// This operation blocks if the queue is empty and the parameter blocking is
// true.
func (q *Queue) Dequeue(blocking bool) message.Message {
	q.Lock()
	defer q.Unlock()

	// if blocking is true and nothing is in the queue, this function shall block
	// until the queue isn't empty anymore
	if q.length == 0 {
		if blocking {
			q.wg.Wait()
		} else {
			return nil
		}
	}

	// remove from the queue if possible, returns nil if no element available, the
	// removed first element in the queue otherwises
	var msg message.Message = nil
	if q.length > 0 {
		n := q.head
		if q.length == 1 {
			q.head, q.tail = nil, nil
		} else {
			q.head = q.head.next
			q.head.prev = nil
		}
		q.length--
		msg = n.data
	}
	return msg
}
