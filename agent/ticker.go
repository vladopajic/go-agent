package agent

import "time"

type Ticker interface {
	Agent
	C() <-chan time.Time
}

func NewTicker(tickInterval time.Duration) Ticker {
	return &tickerImpl{
		tickInterval: tickInterval,
		out:          make(chan time.Time, 1),
	}
}

type tickerImpl struct {
	tickInterval time.Duration
	out          chan time.Time
	ticker       *time.Ticker
}

func (a *tickerImpl) Start() {
	a.ticker = time.NewTicker(a.tickInterval)
	go a.doWork()
}

func (a *tickerImpl) doWork() {
	for {
		t, ok := <-a.ticker.C
		if !ok {
			close(a.out)
			return
		}
		a.out <- t
	}
}

func (a *tickerImpl) Stop() {
	a.ticker.Stop()
	close(a.out)
}

func (a *tickerImpl) C() <-chan time.Time {
	return a.out
}
