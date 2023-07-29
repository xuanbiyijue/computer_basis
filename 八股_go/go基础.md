# make 和 new 的区别
* 共同点：用于分配内存
* 不同点：
  * 作用的类型不同：make -> (chan、slice、map), new -> (string、int、数组)
  * 返回的类型不同：make -> 变量本身（并初始化），new -> 指针（并将该块内存置为0）

* 补充：分配的内存，是在堆还是栈上？  
https://juejin.cn/post/7048823640925667364  
  > **内存逃逸**简单来讲就是 程序中函数都有自己的局部变量和返回地址空间（栈帧），当某个变量想要在函数结束之后继续使用，需要将其分配到堆上，这种从栈上逃逸到堆上的现象 就是内存逃逸（这个说法不严谨）。  
  **逃逸分析**就是程序在编译期间通过分析函数中变量，哪些需要在堆上分配哪些在栈上分配。逃逸分析为什么是在编译期间做的呢，因为函数局部变量占用多少内存需要在程序编译期间确定，运行期间无法更改。go的逃逸分析 会遵循2个原则：1. 指向栈对象的指针不能存在堆上。2. 指向栈对象的指针生命周期不能超过该对象生命周期，也就是栈对象指针不能在栈对象销毁后继续存活。  

  * 对于make和new分配的内存，go编译器尽量将变量分配在栈上，如果变量未发生逃逸，那么就会在栈上分配，否则分配在堆上。
  * 如果变量占用内存很大超过了栈空间，或者栈空间不足，那么就会分配在堆上。注意，占用内存很大的局部变量，这个阈值 不同版本golang的大小限制不一样。

# select 语句
* select 是 Go 中的一个控制结构，类似于 switch 语句。
* select 语句只能用于通道操作，每个 case 必须是一个通道操作，要么是发送要么是接收。
* select 语句会监听所有指定的通道上的操作，一旦其中一个通道准备好就会执行相应的代码块。如果多个通道都准备好，那么 select 语句会随机选择一个通道执行。如果所有通道都没有准备好，那么执行 default 块中的代码。
* 至少有一个 case，不然会 panic
* 已关闭的 channel 也是可读的
```go
select {
  case <- channel1:
    // 执行的代码
  case value := <- channel2:
    // 执行的代码
  case channel3 <- value:
    // 执行的代码

    // 你可以定义任意数量的 case

  default:
    // 所有通道都没有准备好，执行的代码
}
```
* select 的底层原理：  
https://www.jianshu.com/p/5923cde1b6a3
  * select这个语句底层实现实际上主要由两部分组成：case语句和执行函数。
  * Go 实现 select 时，定义了一个数据结构表示每个 case 语句(包含defaut)，select 执行过程可以类比成一个函数，函数输入 case 数组，输出选中的 case，然后程序流程转到选中的 case块。
  * 然后执行select语句实际上就是调用 `func selectgo(cas0 *scase, order0 *uint16, ncases int) (int, bool)` 函数。
```go
type scase struct {
    c           *hchan         // chan
    elem        unsafe.Pointer // 读或者写的缓冲区地址
    kind        uint16   //case语句的类型，是default、传值写数据(channel <-) 还是  取值读数据(<- channel)
    pc          uintptr // race pc (for race detector / msan), 竞争检测相关
    releasetime int64
}
```
  


# 数组和切片的区别
* 共同点：
  * 只能存储一组相同类型的数据结构
  * 都是通过下标来访问，并且有长度和容量，长度通过 len 获取，容量通过 cap 获取
  * 在函数传递中都是值传递（只会拷贝切片本身（指针、长度和容量），不会拷贝底层的数组数据）
* 不同点：
  * 数组是定长；切片可以自动扩容；
  * 数组是值类型，切片是引用类型（每个切片都引用了一个底层数组，切片本身不能存储任何数据，修改切片的时候修改的是底层数组中的数据。切片一旦扩容，指向一个新的底层数组，内存地址也就随之改变）


