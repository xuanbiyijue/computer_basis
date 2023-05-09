https://www.zhihu.com/tardis/zm/art/110085652?source_id=1005
# 1. 什么是 context？为什么需要 context？
在并发程序中，由于超时、取消操作或者一些异常情况，往往需要进行抢占操作或者中断后续操作。  
考虑下面这种情况：  
* 假如主协程中有多个任务1, 2, …m，主协程对这些任务有超时控制；
* 而其中任务1又有多个子任务1, 2, …n，
* 任务1对这些子任务也有自己的超时控制，那么这些子任务既要感知主协程的取消信号，也需要感知任务1的取消信号。

如果使用done channel的用法，我们需要定义两个done channel，子任务们需要同时监听这两个done channel。但是如果层级更深，如果这些子任务还有子任务，那么使用done channel的方式将会变得非常繁琐且混乱。  
因此，需要使用 context：  
* 上层任务取消后，所有的下层任务都会被取消；
* 中间某一层的任务取消后，只会将当前任务的下层任务取消，而不会影响上层的任务以及同级任务。
> Context是Go并发编程中常用到一种编程模式。它可以用来在多个goroutine之间传递请求作用域的变量、取消信号以及请求的截止时间等信息。

> * 上游任务仅仅使用context通知下游任务不再需要，但不会直接干涉和中断下游任务的执行，由下游任务自行决定后续的处理操作，也就是说context的取消操作是无侵入的；
> * context是线程安全的，因为context本身是不可变的（immutable），因此可以放心地在多个协程中传递使用。

> 其主要的应用 ：1：上下文控制，2：多个 goroutine 之间的数据交互等，3：超时控制：到某个时间点超时，过多久超时。

# 2. Context 接口结构
```go
type Context interface {
    // 返回绑定当前context的任务被取消的截止时间；如果没有设定期限，将返回ok == false。
    Deadline() (deadline time.Time, ok bool) 
    // 当绑定当前context的任务被取消时，将返回一个关闭的channel；如果当前context不会被取消，将返回nil。
    Done() <-chan struct{}
    // 如果Done返回的channel没有关闭，将返回nil;如果Done返回的channel已经关闭，将返回非空的值表示任务结束的原因。如果是context被取消，Err将返回Canceled；如果是context超时，Err将返回DeadlineExceeded。
    Err() error
    // 返回context存储的键值对中当前key对应的值，如果没有对应的key,则返回nil。
    Value(key interface{}) interface{}
}
```

# 3. emptyCtx
emptyCtx是一个int类型的变量，但实现了context的接口。emptyCtx没有超时时间，不能取消，也不能存储任何额外信息，所以emptyCtx用来作为context树的根节点。
```go
// An emptyCtx is never canceled, has no values, and has no deadline. It is not
// struct{}, since vars of this type must have distinct addresses.
type emptyCtx int

func (*emptyCtx) Deadline() (deadline time.Time, ok bool) {
    return
}

func (*emptyCtx) Done() <-chan struct{} {
    return nil
}

func (*emptyCtx) Err() error {
    return nil
}

func (*emptyCtx) Value(key interface{}) interface{} {
    return nil
}

func (e *emptyCtx) String() string {
    switch e {
    case background:
        return "context.Background"
    case todo:
        return "context.TODO"
    }
    return "unknown empty Context"
}

var (
    background = new(emptyCtx)
    todo       = new(emptyCtx)
)

func Background() Context {
    return background
}

func TODO() Context {
    return todo
}
```
* 一般不会直接使用emptyCtx，而是使用由emptyCtx实例化的两个变量，分别可以通过调用Background和TODO方法得到
* Background和TODO方法得到的context区别：
  Background和TODO只是用于不同场景下： 
  * Background通常被用于主函数、初始化以及测试中，作为一个顶层的context，也就是说一般我们创建的context都是基于Background；
  * 而TODO是在不确定使用什么context的时候才会使用。


