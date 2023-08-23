package singleton

import "sync"

var (
	s    *singleton
	once sync.Once
)

type Instance interface {
	Work()
}

type singleton struct {
}

func (s *singleton) Work() {
}

func newSingleton() *singleton {
	return &singleton{}
}

func GetInstance() Instance {
	once.Do(func() {
		s = newSingleton()
	})
	return s
}
