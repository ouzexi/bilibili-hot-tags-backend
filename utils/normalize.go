package utils

import (
	"sort"
	"strings"
)

type RequestParams struct {
	Keyword string `json:"keyword"`
	Order   string `json:"order"`
}
type ResponseData struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Ttl     int                `json:"ttl"`
	Data    ResponseDataResult `json:"data"`
}

type ResponseDataResult struct {
	Result []ResponseDataResultTags `json:"result"`
}

type ResponseDataResultTags struct {
	Tag string `json:"tag"`
}

type ChartData struct {
	Item  string `json:"item"`
	Count int    `json:"count"`
}

func TransferRes(res ResponseData) []ChartData {
	result := res.Data.Result
	var tags []string
	for _, rItem := range result {
		tags = append(tags, rItem.Tag)
	}

	tagsMap := make(map[string]int)

	for _, tagsItem := range tags {
		tItems := strings.Split(func() string {
			if len(tagsItem) > 0 {
				return tagsItem
			}
			return ""
		}(), ",")

		for _, tItem := range tItems {
			if _, ok := tagsMap[tItem]; !ok {
				tagsMap[tItem] = 0
			}
			tagsMap[tItem]++
		}
	}

	// 创建一个切片来存储map的键
	keys := make([]string, 0, len(tagsMap))
	for k := range tagsMap {
		keys = append(keys, k)
	}

	// 根据tagsMap的值对keys进行排序
	sort.Slice(keys, func(i, j int) bool {
		return tagsMap[keys[i]] > tagsMap[keys[j]] // 降序排列
	})

	// 获取前10个元素
	if len(keys) > 10 {
		keys = keys[:10]
	}

	chartData := make([]ChartData, 0)
	for _, k := range keys {
		chartData = append(chartData, ChartData{
			Item:  k,
			Count: tagsMap[k],
		})
	}
	return chartData
}
