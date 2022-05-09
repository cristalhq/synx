package synx

type ErrCh chan error

func NewErrCh() ErrCh { return make(chan error, 1) }

func (e ErrCh) Get() error { return <-e }

func (e ErrCh) Set(err error) { e <- err }
