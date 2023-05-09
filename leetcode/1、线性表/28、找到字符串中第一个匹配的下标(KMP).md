# 28. 找出字符串中第一个匹配项的下标
https://leetcode.cn/problems/find-the-index-of-the-first-occurrence-in-a-string/

## 题目描述
给你两个字符串 haystack 和 needle ，请你在 haystack 字符串中找出 needle 字符串的第一个匹配项的下标（下标从 0 开始）。如果 needle 不是 haystack 的一部分，则返回  -1 。

## 示例
```
输入：haystack = "sadbutsad", needle = "sad"
输出：0
解释："sad" 在下标 0 和 6 处匹配。
第一个匹配项的下标是 0 ，所以返回 0 。
```
```
输入：haystack = "leetcode", needle = "leeto"
输出：-1
解释："leeto" 没有在 "leetcode" 中出现，所以返回 -1 。
```

## 题解
* 解法1: 利用go的特性解题
```go
func strStr(haystack string, needle string) int {
    for i:=0; i<=len(haystack)-len(needle); i++ {
        subStr := haystack[i:i+len(needle)]
        if subStr == needle {
            return i 
        }
    }
    return -1
}
```

* 解法2：指针+暴力破解
```go
func strStr(haystack string, needle string) int {
    for i:=0; i<=len(haystack)-len(needle); i++ {
        flag := true
        p_needle := 0
        for j:=i; j-i<len(needle); j++ {
            if haystack[j] != needle[p_needle] {
                flag = false
                break
            }
            p_needle++
        }
        if flag == true {
            return i 
        }
    }
    return -1
}
```

* 解法3: KMP算法  
待定