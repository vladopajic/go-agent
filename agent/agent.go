package agent

import (
	"reflect"

	log "github.com/sirupsen/logrus"
)

type Factory func() Agent

type Agent interface {
	Start()
	Stop()
}

type Context interface {
	EndWorkC() <-chan struct{}
	SignalEnd()
}

type contextImpl struct {
	endWork chan struct{}
}

func NewContext() Context {
	return &contextImpl{
		endWork: make(chan struct{}, 1),
	}
}

func (c *contextImpl) EndWorkC() <-chan struct{} {
	return c.endWork
}

func (c *contextImpl) SignalEnd() {
	c.endWork <- struct{}{}
}

type Worker interface {
	DoWork(c Context) (workEnded bool)
}

func NewWithWorker(w Worker, opt ...Option) Agent {
	return &agentImpl{
		options: newOptions(opt),
		ctx:     NewContext(),
		worker:  w,
	}
}

func StartNewWithWorker(w Worker) Agent {
	a := NewWithWorker(w)
	a.Start()

	return a
}

type agentImpl struct {
	options       Options
	ctx           Context
	worker        Worker
	workEndedSigC chan struct{}
	workerRunning bool
}

// Stop ends executions of underlaying `WorkerFunc`.
func (a *agentImpl) Stop() {
	if !a.workerRunning {
		return
	}

	a.workEndedSigC = make(chan struct{})
	a.ctx.SignalEnd()
	<-a.workEndedSigC
}

// Start begins executions of underlaying `WorkerFunc`.
func (a *agentImpl) Start() {
	if a.workerRunning {
		return
	}

	a.workerRunning = true

	go a.doWork()
}

// doWork executes `Worker` of this `Agent` until
// `Agent` or `Worker` has signaled to stop.
func (a *agentImpl) doWork() {
	log.WithFields(log.Fields{
		"id":     a.options.ID,
		"worker": workerName(a.worker),
	}).Debug("starting agent")

	a.options.OnStartFunc()
	defer a.options.OnStopFunc()

	for workEnded := false; !workEnded; {
		workEnded = a.worker.DoWork(a.ctx)
	}

	a.workerRunning = false
	if c := a.workEndedSigC; c != nil {
		c <- struct{}{}
	}

	log.WithFields(log.Fields{
		"id":     a.options.ID,
		"worker": workerName(a.worker),
	}).Debug("stopping agent")
}

func StartAll(agents ...Agent) {
	for _, a := range agents {
		a.Start()
	}
}

func StopAll(agents ...Agent) {
	for _, a := range agents {
		a.Stop()
	}
}

func Combine(agents ...Agent) Agent {
	return &combinedAgentImpl{agents}
}

type combinedAgentImpl struct {
	agents []Agent
}

func (a *combinedAgentImpl) Stop() {
	StopAll(a.agents...)
}

func (a *combinedAgentImpl) Start() {
	StartAll(a.agents...)
}

func workerName(w Worker) string {
	val := reflect.Indirect(reflect.ValueOf(w))
	return val.Type().Name()
}
