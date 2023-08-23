package factory

/*
缺点：
1、需要为每个水果单独实现一个工厂类,代码冗余度较高
2、原本构造多个水果类时存在的公共切面不复存在，一些通用的逻辑需要在每个水果工厂实现类中重复声明一遍
*/

/*
抽象方法层
*/

// FruitFactory 水果工厂接口
type FruitFactory interface {
	CreateFruit() Fruit
}

// Fruit 水果接口
type Fruit interface {
	Eat()
}

/*
具体实现层
*/

// OrangeFactory 橘子工厂
type OrangeFactory struct {
}

func (o *OrangeFactory) CreateFruit() Fruit {
	return NewOrange("")
}

func NewOrangeFactory() FruitFactory {
	return &OrangeFactory{}
}

// StrawberryFactory 草莓工厂
type StrawberryFactory struct {
}

func (s *StrawberryFactory) CreateFruit() Fruit {
	return NewStrawberry("")
}

func NewStrawberryFactory() FruitFactory {
	return &StrawberryFactory{}
}

// CherryFactory 樱桃工厂
type CherryFactory struct {
}

func (c *CherryFactory) CreateFruit() Fruit {
	return NewCherry("")
}

func NewCherryFactory() FruitFactory {
	return &CherryFactory{}
}