# 4. 基础context类型 1/2：valueCtx
```go
type valueCtx struct {
    Context
    key, val interface{}
}

func (c *valueCtx) Value(key interface{}) interface{} {
    if c.key == key {
        return c.val
    }
    return c.Context.Value(key)
}
```
* valueCtx 嵌入了一个 Context, 表示父节点context
* valueCtx类型还携带一组键值对，也就是说这种context可以携带额外的信息。
* valueCtx实现了Value方法，用以在context链路上获取key对应的值，如果当前context上不存在需要的key,会沿着context链向上寻找key对应的值，直到根节点。
```go
func WithValue(parent Context, key, val interface{}) Context {
    if key == nil {
        panic("nil key")
    }
    if !reflect.TypeOf(key).Comparable() {
        panic("key is not comparable")
    }
    return &valueCtx{parent, key, val}
}
```
* WithValue函数  
WithValue用以向context添加键值对。这里添加键值对不是在原context结构体上直接添加，而是以此context作为父节点，重新创建一个新的valueCtx子节点，将键值对添加在子节点上，由此形成一条context链。获取value的过程就是在这条context链上由尾部上前搜寻：
![img](https://pic4.zhimg.com/v2-6e74cf6f7a4f1701d262c4c0939df52f_b.webp?consumer=ZHI_MENG)


# 5. 基础context类型 2/2：cancelCtx(可取消的ctx)
```go
type cancelCtx struct {
    Context

    mu       sync.Mutex            // protects following fields
    done     chan struct{}         // created lazily, closed by first cancel call
    children map[canceler]struct{} // set to nil by the first cancel call
    err      error                 // set to non-nil by the first cancel call
}

type canceler interface {
    cancel(removeFromParent bool, err error)
    Done() <-chan struct{}
}
```
* 跟valueCtx类似，cancelCtx中也有一个context变量作为父节点；
* 变量done表示一个channel，用来表示传递关闭信号；
* children表示一个map，存储了当前context节点下的子节点；
* err用于存储错误信息表示任务结束的原因。
```go
func (c *cancelCtx) Done() <-chan struct{} {
    c.mu.Lock()
    if c.done == nil {
        c.done = make(chan struct{})
    }
    d := c.done
    c.mu.Unlock()
    return d
}

func (c *cancelCtx) Err() error {
    c.mu.Lock()
    err := c.err
    c.mu.Unlock()
    return err
}

func (c *cancelCtx) cancel(removeFromParent bool, err error) {
    if err == nil {
        panic("context: internal error: missing cancel error")
    }
    c.mu.Lock()
    if c.err != nil {
        c.mu.Unlock()
        return // already canceled
    }
    // 设置取消原因
    c.err = err
    设置一个关闭的channel或者将done channel关闭，用以发送关闭信号
    if c.done == nil {
        c.done = closedchan
    } else {
        close(c.done)
    }
    // 将子节点context依次取消
    for child := range c.children {
        // NOTE: acquiring the child's lock while holding parent's lock.
        child.cancel(false, err)
    }
    c.children = nil
    c.mu.Unlock()

    if removeFromParent {
        // 将当前context节点从父节点上移除
        removeChild(c.Context, c)
    }
}
```
* WithCancel 函数
WithCancel函数用来创建一个可取消的context，即cancelCtx类型的context。WithCancel返回一个context和一个CancelFunc，调用CancelFunc即可触发cancel操作。直接看源码：
```go
type CancelFunc func()

func WithCancel(parent Context) (ctx Context, cancel CancelFunc) {
    c := newCancelCtx(parent)
    propagateCancel(parent, &c)
    return &c, func() { c.cancel(true, Canceled) }
}

// newCancelCtx returns an initialized cancelCtx.
func newCancelCtx(parent Context) cancelCtx {
    // 将parent作为父节点context生成一个新的子节点
    return cancelCtx{Context: parent}
}

func propagateCancel(parent Context, child canceler) {
    if parent.Done() == nil {
        // parent.Done()返回nil表明父节点以上的路径上没有可取消的context
        return // parent is never canceled
    }
    // 获取最近的类型为cancelCtx的祖先节点
    if p, ok := parentCancelCtx(parent); ok {
        p.mu.Lock()
        if p.err != nil {
            // parent has already been canceled
            child.cancel(false, p.err)
        } else {
            if p.children == nil {
                p.children = make(map[canceler]struct{})
            }
            // 将当前子节点加入最近cancelCtx祖先节点的children中
            p.children[child] = struct{}{}
        }
        p.mu.Unlock()
    } else {
        go func() {
            select {
            case <-parent.Done():
                child.cancel(false, parent.Err())
            case <-child.Done():
            }
        }()
    }
}

func parentCancelCtx(parent Context) (*cancelCtx, bool) {
    for {
        switch c := parent.(type) {
        case *cancelCtx:
            return c, true
        case *timerCtx:
            return &c.cancelCtx, true
        case *valueCtx:
            parent = c.Context
        default:
            return nil, false
        }
    }
}
```
cancelCtx取消时，会将后代节点中所有的cancelCtx都取消，propagateCancel即用来建立当前节点与祖先节点这个取消关联逻辑: 
  * 如果parent.Done()返回nil，表明父节点以上的路径上没有可取消的context，不需要处理；
  * 如果在context链上找到到cancelCtx类型的祖先节点，则判断这个祖先节点是否已经取消，如果已经取消就取消当前节点；否则将当前节点加入到祖先节点的children列表。
  * 否则开启一个协程，监听parent.Done()和child.Done()，一旦parent.Done()返回的channel关闭，即context链中某个祖先节点context被取消，则将当前context也取消。

```go
type timerCtx struct {
    cancelCtx
    timer *time.Timer // Under cancelCtx.mu.

    deadline time.Time
}

func (c *timerCtx) Deadline() (deadline time.Time, ok bool) {
    return c.deadline, true
}

func (c *timerCtx) cancel(removeFromParent bool, err error) {
    将内部的cancelCtx取消
    c.cancelCtx.cancel(false, err)
    if removeFromParent {
        // Remove this timerCtx from its parent cancelCtx's children.
        removeChild(c.cancelCtx.Context, c)
    }
    c.mu.Lock()
    if c.timer != nil {
        取消计时器
        c.timer.Stop()
        c.timer = nil
    }
    c.mu.Unlock()
}
```
* timerCtx: 可以定时取消的context  
timerCtx内部使用cancelCtx实现取消，另外使用定时器timer和过期时间deadline实现定时取消的功能。timerCtx在调用cancel方法，会先将内部的cancelCtx取消，如果需要则将自己从cancelCtx祖先节点上移除，最后取消计时器。

* WithDeadline  
WithDeadline返回一个基于parent的可取消的context，并且其过期时间deadline不晚于所设置时间d。
  * 如果父节点parent有过期时间并且过期时间早于给定时间d，那么新建的子节点context无需设置过期时间，使用WithCancel创建一个可取消的context即可；
  * 否则，就要利用parent和过期时间d创建一个定时取消的timerCtx，并建立新建context与可取消context祖先节点的取消关联关系，接下来判断当前时间距离过期时间d的时长dur：
  * 如果dur小于0，即当前已经过了过期时间，则直接取消新建的timerCtx，原因为DeadlineExceeded；
否则，为新建的timerCtx设置定时器，一旦到达过期时间即取消当前timerCtx。
```go
func WithDeadline(parent Context, d time.Time) (Context, CancelFunc) {
    if cur, ok := parent.Deadline(); ok && cur.Before(d) {
        // The current deadline is already sooner than the new one.
        return WithCancel(parent)
    }
    c := &timerCtx{
        cancelCtx: newCancelCtx(parent),
        deadline:  d,
    }
    // 建立新建context与可取消context祖先节点的取消关联关系
    propagateCancel(parent, c)
    dur := time.Until(d)
    if dur <= 0 {
        c.cancel(true, DeadlineExceeded) // deadline has already passed
        return c, func() { c.cancel(false, Canceled) }
    }
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.err == nil {
        c.timer = time.AfterFunc(dur, func() {
            c.cancel(true, DeadlineExceeded)
        })
    }
    return c, func() { c.cancel(true, Canceled) }
}
```

* WithTimeout  
与WithDeadline类似，WithTimeout也是创建一个定时取消的context，只不过WithDeadline是接收一个过期时间点，而WithTimeout接收一个相对当前时间的过期时长timeout:  
```go
func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc) {
    return WithDeadline(parent, time.Now().Add(timeout))
}
```

# 6. context 的使用
通过 context 实现更优雅实现协程间取消信号的同步：
```go
func main() {
    messages := make(chan int, 10)

    // producer
    for i := 0; i < 10; i++ {
        messages <- i
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

    // consumer
    go func(ctx context.Context) {
        ticker := time.NewTicker(1 * time.Second)
        for _ = range ticker.C {
            select {
            case <-ctx.Done():
                fmt.Println("child process interrupt...")
                return
            default:
                fmt.Printf("send message: %d\n", <-messages)
            }
        }
    }(ctx)

    defer close(messages)
    defer cancel()

    select {
    case <-ctx.Done():
        time.Sleep(1 * time.Second)
        fmt.Println("main process exit!")
    }
}
```