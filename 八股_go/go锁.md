# 1. sync.Mutex
https://blog.csdn.net/qq_37005831/article/details/110311956

* 概述  
Mutex是互斥锁，Mutex的零值是解锁的Mutex。Mutex在第一次使用后不得复制。

* sync.Mutex解析
```go
type Mutex struct {
	state int32 //当前锁的状态，该int32字段通过位移操作使之可以包含不同意义
	sema  uint32 //是一个信号变量, 用于负责go程的唤醒和阻塞休眠
}

const (
	mutexLocked = 1 << iota // 锁是否被持有 即是否已经锁住
	mutexWoken // 是否有被唤醒的go程
	mutexStarving //是否处于饥饿状态，此标记可以确保某些go程不会长久获取不到锁
	mutexWaiterShift = iota // 目前等待锁的go程数量
	
	starvationThresholdNs = 1e6 //进入饥饿状态的阈值时间 1ms
)
```

* 加锁  
Lock方法被分成了两个情况进行加锁:
  * Fast Path 快速加锁方式
  * Slow Path 也就是说当前锁已经被人占用，需要进行一些其他操作。  

  现在进行一下Lock过程的总结：
  * 如果当前锁没被其他go程获取，那么就直接获取锁，这也是最直接的方式，
  * 如果当前锁被其他go程占用，并且还没有进入饥饿状态时，进行自旋等待，并且通知UnLock有正在自旋go程正在等待锁，在释放锁时就不要唤醒其他go程，然后判断是否需要进入饥饿状态，
  * 进入饥饿状态的条件是，当前go程等待锁的时间已经超出starvationThresholdNs常量所设定的时间即1ms，此时进入饥饿状态，饥饿状态会把当前go程放入等待锁的队列的最前端，使得其能在UnLock后立刻获得锁，防止该go程被饿死。退出饥饿模式需要符合一下两个条件中的任意一条：
    * 此 go程已经是队列中的最后一个 waiter 了，没有其它的等待锁的 goroutine 了；
    * 此 go程的等待时间小于 1 毫秒。

  那么当有很多的go程都在争相获取锁的时候，会按照什么顺序获取锁呢？
  等待的goroutine们是以FIFO排队的
  * 当Mutex处于正常模式时，若此时没有新goroutine与队头goroutine竞争，则队头goroutine获得。若有新goroutine竞争大概率新goroutine获得。
  * 当队头goroutine竞争锁失败1ms后，它会将Mutex调整为饥饿模式。进入饥饿模式后，锁的所有权会直接从解锁goroutine移交给队头goroutine，此时新来的goroutine直接放入队尾。
  * 当一个goroutine获取锁后，如果发现自己满足下列条件中的任何一个
    * 它是队列中最后一个
    * 它等待锁的时间少于1ms，则将锁切换回正常模式

  那么问题来了，为什么正常模式下会让新来的go程获取到锁呢？  
  因为新来的go程当前正在占用cpu的时间片，那么如果我们能够把锁交给正在占用 cpu 时间片的 go程 的话，那就不需要做上下文的切换，在高并发的情况下，可能会有更好的性能。

```go
func (m *Mutex) Lock() {
	// Fast path: grab unlocked mutex.
    // 首先进行CAS判断，首先进行判断当前的state字段是否为0，如果为0则代表当前互斥锁未被占用，因此可以直接对其进行加锁操作然后直接返回
	if atomic.CompareAndSwapInt32(&m.state, 0, mutexLocked) {
        // go提供的竞态检测器，用于检测当前是否有其他操作同时操纵此Mutex对象。
		if race.Enabled {
			race.Acquire(unsafe.Pointer(m))
		}
		return
	}
	// Slow path (outlined so that the fast path can be inlined)
	m.lockSlow()
}
```

* 解锁  
  解锁也有两种情况  
  * Fast Path： 直接将state标记位-1进行解锁
  * Slow Path： 如果返回值不为0那么就代表还有等待锁的go程这样就需要对其进行一些处理，就进入了Slow Path情况