# for range 的时候它的地址会发生变化么？
不会。在 `for a, b := range c` 遍历中， a 和 b 在内存中只会存在一份，即之后每次循环时遍历到的数据都是以值覆盖的方式赋给 a 和 b，a，b 的内存地址始终不变。由于有这个特性，for 循环里面如果开协程，不要直接把 a 或者 b 的地址传给协程。

# defer，多个 defer 的顺序，defer 在什么时机会修改返回值？
https://blog.csdn.net/Cassie_zkq/article/details/108567205

* 作用：延迟函数的执行（一般用于已打开资源的关闭或捕获panic）
* 顺序：后入先出（栈）。defer的底层实现是链表组成的栈，每次插在head.Next
* defer、return、返回值三者的执行逻辑:
  * return最先执行，return负责将结果写入返回值中
  * 接着defer开始执行
  * 最后函数携带当前返回值（可能和最初的返回值不相同）退出
```go
// 无名返回值
func a() int {
	var i int
	defer func() {
		i++
		fmt.Println("defer2:", i) 
	}()
	defer func() {
		i++
		fmt.Println("defer1:", i) 
	}()
	return i
}
/*
输出为：
defer1: 1
defer2: 2
return: 0

解释：
返回值由变量i赋值，则返回值i=0。第二个defer中i++ = 1， 第一个defer中i++ = 2，所以最终i的值是2。但是返回值已经被赋值了，即使后续修改i也不会影响返回值。最终返回值返回，所以main中打印0。
*/
```
```go
// 有名返回值
func b() (i int) {
	defer func() {
		i++
		fmt.Println("defer2:", i)
	}()
	defer func() {
		i++
		fmt.Println("defer1:", i)
	}()
	return i //或者直接写成return
}
/*
输出为：
defer1: 1
defer2: 2
return: 2

解释：
这里已经指明了返回值就是i，所以后续对i进行修改都相当于在修改返回值，所以最终函数的返回值是2。
*/
```

# rune 类型
rune 等价于 int32, byte 等价于int8

# 反射与序列化
* 反射：程序在运行时可以访问、检测和修改它本身状态或行为；能够获取给定数据对象的类型和结构，并有机会修改它。
* go语言反射里最重要的两个概念是**Type**和**Value**：
  * Type用于获取类型相关的信息（比如Slice的长度，struct的成员，函数的参数个数）
  * Value用于获取和修改原始数据的值（比如修改slice和map中的元素，修改struct的成员变量）。
* 结构体中的tag就是通过反射实现的（序列化Marshal：转换为比特流；反序列化Unmarshal）
* **Marshal要点**：
  * 只要是可导出成员（变量首字母大写），都可以转成json。
  * 如果变量打上了json标签，如Name旁边的 `json:"name"` ，那么转化成的json key就用该标签“name”，否则取变量名作为key
  * 指针变量，编码时自动转换为它所指向的值
  * json 在序列化遇到管道/函数等无法被序列化的情况时，会发生 error 错误。
  * 小写的变量名不会被序列化
```go
import (
	"fmt"
	"reflect"
)

type User struct {
	name string `json:name-field`
	age  int
}

func main() {
	user := &User{"John Doe The Fourth", 20}

	field, ok := reflect.TypeOf(user).Elem().FieldByName("name")
	if !ok {
		panic("Field not found")
	}
	fmt.Println(getStructTag(field))  // 输出：json:name-field
}

func getStructTag(f reflect.StructField) string {
	return string(f.Tag)
}
```

# map
* map 是否并发安全？  
map 是一个**非线程安全**的类型。那么如何使得map变成线程安全？

