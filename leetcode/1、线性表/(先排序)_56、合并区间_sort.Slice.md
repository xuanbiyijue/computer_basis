# 56. 合并区间
https://leetcode.cn/problems/merge-intervals/

## 题目描述
以数组 intervals 表示若干个区间的集合，其中单个区间为 intervals[i] = [starti, endi] 。请你合并所有重叠的区间，并返回 一个不重叠的区间数组，该数组需恰好覆盖输入中的所有区间 。

## 示例
```
输入：intervals = [[1,3],[2,6],[8,10],[15,18]]
输出：[[1,6],[8,10],[15,18]]
解释：区间 [1,3] 和 [2,6] 重叠, 将它们合并为 [1,6].
```
```
输入：intervals = [[1,4],[4,5]]
输出：[[1,5]]
解释：区间 [1,4] 和 [4,5] 可被视为重叠区间。
```

## 题解
遇事不决，就先排序。如果我们按照区间的左端点排序，那么在排完序的列表中，可以合并的区间一定是连续的。
```go
func merge(intervals [][]int) [][]int {
    res := make([][]int, 0)
    // 先排序
    sort.Slice(intervals, func(i, j int) bool {
		if intervals[i][0] == intervals[j][0] {
			//若左端点相等，则按照右端点升序排序
			return intervals[i][1] <= intervals[j][1]
		} else {
			//若左端点不相等，则按照左端点升序排序
			return intervals[i][0] < intervals[j][0]
		}
	})

    for i:=1; i<len(intervals); i++ {
        if intervals[i][0] <= intervals[i-1][1] {
            intervals[i][0] = intervals[i-1][0]
            if intervals[i][1] < intervals[i-1][1] {
                intervals[i][1] = intervals[i-1][1]
            }
        }else {
            res = append(res, intervals[i-1])
        }
    }
    // 处理最后一个
    res = append(res, intervals[len(intervals)-1])

    return res
}
```
* 20230512  
```go
func merge(intervals [][]int) [][]int {
    sort.Slice(intervals, func(i, j int) bool {
        if intervals[i][0] < intervals[j][0] {
            return true
        } else if intervals[i][0] == intervals[j][0] {
            return intervals[i][1] < intervals[j][1]
        }
        return false
    })
    res := make([][]int, 0)
    for i:=1; i<len(intervals); i++ {
        if intervals[i][0] <= intervals[i-1][1] {
            if intervals[i][1] <= intervals[i-1][1] {
                intervals[i] = intervals[i-1]
            } else {
                intervals[i] = []int{intervals[i-1][0], intervals[i][1]}
            }
        } else {
            res = append(res, intervals[i-1])
        }
    }
    res = append(res, intervals[len(intervals)-1])
    return res 
}
```
 
 
# 57、插入区间
https://leetcode.cn/problems/insert-interval/

## 题目描述
给你一个 无重叠的 ，按照区间起始端点排序的区间列表。

在列表中插入一个新的区间，你需要确保列表中的区间仍然有序且不重叠（如果有必要的话，可以合并区间）。

## 示例
```
输入：intervals = [[1,3],[6,9]], newInterval = [2,5]
输出：[[1,5],[6,9]]
```
```
输入：intervals = [[1,2],[3,5],[6,7],[8,10],[12,16]], newInterval = [4,8]
输出：[[1,2],[3,10],[12,16]]
解释：这是因为新的区间 [4,8] 与 [3,5],[6,7],[8,10] 重叠。
```
```
输入：intervals = [], newInterval = [5,7]
输出：[[5,7]]
```

## 题解
```go
func insert(intervals [][]int, newInterval []int) (ans [][]int) {
    left, right := newInterval[0], newInterval[1]
    merged := false
    for _, interval := range intervals {
        if interval[0] > right {
            // 在插入区间的右侧且无交集
            if !merged {
                ans = append(ans, []int{left, right})
                merged = true
            }
            ans = append(ans, interval)
        } else if interval[1] < left {
            // 在插入区间的左侧且无交集
            ans = append(ans, interval)
        } else {
            // 与插入区间有交集，计算它们的并集
            left = min(left, interval[0])
            right = max(right, interval[1])
        }
    }
    if !merged {
        ans = append(ans, []int{left, right})
    }
    return
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}
```

* 20230514  
```go
func insert(intervals [][]int, newInterval []int) (ans [][]int) {
    left, right := newInterval[0], newInterval[1]
    merged := false
    for _, interval := range intervals {
        if right < interval[0] {
            if !merged {
                ans = append(ans, []int{left, right})
                merged = true
            }
            ans = append(ans, interval)
        } else if left > interval[1] {
            ans = append(ans, interval)
        } else {
            left, right = min(left, interval[0]), max(right, interval[1])
        }
    }
    if !merged {
        ans = append(ans, []int{left, right})
    }
    return
}

func max(a, b int) int {if a > b {return a}; return b}
func min(a, b int) int {if a < b {return a}; return b}
```


# 228. 汇总区间
https://leetcode.cn/problems/summary-ranges/

## 题目描述
给定一个  无重复元素 的 有序 整数数组 nums 。

返回 恰好覆盖数组中所有数字 的 最小有序 区间范围列表 。也就是说，nums 的每个元素都恰好被某个区间范围所覆盖，并且不存在属于某个范围但不属于 nums 的数字 x 。

列表中的每个区间范围 [a,b] 应该按如下格式输出：

* "a->b" ，如果 a != b
* "a" ，如果 a == b

## 示例
```
输入：nums = [0,1,2,4,5,7]
输出：["0->2","4->5","7"]
解释：区间范围是：
[0,2] --> "0->2"
[4,5] --> "4->5"
[7,7] --> "7"
```
```
输入：nums = [0,2,3,4,6,8,9]
输出：["0","2->4","6","8->9"]
解释：区间范围是：
[0,0] --> "0"
[2,4] --> "2->4"
[6,6] --> "6"
[8,9] --> "8->9"
```

## 题解
```go
func summaryRanges(nums []int) []string {
    res := make([]string, 0)
    for i:=0; i<len(nums);  {
        left := i 
        for i++; i<len(nums)&&nums[i]-nums[i-1]==1; i++ {}
        cur := strconv.Itoa(nums[left])
        if left < i-1 {
            cur += ("->" + strconv.Itoa(nums[i-1]))
        }
        res = append(res, cur)
    }
    return res
}
```

## 总结
* 可以用循环里的循环来更新i

