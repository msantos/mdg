package fdpair

import "os"

type Opt struct {
	*os.File

	OpenOutput  func(r *os.File) (*os.File, error)
	CloseOutput func(r, w *os.File) error
}

type Option func(*Opt)

func New(opt ...Option) *Opt {
	o := &Opt{
		File:        os.Stdin,
		OpenOutput:  func(_ *os.File) (*os.File, error) { return os.Stdout, nil },
		CloseOutput: func(_, _ *os.File) error { return nil },
	}

	for _, fn := range opt {
		fn(o)
	}

	return o
}
