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
```go
// 方法一:前缀表使用减1实现

// getNext 构造前缀表next
// params:
//		  next 前缀表数组
//		  s 模式串
func getNext(next []int, s string) {
	j := -1 // j表示 最长相等前后缀长度
	next[0] = j

	for i := 1; i < len(s); i++ {
		for j >= 0 && s[i] != s[j+1] {
			j = next[j] // 回退前一位
		}
		if s[i] == s[j+1] {
			j++
		}
		next[i] = j // next[i]是i（包括i）之前的最长相等前后缀长度
	}
}
func strStr(haystack string, needle string) int {
	if len(needle) == 0 {
		return 0
	}
	next := make([]int, len(needle))
	getNext(next, needle)
	j := -1 // 模式串的起始位置 next为-1 因此也为-1
	for i := 0; i < len(haystack); i++ {
		for j >= 0 && haystack[i] != needle[j+1] {
			j = next[j] // 寻找下一个匹配点
		}
		if haystack[i] == needle[j+1] {
			j++
		}
		if j == len(needle)-1 { // j指向了模式串的末尾
			return i - len(needle) + 1
		}
	}
	return -1
}
```