```go
  func (m *Mutex) Unlock() {
	if race.Enabled {
		_ = m.state
		race.Release(unsafe.Pointer(m))
	}

	// Fast path: drop lock bit.
	new := atomic.AddInt32(&m.state, -mutexLocked)
	if new != 0 {
		// Outlined slow path to allow inlining the fast path.
		// To hide unlockSlow during tracing we skip one extra frame when tracing GoUnblock.
		m.unlockSlow(new)
	}
}
```
```go
/*
首选判断是否是重复解锁，如果是就报错，否则判断是否是饥饿状态如果是就直接唤醒当前等待锁队列的第一个go程，否则就判断是否是还有没有等待的go程或者是否已经被Lock标记不需要唤醒其他go程，如果是的话就直接返回，否则就将等待数量-1并且设置唤醒标记然后唤醒一个等待锁队列中的go程，让其得到锁。
*/
func (m *Mutex) unlockSlow(new int32) {
	//首选判断是不是被解锁过了，因为不可二次解锁
	if (new+mutexLocked)&mutexLocked == 0 {
		throw("sync: unlock of unlocked mutex")
	}
	//如果不是饥饿状态则做处理 否则就唤醒处于等待队列第一个的go程
	if new&mutexStarving == 0 {
		old := new
		for {
		//判断是否是最后一个go程或者唤醒标记处于已标记状态则不唤醒任何等待go程
			if old>>mutexWaiterShift == 0 || old&(mutexLocked|mutexWoken|mutexStarving) != 0 {
				return
			}
			//否则就等待队列-1 并且设置唤醒标记
			new = (old - 1<<mutexWaiterShift) | mutexWoken
			//然后进行CAS处理
			if atomic.CompareAndSwapInt32(&m.state, old, new) {
				runtime_Semrelease(&m.sema, false, 1)
				return
			}
			old = m.state
		}
	} else {
			runtime_Semrelease(&m.sema, true, 1)
	}
}
```


# 2. sync.RWMutex
https://blog.csdn.net/qq_37005831/article/details/110739530

* 概述  
读写锁，就是一个可以并发读但是不可以并发写的锁

* 解析
	```go
	type RWMutex struct {
		w           Mutex  // 一个互斥锁的字段，用户进行写时加互斥锁
		writerSem   uint32 // 一个writer的信号量，类似互斥锁中的信号量
		readerSem   uint32 // 一个reader的信号量，类似互斥锁中的信号量
		readerCount int32  // 两种作用，1:标记有多少拿到读锁的reader，2:是否有writer需要竞争
		readerWait  int32  // writer需要等待读锁解锁的reader的数量
	}
	const rwmutexMaxReaders = 1 << 30 // 最大reader的上限。即最多有多少的reader同时能拿到读锁
	```
  由于是读写锁，那么加锁解锁过程就不能像互斥锁一样只是单一的Lock和Unlock，读写锁的提供的操作有五个分别是：
  * Lock/Unlock：用于 writer(写操作) 时调用的方法，如果调用时读锁已经被reader所持有，那么将会等待，直到所有持有读锁的reader解锁后才会进行writer写锁获取。Unlock是其配对的解锁操作，解锁后通知新来的等待读锁的reader获取读锁。
  * RLock/RUnlock：用于 reader（读操作）时调用的方法，当此时没有写锁被获取时，直接获取到读锁。当有写锁被获取时，等待写锁的释放后才会被唤醒并获取读锁。RUlock 是其相反的方法，并且当没有需要等待的读锁时，会通知等待获取写锁的writer进行写锁的获取。
  * RLocker：这个方法的作用是返回一个读锁的Locker对象，调用Lock和Unlock的时候会调用RLock和RUlock，个人认为**这个方法可以构造一个只读锁**。

