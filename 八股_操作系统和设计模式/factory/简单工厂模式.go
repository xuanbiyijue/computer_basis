package factory

import (
	"fmt"
)

/*
抽象层
*/

// Fruit 水果接口
type Fruit interface {
	Eat()
}

/*
具体实现层
*/

// Orange 橘子具体实现
type Orange struct {
	name string
}

func NewOrange(name string) Fruit {
	return &Orange{
		name: name,
	}
}

func (o *Orange) Eat() {
	fmt.Printf("i am orange: %s, i am about to be eaten...\n", o.name)
}

// Strawberry 草莓具体实现
type Strawberry struct {
	name string
}

func NewStrawberry(name string) Fruit {
	return &Strawberry{
		name: name,
	}
}

func (s *Strawberry) Eat() {
	fmt.Printf("i am strawberry: %s, i am about to be eaten...\n", s.name)
}

// Cherry 樱桃具体实现
type Cherry struct {
	name string
}

func NewCherry(name string) Fruit {
	return &Cherry{
		name: name,
	}
}

func (c *Cherry) Eat() {
	fmt.Printf("i am cherry: %s, i am about to be eaten...\n", c.name)
}

/*
工厂层
*/

// FruitFactory 水果工厂
type FruitFactory struct {
}

func NewFruitFactory() *FruitFactory {
	return &FruitFactory{}
}

func (f *FruitFactory) CreateFruit(typ string) (Fruit, error) {
	name := "水果1号"
	switch typ {
	case "orange":
		return NewOrange(name), nil
	case "strawberry":
		return NewStrawberry(name), nil
	case "cherry":
		return NewCherry(name), nil
	default:
		return nil, fmt.Errorf("fruit typ: %s is not supported yet\n", typ)
	}
}
