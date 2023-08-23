package singleton

// Instance 单例模式应该实现的接口，通过此接口，可以避免包内私有变量被导出
type Instance interface {
	Work()
}

var s *singleton

type singleton struct{}

func (s *singleton) Work() {}

func newSingleton() *singleton {
	return &singleton{}
}

func init() {
	s = newSingleton()
}

func GetInstance() Instance {
	return s
}
