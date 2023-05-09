# 49. 字母异位词分组
链接: https://leetcode.cn/problems/group-anagrams/

## 题目描述
给你一个字符串数组，请你将 字母异位词 组合在一起。可以按任意顺序返回结果列表。

字母异位词 是由重新排列源单词的字母得到的一个新单词，所有源单词中的字母通常恰好只用一次。

## 示例
```
输入: strs = ["eat", "tea", "tan", "ate", "nat", "bat"]
输出: [["bat"],["nat","tan"],["ate","eat","tea"]]
```

```
输入: strs = [""]
输出: [[""]]
```

```
输入: strs = ["a"]
输出: [["a"]]
```

## 题解
由于互为字母异位词的两个字符串包含的字母相同，因此两个字符串中的相同字母出现的次数一定是相同的，故可以将每个字母出现的次数使用字符串表示，作为哈希表的键。由于字符串只包含小写字母，因此对于每个字符串，可以使用长度为 2626 的数组记录每个字母出现的次数。
```go
func groupAnagrams(strs []string) [][]string {
    mp := map[[26]int][]string{}
    for _, str := range strs {
        cnt := [26]int{}
        for _, b := range str {
            cnt[b-'a']++
        }
        mp[cnt] = append(mp[cnt], str)
    }
    ans := make([][]string, 0, len(mp))
    for _, v := range mp {
        ans = append(ans, v)
    }
    return ans
}
```

## 总结
* golang可以使用数组作为map的key



# 187. 重复的DNA序列
https://leetcode.cn/problems/repeated-dna-sequences/

## 题目描述
DNA序列 由一系列核苷酸组成，缩写为 'A', 'C', 'G' 和 'T'.。

例如，"ACGAATTCCG" 是一个 DNA序列 。
在研究 DNA 时，识别 DNA 中的重复序列非常有用。

给定一个表示 DNA序列 的字符串 s ，返回所有在 DNA 分子中出现不止一次的 长度为 10 的序列(子字符串)。你可以按 任意顺序 返回答案。


## 示例
```
输入：s = "AAAAACCCCCAAAAACCCCCCAAAAAGGGTTT"
输出：["AAAAACCCCC","CCCCCAAAAA"]
```

## 题解
```go
func findRepeatedDnaSequences(s string) []string {
    res := make([]string, 0)
    dict := make(map[string]int)
    for i:=0; i<=len(s)-10; i++ {
        sub := s[i:i+10]
        dict[sub]++
        if dict[sub] == 2 {
            res = append(res, sub)
        }
    } 
    return res
}
```

## 总结
* 空map可以直接 `dict[sub]++`
* 使用 `dict[sub] == 2` 保证只append一次



# 