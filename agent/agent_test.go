package agent_test

import (
	"testing"

	. "github.com/vladopajic/go-agent/agent"

	"github.com/stretchr/testify/assert"
)

func Test_NewAgent(t *testing.T) {
	t.Parallel()

	const count = 20

	w := &worker{doWorkC: make(chan chan int, count)}
	a := NewWithWorker(w)

	a.Start()
	defer a.Stop()

	for i := 0; i < count; i++ {
		p := make(chan int)
		w.doWorkC <- p
		assert.Equal(t, i, <-p)
	}
}

func Test_NewAgent_StartStop(t *testing.T) {
	t.Parallel()

	const count = 20

	w := &worker{doWorkC: make(chan chan int, count)}
	a := NewWithWorker(w)

	for i := 0; i < count; i++ {
		a.Start()

		p := make(chan int)
		w.doWorkC <- p
		assert.Equal(t, i, <-p)

		a.Stop()
	}
}

func Test_NewAgent_StopAfterNoWork(t *testing.T) {
	t.Parallel()

	const count = 20

	w := &worker{doWorkC: make(chan chan int, count)}
	a := NewWithWorker(w)

	a.Start()
	defer a.Stop()

	for i := 0; i < count; i++ {
		p := make(chan int)
		w.doWorkC <- p
		assert.Equal(t, i, <-p)
	}

	go close(w.doWorkC)
}

func Test_NewAgent_OptOnStartStop(t *testing.T) {
	t.Parallel()

	onStartC := make(chan struct{}, 1)
	onStopC := make(chan struct{}, 1)

	w := &worker{doWorkC: make(chan chan int, 1)}
	a := NewWithWorker(w,
		OptOnStart(func() {
			onStartC <- struct{}{}
		}),
		OptOnStop(func() {
			onStopC <- struct{}{}
		}),
	)

	a.Start()
	<-onStartC

	a.Stop()
	<-onStopC
}

type worker struct {
	workIteration int
	doWorkC       chan chan int
}

func (w *worker) DoWork(c Context) bool {
	select {
	case p, ok := <-w.doWorkC:
		if !ok {
			return true
		}

		p <- w.workIteration
		w.workIteration++

		return false

	case <-c.EndWorkC():
		return true
	}
}

func Test_Context(t *testing.T) {
	t.Parallel()

	ctx := NewContext()

	assert.NotNil(t, ctx.EndWorkC())
	assert.Len(t, ctx.EndWorkC(), 0)
	ctx.SignalEnd()
	assert.Len(t, ctx.EndWorkC(), 1)
}
