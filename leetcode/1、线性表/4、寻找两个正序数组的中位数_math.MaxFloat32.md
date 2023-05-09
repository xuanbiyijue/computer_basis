# 寻找两个正序数组的中位数（困难）

链接：https://leetcode.cn/problems/median-of-two-sorted-arrays/

## 题目描述
给定两个大小分别为 m 和 n 的正序（从小到大）数组 nums1 和 nums2。请你找出并返回这两个正序数组的 中位数 。要求复杂度 $O(log(M+N))$

## 示例
```
输入：nums1 = [1,3], nums2 = [2]
输出：2.00000
解释：合并数组 = [1,2,3] ，中位数 2
```

```
输入：nums1 = [1,2], nums2 = [3,4]
输出：2.50000
解释：合并数组 = [1,2,3,4] ，中位数 (2 + 3) / 2 = 2.5
```

## 提示
* nums1.length == m
* nums2.length == n
* 0 <= m <= 1000
* 0 <= n <= 1000
* 1 <= m + n <= 2000
* $-10^6$ <= nums1[i], nums2[i] <= $10^6$


## 解法
* 解法1：注意两个给定数组都为正序，组建一个新数组，再使用双指针。这种解法时间复杂度为 $O(m+n)$，空间复杂度为$O(m+n)$，较为简单
```go
func findMedianSortedArrays(nums1 []int, nums2 []int) float64 {
    mergeNums := make([]int, len(nums1)+len(nums2))
    l, r := 0, 0
    for i:=0; i<len(mergeNums); i++ {
        if l < len(nums1) && r < len(nums2) {
            if nums1[l] < nums2[r] {
                mergeNums[i] = nums1[l]
                l += 1
            } else {
                mergeNums[i] = nums2[r]
                r += 1
            }
        } else {
            // 处理剩余元素
            if l == len(nums1) {
                mergeNums[i] = nums2[r]
                r += 1
            } else {
                mergeNums[i] = nums1[l]
                l += 1
            }
        }
    }
    l, r = 0, 0
    for ; r<len(mergeNums)-1; l,r=l+1,r+2 {}
    if r == len(mergeNums)-1 {
        return float64(mergeNums[l])
    }
    return float64(mergeNums[l] + mergeNums[l-1]) / 2.0
}
```

* 解法2：
```go

```

## 总结
* 创建数组：`make([]int, len(nums1)+len(nums2))`
* 强制数据转换：`float64()`
* 两种浮点型数：float32 和 float64，分别使用 `math.MaxFloat32` 和 `math.MaxFloat64`