> 以下代码删除了所有 `if race.Enabled { //todo }` 语句，因为其是判断是进行判断当前程序是否开启了race竞态检测模式的代码，即在运行go程序时是否采用go run race xxx.go这种进行进行竞态检测运行模式，所以进行省略
* RLock()
```go
func (rw *RWMutex) RLock() {
	// 首先对读计数器进行+1 并且判断+1后的值是否小于0 如果小于0则代表当前有已经被获取的写锁
	if atomic.AddInt32(&rw.readerCount, 1) < 0 {
		// 此时需要进行阻塞挂起，等待写锁的解锁
		runtime_SemacquireMutex(&rw.readerSem, false, 0)
	}
}
```

* RUnlock()
```go
func (rw *RWMutex) RUnlock() {
		// 将已经加锁的读锁数量-1，如果此时-1后小于0时，则代表
		// 1:有可能反复解锁，此时需要抛出panic
		// 2:有writer正在等待获取写锁
		if r := atomic.AddInt32(&rw.readerCount, -1); r < 0 {
		rw.rUnlockSlow(r)
	}
}

func (rw *RWMutex) rUnlockSlow(r int32) {
	// 不可重复解锁，此时抛出panic
	if r+1 == 0 || r+1 == -rwmutexMaxReaders {
		race.Enable()
		throw("sync: RUnlock of unlocked RWMutex")
	}
	// 此时有一个writer正在等待获取写锁，
	// 如果当前解锁的reader是最后一个需要等待的读锁
	// 则唤醒等待读锁释放完的writer进行写锁的获取
	if atomic.AddInt32(&rw.readerWait, -1) == 0 {
		runtime_Semrelease(&rw.writerSem, false, 1)
	}
}
```

* Lock()  
写锁的加锁过程必须先对整体的结构体的Mutex进行加锁，以免有其他的写操作同时对写锁的竞争导致data race。然后进行当前持有读锁的reader的数量进行取反，并且将其值交给readerWait用于标记需要等待释放锁的reader的数量，如果该字段不等于0则代表需要进行读锁解锁等待。当reader调用RUlock时会进行对此字段的-1并且判断，如果此字段为0时，则唤醒writer的阻塞，使得writer获取到写锁。
```go
func (rw *RWMutex) Lock() { 
	// 先将Mutex字段进行加锁，以免有其他写锁操作或者其他操作破坏数据
	rw.w.Lock()
	// 将readerCount进行取反操作 这也是此字段除了标记reader数量的第二个功能，进行写锁标记
	// 即标记有writer需要竞争
	r := atomic.AddInt32(&rw.readerCount, -rwmutexMaxReaders) + rwmutexMaxReaders
	// 此时将取反的r值交给readerWait代表仍需要等待释放锁的reader的数量
	// 如果该数量为0 那么代表不需要等待则直接获取写锁即可
	// 否则就将writer挂起阻塞直至RUlock唤醒
	if r != 0 && atomic.AddInt32(&rw.readerWait, r) != 0 {
		runtime_SemacquireMutex(&rw.writerSem, false, 0)
	}
 }
```

* Unlock()  
写锁的解锁方式很简单，先进行readerCount的取反，以便告知无writer正在竞争，然后依次去唤醒这些等待的reader去获取读锁，然后将互斥锁写锁，以便后续的writer进行写操作，在写操作时，加锁时先进行互斥锁的加锁，解锁时后进行互斥锁的解锁，为的是保证字段的修改也受到互斥锁的保护。
```go
func (rw *RWMutex) Unlock() {
	// 写锁进行解锁时首先将加锁时取反的readerCount再次取反
	// 也就是解除当前有写锁正在竞争的标记
	r := atomic.AddInt32(&rw.readerCount, rwmutexMaxReaders)
	// 如果取反后这个值大于rwmutexMaxReaders 就代表重复解锁
	// 抛出panic
	if r >= rwmutexMaxReaders {
		race.Enable()
		throw("sync: Unlock of unlocked RWMutex")
	}
	// 解锁完毕后需要根据等待的readerCount的数量去依次唤醒这些reader 
	// 这些reader是在Lock后再次请求获取读锁的reader的数量
	for i := 0; i < int(r); i++ {
		runtime_Semrelease(&rw.readerSem, false, 0)
	}
	// 把写锁的互斥锁解锁，以便于其他writer进行写操作的竞争
	rw.w.Unlock()
}
```

