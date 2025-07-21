package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type POI struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Type     string `json:"type"`
	Typecode string `json:"typecode"`
	Address  string `json:"address"`
	Tel      string `json:"tel"`
}

type APIResp struct {
	Pois []POI `json:"pois"`
}

func main() {
	key := os.Getenv("AMAP_KEY")
	if key == "" {
		fmt.Println("请先设置环境变量AMAP_KEY")
		return
	}
	city := "北京"
	keyword := "三甲医院"
	outputFile := "beijing_tier3_hospitals_by_keyword_go.json"

	allPois := make([]POI, 0)
	page := 1
	for {
		url := fmt.Sprintf("https://restapi.amap.com/v3/place/text?key=%s&keywords=%s&city=%s&citylimit=true&offset=25&page=%d", key, keyword, city, page)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("请求失败:", err)
			break
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		var data APIResp
		json.Unmarshal(body, &data)
		if len(data.Pois) == 0 {
			break
		}
		allPois = append(allPois, data.Pois...)
		fmt.Printf("已抓取第%d页，累计%d条\n", page, len(allPois))
		if len(data.Pois) < 25 {
			break
		}
		page++
		time.Sleep(500 * time.Millisecond)
	}

	// 去重
	unique := make(map[string]POI)
	for _, poi := range allPois {
		unique[poi.ID] = poi
	}
	result := make([]POI, 0, len(unique))
	for _, poi := range unique {
		result = append(result, poi)
	}
	b, _ := json.MarshalIndent(result, "", "  ")
	ioutil.WriteFile(outputFile, b, 0644)
	fmt.Printf("最终去重后共%d条，已保存到 %s\n", len(result), outputFile)
	fmt.Println("前10条POI示例:")
	for i, poi := range result {
		if i >= 10 {
			break
		}
		fmt.Printf("%d. %s | %s | %s\n", i+1, poi.Name, poi.Address, poi.ID)
	}
}
