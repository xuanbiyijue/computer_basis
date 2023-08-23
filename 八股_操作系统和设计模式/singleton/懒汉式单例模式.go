package singleton

import "sync"

var (
	s   *singleton
	mux sync.Mutex
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
	if s != nil {
		return s
	}
	mux.Lock()
	defer mux.Unlock()
	// 这里还要再判断一次，是为了避免两个Goroutine竞争锁，前者初始化后，后者的重复初始化
	if s != nil {
		return s
	}
	s = newSingleton()

	return s
}
