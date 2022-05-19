package agent

type Option func(o Options) Options

func OptOnStart(f func()) Option {
	return func(o Options) Options {
		o.OnStartFunc = f
		return o
	}
}

func OptOnStop(f func()) Option {
	return func(o Options) Options {
		o.OnStopFunc = f
		return o
	}
}

func OptID(id string) Option {
	return func(o Options) Options {
		o.ID = id
		return o
	}
}

func newOptions(opts []Option) Options {
	options := newZeroOptions()

	for _, opt := range opts {
		options = opt(options)
	}

	return options
}

func newZeroOptions() Options {
	return Options{
		OnStartFunc: func() {},
		OnStopFunc:  func() {},
	}
}

type Options struct {
	ID          string
	OnStartFunc func()
	OnStopFunc  func()
}