* 怎么对 map 进行并发访问？
  * 使用Mutex  
  使用互斥锁对map对其进行封装，实现如下：
  ```go
  type MutexMap struct {
    lock sync.Mutex
    m    map[string]int
  }

  func (receiver *MutexMap) Get(key string) (int, bool) {
    receiver.lock.Lock()
    value, ok := receiver.m[key]
    receiver.lock.Unlock()
    return value, ok
  }

  func (receiver *MutexMap) Set(key string, value int) {
    receiver.lock.Lock()
    receiver.m[key] = value
    receiver.lock.Unlock()
  }
  func (receiver *MutexMap) Del(key string) {
    receiver.lock.Lock()
    delete(receiver.m, key)
    receiver.lock.Unlock()
  }
  ```
  **改进**：由于互斥锁的特性，虽然可以提供线程安全的 map，但是在大量并发读写的情况下，锁的竞争会非常激烈。尤其在读的并发非常大的时候，互斥锁会严重影响读取的性能，因为在通过map底层源码发现，并发的读并不会产生panic只有并发读写时，才会发生，因此，互斥锁会导致整体读的效率下降很多，此时我们就应该使用读写锁来进行优化
  ```go
  type RWMutexMap struct {
    lock sync.RWMutex
    m    map[string]int
  }

  func (receiver *RWMutexMap) Get(key string) (int, bool) {
    receiver.lock.RLock()
    value, ok := receiver.m[key]
    receiver.lock.RUnlock()
    return value, ok
  }

  func (receiver *RWMutexMap) Set(key string, value int) {
    receiver.lock.Lock()
    receiver.m[key] = value
    receiver.lock.Unlock()
  }
  func (receiver *RWMutexMap) Del(key string) {
    receiver.lock.Lock()
    delete(receiver.m, key)
    receiver.lock.Unlock()
  }
  ```

  * 使用标准库的 sync.Map  
  标准库中提供了一种官方实现的线程安全的map结构，即sync.Map（性能不错，但在读写都多的情况下不行）。这个 sync.Map并不是用来替换内建的 map类型的，它被应用在一些特殊的场景里。官方文档指出可以用在以下场景：
    * 1.只会增长的缓存系统中，一个 key 只写入一次而被读很多次，即读多写少
    * 2.多个 goroutine 为不相交的键集读、写和重写键值对，即多个goroutineCRUD操作不同的key-value

* 未初始化的map（nil map）和空map之间的区别？
  * 可以对未初始化的map进行取值，但取出来的东西是空
  * 不能对未初始化的map进行赋值，这样将会抛出一个异常
  * 通过fmt打印map时，空map和nil map结果是一样的，都为map[]

* 那些类型可以作为 map 的 key？  
在golang规范中，可比较的类型都可以作为map key  
**不能**作为map key 的类型包括：
  * slices
  * maps
  * functions


* map 的底层实现   
https://zhuanlan.zhihu.com/p/273666774
https://cloud.tencent.com/developer/article/1746966
> 简答：底层使用 hash table，并用链表（拉链法）来解决冲突。每个 map 的底层结构是 hmap，是有若干个结构为 bmap 的 bucket 组成的数组。出现冲突时，不是每一个 key 都申请一个结构通过链表串起来，而是以 bmap 为最小粒度挂载，一个 bmap 可以放 8 个 kv。在哈希函数的选择上，会在程序启动时，检测 cpu 是否支持 aes，如果支持，则使用 aes hash，否则使用 memhash。

