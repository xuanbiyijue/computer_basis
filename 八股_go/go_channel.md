https://zhuanlan.zhihu.com/p/395278270
# 1. chan 的底层实现
channel在底层是一个hchan结构体，位于src/runtime/chan.go里。其定义如下:
```go
type hchan struct {
    qcount   uint   // channel 里的元素计数
    dataqsiz uint   // 可以缓冲的数量，如 ch := make(chan int, 10)。 此处的 10 即 dataqsiz
    elemsize uint16 // 要发送或接收的数据类型大小
    buf      unsafe.Pointer // 当 channel 设置了缓冲数量时，该 buf 指向一个存储缓冲数据的区域，该区域是一个循环队列的数据结构
    closed   uint32 // 关闭状态
    sendx    uint  // 当 channel 设置了缓冲数量时，数据区域即循环队列此时已发送数据的索引位置
    recvx    uint  // 当 channel 设置了缓冲数量时，数据区域即循环队列此时已接收数据的索引位置
    recvq    waitq // 想读取数据但又被阻塞住的 goroutine 队列
    sendq    waitq // 想发送数据但又被阻塞住的 goroutine 队列

    lock mutex
    ...
}
```
![img](https://pic4.zhimg.com/80/v2-9452c08ff058590cea1b40a39fd8c70f_720w.webp)

# 2. 无缓冲 chan 读写
* 先写再读
![img](https://pic2.zhimg.com/80/v2-f2738822fbf052dff0ee47cb225fc405_720w.webp)
可以看到，由于 channel 是无缓冲的，所以 G1 暂时被挂在 sendq 队列里，然后 G1 调用了 gopark 休眠了起来。  
接着，又有 goroutine 来 channel 读取数据了：
![img](https://pic3.zhimg.com/80/v2-451480d87d1d19fa72fce11779ef7742_720w.webp)
此时 G2 发现 sendq 等待队列里有 goroutine 存在，于是直接从 G1 copy 数据过来，并且会对 G1 设置 goready 函数，这样下次调度发生时， G1 就可以继续运行，并且会从等待队列里移除掉。  

* 先读再写
![img](https://pic3.zhimg.com/80/v2-b18555f9ae91578c29213fade271e97a_720w.webp)
G1 暂时被挂在了 recvq 队列，然后休眠起来。  
G2 在写数据时，发现 recvq 队列有 goroutine 存在，于是直接将数据发送给 G1。  
同时设置 G1 goready 函数，等待下次调度运行。
![img](https://pic4.zhimg.com/80/v2-d3692a423f1a347213ec28d0bb3e921b_720w.webp)

# 3. 有缓冲 chan
会优先判断缓冲数据区域是否已满，如果未满，则将数据保存在缓冲数据区域，即环形队列里。  
如果已满，则和之前的流程是一样的，会挂在等待队列。  
![img](https://pic4.zhimg.com/80/v2-53c29200c3be1f606ba80741d7979a0f_720w.webp)  

当 G2 要读取数据时，会优先从缓冲数据区域去读取，并且在读取完后，会检查 sendq 队列，如果 goroutine 有等待队列，则会将它上面的 data 补充到缓冲数据区域，并且也对其设置 goready 函数。

# 4. channel 的 deadlock
往 channel 里读写数据时是有可能被阻塞住的，一旦被阻塞，则需要其他的 goroutine 执行对应的读写操作，才能解除阻塞状态。

然而，阻塞后一直没能发生调度行为，没有可用的 goroutine 可执行，则会一直卡在这个地方，程序就失去执行意义了。此时 Go 就会报 deadlock 错误，如下代码：  
```go
 func main() {
  ch := make(chan int)
  <-ch

  // 执行后将 panic：
  // fatal error: all goroutines are asleep - deadlock!
 }
```
也就是说，在一个 goroutine 创建了 chan，不能立刻对其写入。因此，在使用 channel 时要注意 goroutine 的一发一取，避免 goroutine 永久阻塞！  

# 5. chan 是否并发安全？
是，因为 chan 主要用在多线程上，为了保证数据的一致性，必须设计为并发安全。  
在 hchan 结构体中就采用了 Mutex 锁来实现数据读取安全。

# 6. Channel 分配在栈上还是堆上？哪些对象分配在堆上，哪些对象分配在栈上？
* Channel 被设计用来实现协程间通信的组件，其作用域和生命周期不可能仅限于某个函数内部，所以 golang 直接将其分配在堆上

* 准确地说，你并不需要知道。Golang 中的变量只要被引用就一直会存活，存储在堆上还是栈上由内部实现决定而和具体的语法没有关系。
知道变量的存储位置确实和效率编程有关系。如果可能，Golang 编译器会将函数的局部变量分配到函数栈帧（stack frame）上。然而，如果编译器不能确保变量在函数 return 之后不再被引用，编译器就会将变量分配到堆上。而且，如果一个局部变量非常大，那么它也应该被分配到堆上而不是栈上。当前情况下，如果一个变量被取地址，那么它就有可能被分配到堆上,然而，还要对这些变量做逃逸分析，如果函数 return 之后，变量不再被引用，则将其分配到栈上。