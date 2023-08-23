package Decorator

type Food interface {
	// 食用主食
	Eat() string
	// 计算主食的话费
	Cost() float32
}

type Rice struct {
}

func NewRice() Food {
	return &Rice{}
}

func (r *Rice) Eat() string {
	return "开动了，一碗香喷喷的米饭..."
}

// 需要花费的金额
func (r *Rice) Cost() float32 {
	return 1
}

type Noodle struct {
}

func NewNoodle() Food {
	return &Noodle{}
}

func (n *Noodle) Eat() string {
	return "嗦面ing...嗦~"
}

// 需要花费的金额
func (n *Noodle) Cost() float32 {
	return 1.5
}

type Decorator Food

func NewDecorator(f Food) Decorator {
	return f
}

type LaoGanMaDecorator struct {
	Decorator
}

func NewLaoGanMaDecorator(d Decorator) Decorator {
	return &LaoGanMaDecorator{
		Decorator: d,
	}
}

func (l *LaoGanMaDecorator) Eat() string {
	// 加入老干妈配料
	return "加入一份老干妈~..." + l.Decorator.Eat()
}

func (l *LaoGanMaDecorator) Cost() float32 {
	// 价格增加 0.5 元
	return 0.5 + l.Decorator.Cost()
}

type HamSausageDecorator struct {
	Decorator
}

func NewHamSausageDecorator(d Decorator) Decorator {
	return &HamSausageDecorator{
		Decorator: d,
	}
}

func (h *HamSausageDecorator) Eat() string {
	// 加入火腿肠配料
	return "加入一份火腿~..." + h.Decorator.Eat()
}

func (h *HamSausageDecorator) Cost() float32 {
	// 价格增加 1.5 元
	return 1.5 + h.Decorator.Cost()
}

type FriedEggDecorator struct {
	Decorator
}

func NewFriedEggDecorator(d Decorator) Decorator {
	return &FriedEggDecorator{
		Decorator: d,
	}
}

func (f *FriedEggDecorator) Eat() string {
	// 加入煎蛋配料
	return "加入一份煎蛋~..." + f.Decorator.Eat()
}

func (f *FriedEggDecorator) Cost() float32 {
	// 价格增加 1 元
	return 1 + f.Decorator.Cost()
}
