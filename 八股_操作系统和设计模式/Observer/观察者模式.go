package Observer

import (
	"context"
	"fmt"
	"sync"
)

type Event struct {
	Topic string
	Val   interface{}
}

/* https://mp.weixin.qq.com/mp/appmsgalbum?__biz=MzkxMjQzMjA0OQ==&action=getalbum&album_id=2935694957926088707&scene=173&from_msgid=2247484352&from_itemidx=1&count=3&nolastread=1#wechat_redirect
抽象层
*/

// Observer 观察者接口
type Observer interface {
	OnChange(ctx context.Context, e *Event) error
}

// EventBus 事件总线
type EventBus interface {
	Subscribe(topic string, o Observer)
	Unsubscribe(topic string, o Observer)
	Publish(ctx context.Context, e *Event)
}

/*
具体实现层
*/

// BaseObserver 基础观察者
type BaseObserver struct {
	name string
}

func (b *BaseObserver) OnChange(ctx context.Context, e *Event) error {
	fmt.Printf("observer: %s, event key: %s, event val: %v", b.name, e.Topic, e.Val)
	// ...
	return nil
}

func NewBaseObserver(name string) *BaseObserver {
	return &BaseObserver{
		name: name,
	}
}

// BaseEventBus 基础事件总线
type BaseEventBus struct {
	mux       sync.RWMutex
	observers map[string]map[Observer]struct{}
}

func (b *BaseEventBus) Subscribe(topic string, o Observer) {
	b.mux.Lock()
	defer b.mux.Unlock()
	_, ok := b.observers[topic]
	if !ok {
		b.observers[topic] = make(map[Observer]struct{})
	}
	b.observers[topic][o] = struct{}{}
}

func (b *BaseEventBus) Unsubscribe(topic string, o Observer) {
	b.mux.Lock()
	defer b.mux.Unlock()
	delete(b.observers[topic], o)
}

func NewBaseEventBus() BaseEventBus {
	return BaseEventBus{
		observers: make(map[string]map[Observer]struct{}),
	}
}

// SyncEventBus 同步模式
type SyncEventBus struct {
	BaseEventBus
}

func NewSyncEventBus() *SyncEventBus {
	return &SyncEventBus{
		BaseEventBus: NewBaseEventBus(),
	}
}

func (s *SyncEventBus) Publish(ctx context.Context, e *Event) {
	s.mux.RLock()
	subscribers := s.observers[e.Topic]
	s.mux.RUnlock()

	errs := make(map[Observer]error)
	for subscriber := range subscribers {
		if err := subscriber.OnChange(ctx, e); err != nil {
			errs[subscriber] = err
		}
	}

	s.handleErr(ctx, errs)
}

func (s *SyncEventBus) handleErr(ctx context.Context, errs map[Observer]error) {
	for o, err := range errs {
		// 处理 publish 失败的 observer
		fmt.Printf("observer: %v, err: %v", o, err)
	}
}