# 3. 除了 mutex 以外还有那些方式安全读写共享变量？
* channel
* 原子性操作（原子性操作可以通过sync/atomic包实现）

# 4. 原子操作与互斥锁的区别
* 使用目的：互斥锁是用来保护一段逻辑，原子操作用于对一个变量的更新保护。
* 底层实现：Mutex由操作系统的调度器实现，而atomic包中的原子操作则由底层硬件指令直接提供支持，这些指令在执行的过程中是不允许中断的，因此原子操作可以在lock-free的情况下保证并发安全，并且它的性能也能做到随CPU个数的增多而线性扩展。

# 5. Mutex 是悲观锁还是乐观锁？悲观锁、乐观锁是什么？
悲观锁。
* 悲观锁：当要对数据库中的一条数据进行修改的时候，为了避免同时被其他人修改，最好的办法就是直接对该数据进行加锁以防止并发。这种借助数据库锁机制，在修改数据之前先锁定，再修改的方式被称之为悲观并发控制
* 乐观锁：是相对悲观锁而言的，乐观锁假设数据一般情况不会造成冲突，所以在数据进行提交更新的时候，才会正式对数据的冲突与否进行检测，如果冲突，则返回给用户异常信息，让用户决定如何去做。乐观锁适用于读多写少的场景，这样可以提高程序的吞吐量

# 6. Mutex 有几种模式？
* 正常模式：  
  * 当前的mutex只有一个goruntine来获取，那么没有竞争，直接返回。
  * 新的goruntine进来，如果当前mutex已经被获取了，则该goruntine进入一个先入先出的waiter队列，在mutex被释放后，waiter按照先进先出的方式获取锁。
  * 新的goruntine进来，mutex处于空闲状态，将参与竞争。新来的 goroutine 有先天的优势，在高并发情况下，被唤醒的 waiter 可能比较悲剧地获取不到锁，这时，它会被插入到队列的前面。如果 waiter 获取不到锁的时间超过阈值 1 毫秒，那么，这个 Mutex 就进入到了饥饿模式。

* 饥饿模式  
在饥饿模式下，Mutex 的拥有者将直接把锁交给队列最前面的 waiter。新来的 goroutine 不会尝试获取锁，并加入到等待队列的尾部。 如果拥有 Mutex 的 waiter 发现下面两种情况的其中之一，它就会把这个 Mutex 转换成正常模式:
  * 此 waiter 已经是队列中的最后一个 waiter 了，没有其它的等待锁的 goroutine 了；
  * 此 waiter 的等待时间小于 1 毫秒。

# 7. 死锁
两种情况下发生死锁：
* 如果同一个线程先后两次调用lock，在第二次调用时，由于锁已经被占用，该线程会挂起等待别的线程释放锁，然而锁正是被自己占用着的，该线程又被挂起而没有机会释放锁，因此就永远处于挂起等待状态了，这叫做死锁（Deadlock）。   
* 若线程A获得了锁1，线程B获得了锁2，这时线程A调用lock试图获得锁2，结果是需要挂起等待线程B释放锁2，而这时线程B也调用lock试图获得锁1，结果是需要挂起等待线程A释放锁1，于是线程A和B都永远处于挂起状态了。  

为了避免死锁，我们需要遵循以下几个原则：
* 避免嵌套锁；
* 避免长时间持有锁；
* 避免锁的交叉依赖；
* 避免使用全局变量。  

如果出现了死锁，我们可以通过以下几种方式来解决：
* 通过goroutine和channel来解决；
* 通过超时机制来解决；
* 通过Context来解决