golang 中 map 是 hashmap 实现的（有的用搜索树实现的）。 源码如下：
  ```go
  // Map contains Type fields specific to maps.
  type Map struct {
      Key  *Type // Key type
      Elem *Type // Val (elem) type

      Bucket *Type // internal struct type representing a hash bucket
      Hmap   *Type // internal struct type representing the Hmap (map header object)
      Hiter  *Type // internal struct type representing hash iterator state
  }
  ```
  前两个字段分别为 key 和 value, 由于 go map 支持多种数据类型, go 会在编译期推断其具体的数据类型；Bucket 是哈希桶； Hmap 表征了 map 底层使用的 HashTable 的元信息, 如当前 HashTable 中含有的元素数据、桶指针等； Hiter 是用于遍历 go map 的数据结构, 将在下文中讨论。  
  **最核心的是 hmap**，源码如下：
  ```go
  type hmap struct {
      count     int // 代表哈希表中的元素个数，调用len(map)时，返回的就是该字段值。
      flags     uint8 // 状态标志，下文常量中会解释四种状态位含义。
      B         uint8  // buckets（桶）的数量（2^B）
      noverflow uint16 // 溢出桶的大概数量。
      hash0     uint32 // 哈希种子。

      buckets    unsafe.Pointer // 指向buckets数组的指针，数组大小为2^B，如果元素个数为0，它为nil。
      oldbuckets unsafe.Pointer // 如果发生扩容，oldbuckets是指向老的buckets数组的指针，老的buckets数组大小是新的buckets的1/2。非扩容状态下，它为nil。
      nevacuate  uintptr        // 表示扩容进度，小于此地址的buckets代表已搬迁完成。

      extra *mapextra // 表示溢出桶
  }
  type mapextra struct {
    overflow    *[]*bmap
    oldoverflow *[]*bmap
    nextOverflow *bmap
  }
  // A bucket for a Go map.
  type bmap struct {
      // tophash包含此桶中每个键的哈希值最高字节（高8位）信息（也就是前面所述的high-order bits）。
      // 如果tophash[0] < minTopHash，tophash[0]则代表桶的搬迁（evacuation）状态。
      tophash [bucketCnt]uint8
  }
  ```
  哈希表 runtime.hmap 指向的桶的类型是 runtime.bmap数组。每一个 runtime.bmap 都能存储 8 个键值对，当哈希表中存储的数据过多，单个桶已经装满时就会使用 extra.nextOverflow 中桶存储溢出的数据。因为桶中最多只能装8个键值对，如果有多余的键值对落到了当前桶，那么就需要再构建一个桶（称为溢出桶），通过overflow指针链接起来。
  ![img](https://pic4.zhimg.com/80/v2-fceb2feda6a4dd151f0730e751436b3f_720w.webp)


* map 的扩容  
以下两种情况发生时触发哈希的扩容：
  * 装载因子已经超过 6.5；
  * 使用了太多溢出桶；  

  两种情况官方采用了不同的解决方案
  * 针对 1，将 B + 1，新建一个buckets数组，新的buckets大小是原来的2倍，然后旧buckets数据搬迁到新的buckets。该方法我们称之为增量扩容。
  * 针对 2，并不扩大容量，buckets数量维持不变，重新做一遍类似增量扩容的搬迁动作，把松散的键值对重新排列一次，以使bucket的使用率更高，进而保证更快的存取。该方法我们称之为等量扩容。  

* map 的遍历
外层循环遍历所有 Bucket, 中层循环横向遍历所有溢出桶, 内层循环遍历 Bucket 的所有 k/v 




# Go 语言中不同的类型如何比较是否相等？
* 像 string，int，float interface 等可以通过 reflect.DeepEqual 和等于号进行比较，
* 像 slice，struct，map 则一般使用 reflect.DeepEqual 来检测是否相等。

# Go 中 init 函数的特征?
一个包下可以有多个 init 函数，每个文件也可以有多个 init 函数。多个 init 函数按照它们的文件名顺序逐个初始化。应用初始化时初始化工作的顺序是，从被导入的最深层包开始进行初始化，层层递出最后到 main 包。不管包被导入多少次，包内的 init 函数只会执行一次。应用初始化时初始化工作的顺序是，从被导入的最深层包开始进行初始化，层层递出最后到 main 包。但包级别变量的初始化先于包内 init 函数的执行。

# Go 中 uintptr 和 unsafe.Pointer 的区别？
unsafe.Pointer 是通用指针类型，它不能参与计算，任何类型的指针都可以转化成 unsafe.Pointer，unsafe.Pointer 可以转化成任何类型的指针，uintptr 可以转换为 unsafe.Pointer，unsafe.Pointer 可以转换为 uintptr。uintptr 是指针运算的工具，但是它不能持有指针对象（意思就是它跟指针对象不能互相转换），unsafe.Pointer 是指针对象进行运算（也就是 uintptr）的桥梁。

# Go 多返回值怎么实现的？
Go 传参和返回值是通过 FP+offset 实现，并且存储在调用函数的栈帧中。FP 栈底寄存器，指向一个函数栈的顶部;PC 程序计数器，指向下一条执行指令;SB 指向静态数据的基指针，全局符号;SP 栈顶寄存器。

