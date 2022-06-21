package agent

import "time"

type Ticker interface {
	Agent
	C() <-chan time.Time
}

func NewTicker(tickInterval time.Duration) Ticker {
	w := &tickerWorker{
		tickInterval: tickInterval,
		out:          make(chan time.Time, 1),
	}

	return &tickerImpl{
		Agent:  NewWithWorker(w, OptOnStart(w.onStart), OptOnStop(w.onStop)),
		worker: w,
	}
}

type tickerImpl struct {
	Agent
	worker *tickerWorker
}

func (t *tickerImpl) C() <-chan time.Time {
	return t.worker.out
}

type tickerWorker struct {
	tickInterval time.Duration
	out          chan time.Time
	ticker       *time.Ticker
}

func (w *tickerWorker) onStart() {
	w.ticker = time.NewTicker(w.tickInterval)
}

func (w *tickerWorker) onStop() {
	w.ticker.Stop()
	close(w.out)
}

func (w *tickerWorker) DoWork(c Context) (workEnded bool) {
	select {
	case t, ok := <-w.ticker.C:
		if !ok {
			return true
		}

		w.out <- t

		return false

	case <-c.EndWorkC():
		return true
	}
}
