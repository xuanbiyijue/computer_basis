# 215. 数组第k大元素
https://leetcode.cn/problems/kth-largest-element-in-an-array/

## 题目描述
给定整数数组 nums 和整数 k，请返回数组中第 k 个最大的元素。

请注意，你需要找的是数组排序后的第 k 个最大的元素，而不是第 k 个不同的元素。

你必须设计并实现时间复杂度为 O(n) 的算法解决此问题。


## 示例
```
输入: [3,2,1,5,6,4], k = 2
输出: 5
```


## 题解
```go
func findKthLargest(nums []int, k int) int {
    sort.Sort(sort.Reverse(sort.IntSlice(nums)))
    return nums[k-1]
}
```