package queue

import (
	"testing"

	"github.com/ms-xy/Holmes-Planner-Monitor/go/message"
)

func TestQueue(t *testing.T) {
	queue := New()

	msg1 := message.NewStringMessage("1")
	msg2 := message.NewStringMessage("2")
	msg3 := message.NewStringMessage("3")
	msg4 := message.NewStringMessage("4")

	queue.Enqueue(msg1)
	out1 := queue.Dequeue(true)
	if msg1 != out1 {
		t.Log("Test 1, queue. Dequeue(true) != msg1")
		t.Fail()
		return
	}

	queue.Enqueue(msg1)
	queue.Enqueue(msg2)
	out1 = queue.Dequeue(true)
	out2 := queue.Dequeue(true)
	if msg1 != out1 || msg2 != out2 {
		t.Log("Test 2, queue. Dequeue(true) != [msg1, msg2]")
		t.Fail()
		return
	}

	queue.Enqueue(msg1)
	queue.Enqueue(msg2)
	queue.Enqueue(msg3)
	out1 = queue.Dequeue(true)
	out2 = queue.Dequeue(true)
	out3 := queue.Dequeue(true)
	if msg1 != out1 || msg2 != out2 || msg3 != out3 {
		t.Log("Test 3, queue. Dequeue(true) != [msg1, msg2, msg3]")
		t.Fail()
		return
	}

	queue.Enqueue(msg1)
	queue.Enqueue(msg2)
	queue.Enqueue(msg3)
	queue.Dequeue(true)
	queue.Enqueue(msg4)
	out2 = queue.Dequeue(true)
	out3 = queue.Dequeue(true)
	out4 := queue.Dequeue(true)
	if msg2 != out2 || msg3 != out3 || msg4 != out4 {
		t.Log("Test 4, queue. Dequeue(true) != [msg2, msg3, msg4]")
		t.Fail()
		return
	}
}

// According to this benchmark we're somewhere around 600ns per op
func BenchmarkBlockingQueue(b *testing.B) {
	queue := New()
	go func() {
		for i := 0; i < b.N; i++ {
			queue.Enqueue(message.NewIntMessage(i))
		}
	}()
	for i := 0; i < b.N; i++ {
		queue.Dequeue(true)
	}
}

// According to this benchmark we're somewhere around 150ns per op
func BenchmarkChannelBuffer1000(b *testing.B) {
	queue := make(chan message.Message, 1000)
	go func() {
		for i := 0; i < b.N; i++ {
			queue <- message.NewIntMessage(i)
		}
		close(queue)
	}()
	for i := 0; i < b.N; i++ {
		_, more := <-queue
		if !more {
			break
		}
	}
}
