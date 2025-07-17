package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// Google Maps API 响应结构
type GoogleMapsResponse struct {
	Results []GoogleMapsResult `json:"results"`
	Status  string             `json:"status"`
}

type GoogleMapsResult struct {
	PlaceID       string                 `json:"place_id"`
	Name          string                 `json:"name"`
	FormattedAddress string              `json:"formatted_address"`
	Geometry      GoogleMapsGeometry     `json:"geometry"`
	Types         []string               `json:"types"`
	Rating        float64                `json:"rating,omitempty"`
	UserRatingsTotal int                 `json:"user_ratings_total,omitempty"`
	OpeningHours  GoogleMapsOpeningHours `json:"opening_hours,omitempty"`
	Photos        []GoogleMapsPhoto      `json:"photos,omitempty"`
	Vicinity      string                 `json:"vicinity,omitempty"`
}

type GoogleMapsGeometry struct {
	Location GoogleMapsLocation `json:"location"`
}

type GoogleMapsLocation struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type GoogleMapsOpeningHours struct {
	OpenNow bool `json:"open_now"`
}

type GoogleMapsPhoto struct {
	PhotoReference string `json:"photo_reference"`
	Height         int    `json:"height"`
	Width          int    `json:"width"`
}

// 爬虫配置
type SpiderConfig struct {
	GoogleMapsAPIKey string
	SearchRadius     int
	MaxResults       int
	DelayBetweenRequests time.Duration
}

// 爬虫实例
type HospitalSpider struct {
	config SpiderConfig
	client *http.Client
}

// 创建新的爬虫实例
func NewHospitalSpider() *HospitalSpider {
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	return &HospitalSpider{
		config: SpiderConfig{
			GoogleMapsAPIKey:    apiKey,
			SearchRadius:        10000, // 10km
			MaxResults:          20,
			DelayBetweenRequests: time.Second * 2,
		},
		client: &http.Client{
			Timeout: time.Second * 30,
		},
	}
}

// 搜索医院
func (s *HospitalSpider) SearchHospitals(lat, lng float64, radius int) ([]Hospital, error) {
	var allHospitals []Hospital
	
	// 1KM 步进搜索
	for distance := 1; distance <= radius/1000; distance++ {
		hospitals, err := s.searchHospitalsInRadius(lat, lng, distance*1000)
		if err != nil {
			log.Printf("Error searching hospitals at distance %dkm: %v", distance, err)
			continue
		}
		
		allHospitals = append(allHospitals, hospitals...)
		
		// 如果找到足够多的医院，提前结束
		if len(allHospitals) >= s.config.MaxResults {
			break
		}
		
		// 延迟避免请求过快
		time.Sleep(s.config.DelayBetweenRequests)
	}
	
	return allHospitals, nil
}

// 在指定半径内搜索医院
func (s *HospitalSpider) searchHospitalsInRadius(lat, lng float64, radius int) ([]Hospital, error) {
	baseURL := "https://maps.googleapis.com/maps/api/place/nearbysearch/json"
	
	params := url.Values{}
	params.Set("location", fmt.Sprintf("%f,%f", lat, lng))
	params.Set("radius", strconv.Itoa(radius))
	params.Set("type", "hospital")
	params.Set("key", s.config.GoogleMapsAPIKey)
	
	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	
	resp, err := s.client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}
	
	var mapsResp GoogleMapsResponse
	if err := json.Unmarshal(body, &mapsResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}
	
	if mapsResp.Status != "OK" && mapsResp.Status != "ZERO_RESULTS" {
		return nil, fmt.Errorf("Google Maps API error: %s", mapsResp.Status)
	}
	
	var hospitals []Hospital
	for _, result := range mapsResp.Results {
		hospital := s.convertToHospital(result, lat, lng)
		hospitals = append(hospitals, hospital)
	}
	
	return hospitals, nil
}

// 转换 Google Maps 结果为医院对象
func (s *HospitalSpider) convertToHospital(result GoogleMapsResult, searchLat, searchLng float64) Hospital {
	// 获取详细信息
	details := s.getPlaceDetails(result.PlaceID)
	
	hospital := Hospital{
		Name:         result.Name,
		Address:      result.FormattedAddress,
		Latitude:     result.Geometry.Location.Lat,
		Longitude:    result.Geometry.Location.Lng,
		Phone:        details.Phone,
		HospitalType: s.determineHospitalType(result.Types),
		BusinessHours: s.formatBusinessHours(details.OpeningHours),
		Qualifications: s.getQualifications(result.PlaceID),
		CreatedAt:    time.Now().Format("2006-01-02 15:04:05"),
		UpdatedAt:    time.Now().Format("2006-01-02 15:04:05"),
	}
	
	// 计算距离
	hospital.Distance = calculateDistance(searchLat, searchLng, hospital.Latitude, hospital.Longitude)
	
	// 设置评分
	if result.Rating > 0 {
		hospital.Rating = result.Rating
		hospital.Confidence = s.calculateConfidence(result.UserRatingsTotal)
	}
	
	return hospital
}

// 获取地点详细信息
func (s *HospitalSpider) getPlaceDetails(placeID string) PlaceDetails {
	baseURL := "https://maps.googleapis.com/maps/api/place/details/json"
	
	params := url.Values{}
	params.Set("place_id", placeID)
	params.Set("fields", "formatted_phone_number,opening_hours,website")
	params.Set("key", s.config.GoogleMapsAPIKey)
	
	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	
	resp, err := s.client.Get(reqURL)
	if err != nil {
		log.Printf("Error getting place details: %v", err)
		return PlaceDetails{}
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading place details response: %v", err)
		return PlaceDetails{}
	}
	
	var detailsResp PlaceDetailsResponse
	if err := json.Unmarshal(body, &detailsResp); err != nil {
		log.Printf("Error parsing place details: %v", err)
		return PlaceDetails{}
	}
	
	return detailsResp.Result
}

// 确定医院类型
func (s *HospitalSpider) determineHospitalType(types []string) string {
	for _, t := range types {
		switch t {
		case "hospital":
			return "综合医院"
		case "health":
			return "医疗机构"
		case "establishment":
			return "医疗机构"
		}
	}
	return "医院"
}

// 格式化营业时间
func (s *HospitalSpider) formatBusinessHours(hours GoogleMapsOpeningHours) string {
	if hours.OpenNow {
		return "24小时营业"
	}
	return "营业时间请咨询"
}

// 获取资质信息（模拟）
func (s *HospitalSpider) getQualifications(placeID string) string {
	// 这里可以接入真实的资质查询API
	// 目前返回模拟数据
	return "ISO认证,卫生部认证"
}

// 计算置信度
func (s *HospitalSpider) calculateConfidence(ratingCount int) float64 {
	if ratingCount == 0 {
		return 0.0
	}
	
	// 评分数量越多，置信度越高
	confidence := float64(ratingCount) / 1000.0
	if confidence > 1.0 {
		confidence = 1.0
	}
	
	return confidence
}

// 保存医院数据到数据库
func (s *HospitalSpider) SaveHospitals(hospitals []Hospital) error {
	for _, hospital := range hospitals {
		// 检查是否已存在
		var existingID int
		err := db.QueryRow("SELECT id FROM hospitals WHERE name = ? AND address = ?", 
			hospital.Name, hospital.Address).Scan(&existingID)
		
		if err == nil {
			// 更新现有记录
			_, err = db.Exec(`
				UPDATE hospitals 
				SET phone = ?, hospital_type = ?, business_hours = ?, qualifications = ?, updated_at = ?
				WHERE id = ?
			`, hospital.Phone, hospital.HospitalType, hospital.BusinessHours, hospital.Qualifications, 
			   time.Now().Format("2006-01-02 15:04:05"), existingID)
		} else {
			// 插入新记录
			_, err = db.Exec(`
				INSERT INTO hospitals (name, address, latitude, longitude, phone, hospital_type, main_departments, business_hours, qualifications)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, hospital.Name, hospital.Address, hospital.Latitude, hospital.Longitude, 
			   hospital.Phone, hospital.HospitalType, hospital.MainDepartments, hospital.BusinessHours, hospital.Qualifications)
		}
		
		if err != nil {
			log.Printf("Error saving hospital %s: %v", hospital.Name, err)
		}
	}
	
	return nil
}

// Place Details API 响应结构
type PlaceDetailsResponse struct {
	Result PlaceDetails `json:"result"`
	Status string       `json:"status"`
}

type PlaceDetails struct {
	Phone        string                    `json:"formatted_phone_number"`
	OpeningHours GoogleMapsOpeningHours    `json:"opening_hours"`
	Website      string                    `json:"website"`
}

// 爬取评级数据
func (s *HospitalSpider) CrawlRatings(hospitalID int, hospitalName string) error {
	// 这里可以接入真实的评级数据源
	// 目前使用模拟数据
	
	ratings := []Rating{
		{
			HospitalID:  hospitalID,
			Source:      "官方评级",
			RatingValue: 4.5 + float64(hospitalID%5)*0.1, // 模拟评分
			Confidence:  0.9,
			RatingDate:  time.Now().Format("2006-01-02"),
		},
		{
			HospitalID:  hospitalID,
			Source:      "用户评分",
			RatingValue: 4.0 + float64(hospitalID%5)*0.2,
			Confidence:  0.7,
			RatingDate:  time.Now().Format("2006-01-02"),
		},
	}
	
	for _, rating := range ratings {
		_, err := db.Exec(`
			INSERT INTO ratings (hospital_id, source, rating_value, confidence, rating_date)
			VALUES (?, ?, ?, ?, ?)
		`, rating.HospitalID, rating.Source, rating.RatingValue, rating.Confidence, rating.RatingDate)
		
		if err != nil {
			log.Printf("Error saving rating: %v", err)
		}
	}
	
	return nil
}

// 爬取评论数据
func (s *HospitalSpider) CrawlReviews(hospitalID int, hospitalName string) error {
	// 模拟评论数据
	reviews := []Review{
		{
			HospitalID:     hospitalID,
			Source:         "Google Maps",
			UserName:       "匿名用户",
			Rating:         4.5,
			ReviewText:     "服务很好，医生专业，环境干净",
			ReviewDate:     time.Now().AddDate(0, 0, -5).Format("2006-01-02"),
			SentimentScore: 0.8,
		},
		{
			HospitalID:     hospitalID,
			Source:         "Google Maps",
			UserName:       "匿名用户",
			Rating:         4.0,
			ReviewText:     "整体不错，等待时间稍长",
			ReviewDate:     time.Now().AddDate(0, 0, -10).Format("2006-01-02"),
			SentimentScore: 0.6,
		},
	}
	
	for _, review := range reviews {
		_, err := db.Exec(`
			INSERT INTO reviews (hospital_id, source, user_name, rating, review_text, review_date, sentiment_score)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, review.HospitalID, review.Source, review.UserName, review.Rating, 
		   review.ReviewText, review.ReviewDate, review.SentimentScore)
		
		if err != nil {
			log.Printf("Error saving review: %v", err)
		}
	}
	
	return nil
} 