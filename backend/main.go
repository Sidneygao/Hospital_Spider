package main

import (
	"bytes" // 用于将body重新包装为io.Reader
	"database/sql"
	"fmt"
	"io"  // 用于读取HTTP响应体
	"log" // 用于输出日志
	"net/http"
	"strconv"
	"strings"
	"time"

	"encoding/json"
	"io/ioutil"
	"os"

	"net/url"

	"math"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

// 数据模型
type Hospital struct {
	ID              int     `json:"id" db:"id"`
	Name            string  `json:"name" db:"name"`
	Address         string  `json:"address" db:"address"`
	Latitude        float64 `json:"latitude" db:"latitude"`
	Longitude       float64 `json:"longitude" db:"longitude"`
	Phone           string  `json:"phone" db:"phone"`
	HospitalType    string  `json:"hospital_type" db:"hospital_type"`
	MainDepartments string  `json:"main_departments" db:"main_departments"`
	BusinessHours   string  `json:"business_hours" db:"business_hours"`
	Qualifications  string  `json:"qualifications" db:"qualifications"`
	CreatedAt       string  `json:"created_at" db:"created_at"`
	UpdatedAt       string  `json:"updated_at" db:"updated_at"`
	Distance        float64 `json:"distance,omitempty"`
	Rating          float64 `json:"rating,omitempty"`
	Confidence      float64 `json:"confidence,omitempty"`
}

type Rating struct {
	ID          int     `json:"id" db:"id"`
	HospitalID  int     `json:"hospital_id" db:"hospital_id"`
	Source      string  `json:"source" db:"source"`
	RatingValue float64 `json:"rating_value" db:"rating_value"`
	Confidence  float64 `json:"confidence" db:"confidence"`
	RatingDate  string  `json:"rating_date" db:"rating_date"`
	CreatedAt   string  `json:"created_at" db:"created_at"`
}

type Review struct {
	ID             int     `json:"id" db:"id"`
	HospitalID     int     `json:"hospital_id" db:"hospital_id"`
	Source         string  `json:"source" db:"source"`
	UserName       string  `json:"user_name" db:"user_name"`
	Rating         float64 `json:"rating" db:"rating"`
	ReviewText     string  `json:"review_text" db:"review_text"`
	ReviewDate     string  `json:"review_date" db:"review_date"`
	SentimentScore float64 `json:"sentiment_score" db:"sentiment_score"`
	CreatedAt      string  `json:"created_at" db:"created_at"`
}

type UserFeedback struct {
	ID         int     `json:"id" db:"id"`
	HospitalID int     `json:"hospital_id" db:"hospital_id"`
	UserIP     string  `json:"user_ip" db:"user_ip"`
	Rating     float64 `json:"rating" db:"rating"`
	Comment    string  `json:"comment" db:"comment"`
	CreatedAt  string  `json:"created_at" db:"created_at"`
}

type SearchResponse struct {
	Status string     `json:"status"`
	Count  int        `json:"count"`
	Data   []Hospital `json:"data"`
}

type DetailResponse struct {
	Status string   `json:"status"`
	Data   Hospital `json:"data"`
}

type FeedbackRequest struct {
	Rating  float64 `json:"rating" binding:"required"`
	Comment string  `json:"comment"`
}

type FeedbackResponse struct {
	Status  string       `json:"status"`
	Message string       `json:"message"`
	Data    UserFeedback `json:"data"`
}

var db *sql.DB

var staticTier3POIs []map[string]interface{}
var localGeocodeCache map[string]map[string]interface{}

func loadStaticTier3POIs() {
	if staticTier3POIs != nil {
		return
	}
	data, err := ioutil.ReadFile("beijing_tier3_hospitals_by_keyword_go.json")
	if err != nil {
		log.Printf("[三甲名单] 加载失败: %v", err)
		staticTier3POIs = []map[string]interface{}{}
		return
	}
	json.Unmarshal(data, &staticTier3POIs)
	log.Printf("[三甲名单] 已加载 %d 条三甲POI", len(staticTier3POIs))
}

// 初始化本地地理编码缓存
func initLocalGeocodeCache() {
	localGeocodeCache = make(map[string]map[string]interface{})
	
	// 预加载一些常见地址的地理编码
	commonAddresses := map[string]map[string]interface{}{
		"北京市朝阳区建国门外大街1号": {
			"location": "116.4074,39.9042",
			"formatted_address": "北京市朝阳区建国门外大街1号",
			"province": "北京市",
			"city": "北京市",
			"district": "朝阳区",
		},
		"北京市海淀区中关村大街1号": {
			"location": "116.3074,39.9842",
			"formatted_address": "北京市海淀区中关村大街1号",
			"province": "北京市",
			"city": "北京市",
			"district": "海淀区",
		},
		"上海市浦东新区陆家嘴环路1000号": {
			"location": "121.4737,31.2304",
			"formatted_address": "上海市浦东新区陆家嘴环路1000号",
			"province": "上海市",
			"city": "上海市",
			"district": "浦东新区",
		},
		"广州市天河区珠江新城花城大道85号": {
			"location": "113.2644,23.1291",
			"formatted_address": "广州市天河区珠江新城花城大道85号",
			"province": "广东省",
			"city": "广州市",
			"district": "天河区",
		},
		"深圳市南山区深南大道10000号": {
			"location": "114.0579,22.5431",
			"formatted_address": "深圳市南山区深南大道10000号",
			"province": "广东省",
			"city": "深圳市",
			"district": "南山区",
		},
	}
	
	for address, geocode := range commonAddresses {
		localGeocodeCache[address] = geocode
	}
	
	log.Printf("[本地地理编码] 已加载 %d 个常见地址", len(localGeocodeCache))
}

// 本地地理编码函数
func localGeocode(address string) (map[string]interface{}, bool) {
	if localGeocodeCache == nil {
		initLocalGeocodeCache()
	}
	
	// 精确匹配
	if geocode, exists := localGeocodeCache[address]; exists {
		return geocode, true
	}
	
	// 模糊匹配（包含关键词）
	for cachedAddress, geocode := range localGeocodeCache {
		if strings.Contains(address, cachedAddress) || strings.Contains(cachedAddress, address) {
			return geocode, true
		}
	}
	
	return nil, false
}

func isInRadius(poiLocation string, centerLng, centerLat, radius float64) bool {
	parts := strings.Split(poiLocation, ",")
	if len(parts) != 2 {
		return false
	}
	lng, _ := strconv.ParseFloat(parts[0], 64)
	lat, _ := strconv.ParseFloat(parts[1], 64)
	return haversine(lng, lat, centerLng, centerLat) < radius
}

func main() {
	// 自动加载.env文件
	_ = godotenv.Load(".env")
	fmt.Println("AMAP_KEY from env:", os.Getenv("AMAP_KEY"))

	// 启动时健康检查AMAP_KEY
	checkAmapKeyHealth()

	// 初始化本地缓存
	initLocalGeocodeCache()
	loadStaticTier3POIs()

	// 初始化数据库
	initDB()
	defer db.Close()

	fmt.Println("Database initialized successfully")
	fmt.Println("Local cache initialized successfully")

	// 创建 Gin 路由
	r := gin.Default()

	// 全局recover，捕获所有panic
	r.Use(func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[全局Panic捕获] %+v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			}
		}()
		c.Next()
	})

	// 配置 CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(config))

	// API 路由
	api := r.Group("/api")
	{
		// 医院搜索 API
		api.GET("/hospitals", getHospitals)
		api.GET("/hospitals/search", searchHospitals)
		api.GET("/hospitals/:id", getHospitalDetail)

		// 医院评级 API
		api.GET("/hospitals/:id/ratings", getHospitalRatingsAPI)
		api.GET("/hospitals/:id/reviews", getHospitalReviews)
		api.GET("/hospitals/:id/feedback", getHospitalFeedback)

		// 用户反馈 API
		api.POST("/hospitals/:id/feedback", submitFeedback)
		api.GET("/places/hospitals", getNearbyHospitals)
	}

	// 启动服务器
	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("Press Ctrl+C to stop the server")
	r.GET("/api/amap/geo", AmapGeoProxy)
	r.GET("/api/amap/around", AmapAroundProxy)

	// 新增：合并POI结果API
	r.GET("/api/merged-pois", getMergedPois)

	r.Run(":8080")
}

// 初始化数据库
func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./hospital_spider.db")
	if err != nil {
		log.Fatal(err)
	}

	// 创建表
	createTables()

	// 插入示例数据
	insertSampleData()
}

// 创建数据库表
func createTables() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS hospitals (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			address TEXT NOT NULL,
			latitude REAL NOT NULL,
			longitude REAL NOT NULL,
			phone TEXT,
			hospital_type TEXT,
			main_departments TEXT,
			business_hours TEXT,
			qualifications TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS ratings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			hospital_id INTEGER NOT NULL,
			source TEXT NOT NULL,
			rating_value REAL NOT NULL,
			confidence REAL DEFAULT 0.5,
			rating_date DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (hospital_id) REFERENCES hospitals(id)
		)`,
		`CREATE TABLE IF NOT EXISTS reviews (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			hospital_id INTEGER NOT NULL,
			source TEXT NOT NULL,
			user_name TEXT,
			rating REAL,
			review_text TEXT,
			review_date DATETIME,
			sentiment_score REAL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (hospital_id) REFERENCES hospitals(id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_feedback (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			hospital_id INTEGER NOT NULL,
			user_ip TEXT,
			rating REAL,
			comment TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (hospital_id) REFERENCES hospitals(id)
		)`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			log.Printf("Error creating table: %v", err)
		}
	}
}

// 插入示例数据
func insertSampleData() {
	// 检查是否已有数据
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM hospitals").Scan(&count)
	if err != nil {
		log.Printf("Error checking data: %v", err)
		return
	}

	if count > 0 {
		return // 已有数据，不重复插入
	}

	// 只插入北京大学第三医院
	hospitals := []Hospital{
		{
			Name:            "北京大学第三医院",
			Address:         "北京市海淀区花园北路49号",
			Latitude:        39.9753,
			Longitude:       116.3541,
			Phone:           "010-82266699",
			HospitalType:    "综合性三级甲等",
			MainDepartments: "综合医疗、教学、科研",
			BusinessHours:   "24小时营业",
			Qualifications:  "三级甲等, 北京大学直属, 综合医院",
		},
	}

	for _, hospital := range hospitals {
		_, err := db.Exec(`
			INSERT INTO hospitals (name, address, latitude, longitude, phone, hospital_type, main_departments, business_hours, qualifications)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, hospital.Name, hospital.Address, hospital.Latitude, hospital.Longitude, hospital.Phone, hospital.HospitalType, hospital.MainDepartments, hospital.BusinessHours, hospital.Qualifications)

		if err != nil {
			log.Printf("Error inserting hospital: %v", err)
		}
	}
}

// 获取医院列表
func getHospitals(c *gin.Context) {
	limit := 10
	skip := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if skipStr := c.Query("skip"); skipStr != "" {
		if s, err := strconv.Atoi(skipStr); err == nil {
			skip = s
		}
	}

	rows, err := db.Query(`
		SELECT id, name, address, latitude, longitude, phone, hospital_type, main_departments, business_hours, qualifications, created_at, updated_at
		FROM hospitals
		LIMIT ? OFFSET ?
	`, limit, skip)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var hospitals []Hospital
	for rows.Next() {
		var h Hospital
		err := rows.Scan(&h.ID, &h.Name, &h.Address, &h.Latitude, &h.Longitude, &h.Phone, &h.HospitalType, &h.MainDepartments, &h.BusinessHours, &h.Qualifications, &h.CreatedAt, &h.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning hospital: %v", err)
			continue
		}
		hospitals = append(hospitals, h)
	}

	response := SearchResponse{
		Status: "success",
		Count:  len(hospitals),
		Data:   hospitals,
	}

	c.JSON(http.StatusOK, response)
}

// 搜索医院
func searchHospitals(c *gin.Context) {
	latStr := c.Query("lat")
	lngStr := c.Query("lng")
	radiusStr := c.Query("radius")
	limitStr := c.Query("limit")
	_ = c.Query("landmark") // 暂时未使用

	// 默认参数
	lat := 39.9042 // 北京默认坐标
	lng := 116.4074
	radius := 10.0
	limit := 10

	// 解析参数
	if latStr != "" {
		if l, err := strconv.ParseFloat(latStr, 64); err == nil {
			lat = l
		}
	}
	if lngStr != "" {
		if l, err := strconv.ParseFloat(lngStr, 64); err == nil {
			lng = l
		}
	}
	if radiusStr != "" {
		if r, err := strconv.ParseFloat(radiusStr, 64); err == nil {
			radius = r
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	// 优先使用本地JSON数据
	loadStaticTier3POIs()
	
	var hospitals []Hospital
	
	// 从本地JSON数据中搜索医院
	if len(staticTier3POIs) > 0 {
		log.Printf("[本地搜索] 使用本地JSON数据，共 %d 条记录", len(staticTier3POIs))
		
		for i, poi := range staticTier3POIs {
			if i >= limit {
				break
			}
			
			// 提取医院信息
			name, _ := poi["name"].(string)
			address, _ := poi["address"].(string)
			location, _ := poi["location"].(string)
			
			// 解析经纬度
			var poiLat, poiLng float64
			if location != "" {
				coords := strings.Split(location, ",")
				if len(coords) == 2 {
					poiLng, _ = strconv.ParseFloat(coords[0], 64)
					poiLat, _ = strconv.ParseFloat(coords[1], 64)
				}
			}
			
			// 计算距离
			distance := calculateDistance(lat, lng, poiLat, poiLng)
			
			// 如果在搜索半径内
			if distance <= radius*1000 { // 转换为米
				hospital := Hospital{
					ID:              i + 1,
					Name:            name,
					Address:         address,
					Latitude:        poiLat,
					Longitude:       poiLng,
					Distance:        distance,
					HospitalType:    "综合医院",
					MainDepartments: "内科,外科,妇产科,儿科",
					BusinessHours:   "24小时",
					Qualifications:  "三级甲等",
					CreatedAt:       time.Now().Format("2006-01-02 15:04:05"),
					UpdatedAt:       time.Now().Format("2006-01-02 15:04:05"),
				}
				
				// 获取评分（模拟数据）
				hospital.Rating = 4.5
				hospital.Confidence = 0.8
				
				hospitals = append(hospitals, hospital)
			}
		}
		
		// 按距离排序
		// 这里可以添加更复杂的排序算法
		
		log.Printf("[本地搜索] 找到 %d 家医院", len(hospitals))
	}
	
	// 如果本地数据不足，从数据库补充
	if len(hospitals) < limit {
		log.Printf("[数据库补充] 从数据库补充数据")
		
		rows, err := db.Query(`
			SELECT id, name, address, latitude, longitude, phone, hospital_type, main_departments, business_hours, qualifications, created_at, updated_at
			FROM hospitals
			LIMIT ?
		`, limit-len(hospitals))

		if err == nil {
			defer rows.Close()
			
			for rows.Next() {
				var h Hospital
				err := rows.Scan(&h.ID, &h.Name, &h.Address, &h.Latitude, &h.Longitude, &h.Phone, &h.HospitalType, &h.MainDepartments, &h.BusinessHours, &h.Qualifications, &h.CreatedAt, &h.UpdatedAt)
				if err != nil {
					log.Printf("Error scanning hospital: %v", err)
					continue
				}

				// 计算距离
				h.Distance = calculateDistance(lat, lng, h.Latitude, h.Longitude)

				// 获取平均评分
				h.Rating, h.Confidence = getHospitalRating(h.ID)

				hospitals = append(hospitals, h)
			}
		}
	}

	response := SearchResponse{
		Status: "success",
		Count:  len(hospitals),
		Data:   hospitals,
	}

	c.JSON(http.StatusOK, response)
}

// 获取医院详情
func getHospitalDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hospital ID"})
		return
	}

	var hospital Hospital
	err = db.QueryRow(`
		SELECT id, name, address, latitude, longitude, phone, hospital_type, main_departments, business_hours, qualifications, created_at, updated_at
		FROM hospitals
		WHERE id = ?
	`, id).Scan(&hospital.ID, &hospital.Name, &hospital.Address, &hospital.Latitude, &hospital.Longitude, &hospital.Phone, &hospital.HospitalType, &hospital.MainDepartments, &hospital.BusinessHours, &hospital.Qualifications, &hospital.CreatedAt, &hospital.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Hospital not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// 获取评分信息
	hospital.Rating, hospital.Confidence = getHospitalRating(hospital.ID)

	response := DetailResponse{
		Status: "success",
		Data:   hospital,
	}

	c.JSON(http.StatusOK, response)
}

// 提交用户反馈
func submitFeedback(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hospital ID"})
		return
	}

	var req FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取用户IP
	userIP := c.ClientIP()

	// 插入反馈
	result, err := db.Exec(`
		INSERT INTO user_feedback (hospital_id, user_ip, rating, comment)
		VALUES (?, ?, ?, ?)
	`, id, userIP, req.Rating, req.Comment)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	feedbackID, _ := result.LastInsertId()

	feedback := UserFeedback{
		ID:         int(feedbackID),
		HospitalID: id,
		UserIP:     userIP,
		Rating:     req.Rating,
		Comment:    req.Comment,
		CreatedAt:  time.Now().Format("2006-01-02 15:04:05"),
	}

	response := FeedbackResponse{
		Status:  "success",
		Message: "反馈已提交",
		Data:    feedback,
	}

	c.JSON(http.StatusOK, response)
}

// 获取附近医院 (代理 Google Places API)
func getNearbyHospitals(c *gin.Context) {
	lat := c.Query("lat")
	lng := c.Query("lng")
	if lat == "" || lng == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lat/lng required"})
		return
	}
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	url := "https://maps.googleapis.com/maps/api/place/nearbysearch/json?location=" + lat + "," + lng + "&radius=5000&type=hospital&key=" + apiKey
	log.Printf("[GoogleAPI] 请求URL: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[GoogleAPI] http.Get错误: %v", err)
		// 兜底SAMPLE
		c.JSON(http.StatusOK, gin.H{
			"results": []gin.H{
				{
					"id":        "sample1",
					"name":      "SAMPLE兜底医院",
					"latitude":  lat,
					"longitude": lng,
					"address":   "SAMPLE地址",
					"_sample":   true,
				},
			},
			"isSample":     true,
			"sampleReason": "Google API请求失败，返回兜底SAMPLE",
		})
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[GoogleAPI] 读取响应体失败: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"results": []gin.H{
				{
					"id":        "sample1",
					"name":      "SAMPLE兜底医院",
					"latitude":  lat,
					"longitude": lng,
					"address":   "SAMPLE地址",
					"_sample":   true,
				},
			},
			"isSample":     true,
			"sampleReason": "Google API响应体读取失败，返回兜底SAMPLE",
		})
		return
	}
	if resp.StatusCode != 200 {
		log.Printf("[GoogleAPI] HTTP状态码: %d, 响应体: %s", resp.StatusCode, string(body))
		c.JSON(http.StatusOK, gin.H{
			"results": []gin.H{
				{
					"id":        "sample1",
					"name":      "SAMPLE兜底医院",
					"latitude":  lat,
					"longitude": lng,
					"address":   "SAMPLE地址",
					"_sample":   true,
				},
			},
			"isSample":     true,
			"sampleReason": "Google API返回非200，返回兜底SAMPLE",
		})
		return
	}
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Printf("[GoogleAPI] JSON解析失败: %v, body: %s", err, string(body))
		c.JSON(http.StatusOK, gin.H{
			"results": []gin.H{
				{
					"id":        "sample1",
					"name":      "SAMPLE兜底医院",
					"latitude":  lat,
					"longitude": lng,
					"address":   "SAMPLE地址",
					"_sample":   true,
				},
			},
			"isSample":     true,
			"sampleReason": "Google API响应JSON解析失败，返回兜底SAMPLE",
		})
		return
	}
	if status, ok := result["status"]; ok && status != "OK" && status != "ZERO_RESULTS" {
		log.Printf("[GoogleAPI] status: %v, error_message: %v, body: %s", status, result["error_message"], string(body))
		c.JSON(http.StatusOK, gin.H{
			"results": []gin.H{
				{
					"id":        "sample1",
					"name":      "SAMPLE兜底医院",
					"latitude":  lat,
					"longitude": lng,
					"address":   "SAMPLE地址",
					"_sample":   true,
				},
			},
			"isSample":     true,
			"sampleReason": "Google API返回错误，返回兜底SAMPLE",
		})
		return
	}
	// 正常返回Google API原始数据
	c.Data(http.StatusOK, "application/json", body)
}

func PlacesSearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "query参数不能为空", http.StatusBadRequest)
		return
	}
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	url := fmt.Sprintf("https://maps.googleapis.com/maps/api/place/textsearch/json?query=%%s&key=%%s", query, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "请求Google Places API失败", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func StaticMapHandler(w http.ResponseWriter, r *http.Request) {
	center := r.URL.Query().Get("center")
	zoom := r.URL.Query().Get("zoom")
	size := r.URL.Query().Get("size")
	if center == "" || zoom == "" || size == "" {
		http.Error(w, "参数不全", http.StatusBadRequest)
		return
	}
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	url := fmt.Sprintf("https://maps.googleapis.com/maps/api/staticmap?center=%%s&zoom=%%s&size=%%s&key=%%s", center, zoom, size, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "请求Google Static Maps API失败", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "image/png")
	io.Copy(w, resp.Body)
}

// 计算两点间距离（简化版，使用欧几里得距离）
func calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	// 这里应该使用 Haversine 公式计算真实地理距离
	// 简化版仅用于演示
	dlat := lat2 - lat1
	dlng := lng2 - lng1
	return (dlat*dlat + dlng*dlng) * 111.0 // 粗略转换为公里
}

// 获取医院评分
func getHospitalRating(hospitalID int) (float64, float64) {
	var avgRating, avgConfidence float64

	err := db.QueryRow(`
		SELECT AVG(rating_value), AVG(confidence)
		FROM ratings
		WHERE hospital_id = ?
	`, hospitalID).Scan(&avgRating, &avgConfidence)

	if err != nil {
		return 0.0, 0.0
	}

	return avgRating, avgConfidence
}

// 高德地理编码代理接口
func AmapGeoProxy(c *gin.Context) {
	log.Printf("[AmapGeoProxy] 收到请求: %s %s, 参数: %v", c.Request.Method, c.Request.URL.String(), c.Request.URL.Query())
	address := c.Query("address")
	if address == "" {
		log.Printf("[AmapGeoProxy][参数异常] address参数缺失")
		c.JSON(400, gin.H{"error": "address参数缺失"})
		return
	}
	
	// 优先使用本地地理编码缓存
	if geocode, found := localGeocode(address); found {
		log.Printf("[AmapGeoProxy] 使用本地缓存: %s", address)
		
		// 构造高德API格式的响应
		response := map[string]interface{}{
			"status": "1",
			"info": "OK",
			"infocode": "10000",
			"count": "1",
			"geocodes": []map[string]interface{}{
				{
					"formatted_address": geocode["formatted_address"],
					"country": "中国",
					"province": geocode["province"],
					"city": geocode["city"],
					"district": geocode["district"],
					"location": geocode["location"],
					"level": "POI",
				},
			},
		}
		
		responseJSON, _ := json.Marshal(response)
		c.Data(http.StatusOK, "application/json", responseJSON)
		return
	}
	
	// 如果本地缓存没有，才调用高德API
	key := os.Getenv("AMAP_KEY")
	if key == "" {
		log.Println("[AmapGeoProxy] AMAP_KEY not set in backend env")
		c.JSON(500, gin.H{"error": "AMAP_KEY not set in backend env"})
		return
	}
	
	log.Printf("[AmapGeoProxy] 本地缓存未找到，调用高德API: %s", address)
	amapUrl := "https://restapi.amap.com/v3/geocode/geo?address=" + url.QueryEscape(address) + "&key=" + key
	log.Println("[AmapGeoProxy] 请求URL:", amapUrl)
	resp, err := http.Get(amapUrl)
	if err != nil {
		log.Println("[AmapGeoProxy] amap request failed:", err)
		c.JSON(500, gin.H{"error": "amap request failed", "detail": err.Error()})
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[AmapGeoProxy] 读取响应体失败: %v", err)
		c.JSON(500, gin.H{"error": "amap response body read failed", "detail": err.Error()})
		return
	}
	log.Printf("[AmapGeoProxy] HTTP状态码: %d, 响应体: %s", resp.StatusCode, string(body))
	// 自动解析高德API响应码
	var amapResp map[string]interface{}
	json.Unmarshal(body, &amapResp)
	if status, ok := amapResp["status"]; ok && status != "1" {
		log.Printf("[AmapGeoProxy][高德API异常] status: %v, info: %v, infocode: %v, body: %s", status, amapResp["info"], amapResp["infocode"], string(body))
	}
	if resp.StatusCode != 200 {
		c.JSON(500, gin.H{"error": "amap api status not 200", "status": resp.StatusCode, "body": string(body)})
		return
	}
	log.Println("[AmapGeoProxy] 高德原始响应:", string(body))
	c.DataFromReader(resp.StatusCode, int64(len(body)), resp.Header.Get("Content-Type"), io.NopCloser(bytes.NewReader(body)), nil)
}

// 高德周边医院搜索代理接口
func AmapAroundProxy(c *gin.Context) {
	log.Printf("[AmapAroundProxy] 收到请求: %s %s, 参数: %v", c.Request.Method, c.Request.URL.String(), c.Request.URL.Query())
	location := c.Query("location")
	if location == "" || !strings.Contains(location, ",") {
		c.JSON(400, gin.H{"error": "location参数缺失或格式错误"})
		return
	}
	radius := c.DefaultQuery("radius", "5000")
	key := os.Getenv("AMAP_KEY")
	if key == "" {
		c.JSON(500, gin.H{"error": "AMAP_KEY not set in backend env"})
		return
	}

	typecodes := []string{"090100", "090101", "090102", "090200", "090300", "090400", "090202"}
	typecodeCategory := map[string]string{
		"090100": "综合医院",
		"090101": "三级甲等医院",
		"090102": "社区医院",
		"090200": "专科医院",
		"090202": "牙科医院",
		"090203": "icon_small_red_cross_normal",
		"090204": "icon_small_red_cross_normal",
		"090205": "icon_small_red_cross_normal",
		"090206": "icon_small_red_cross_normal",
		"090207": "icon_small_red_cross_normal",
		"090208": "icon_small_red_cross_normal",
		"090209": "icon_small_red_cross_normal",
		"090210": "icon_small_red_cross_normal",
		"090211": "icon_small_red_cross_normal",
		"090300": "icon_clinic",
		"090400": "icon_emergency",
	}
	poiMap := make(map[string]map[string]interface{})
	var ledger []RawPOIRecord
	for _, tc := range typecodes {
		url := fmt.Sprintf("https://restapi.amap.com/v3/place/around?key=%s&location=%s&radius=%s&types=%s", key, location, radius, tc)
		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		var amapResp map[string]interface{}
		json.Unmarshal(body, &amapResp)
		pois, _ := amapResp["pois"].([]interface{})
		// 追加到台账
		ledger = append(ledger, RawPOIRecord{Typecode: tc, POIs: pois})
		for _, poi := range pois {
			m, ok := poi.(map[string]interface{})
			if !ok {
				continue
			}
			id, _ := m["id"].(string)
			if cat, ok := typecodeCategory[tc]; ok {
				m["hospital_category"] = cat
			}
			poiMap[id] = m
		}
	}
	// ====== 新增：090100/090101合并预处理 ======
	merge0901xxHospitals(poiMap)
	// 写入amap_query_ledger.json到backend目录，增强健壮性
	log.Println("[台账] 准备写入台账文件 backend/amap_query_ledger.json ...")
	if _, err := os.Stat("backend"); os.IsNotExist(err) {
		errMk := os.Mkdir("backend", 0755)
		if errMk != nil {
			log.Println("[台账] 创建backend目录失败:", errMk)
		}
	}
	ledgerBytes, _ := json.MarshalIndent(ledger, "", "  ")
	errWrite := ioutil.WriteFile("backend/amap_query_ledger.json", ledgerBytes, 0644)
	if errWrite != nil {
		log.Println("[台账] 写台账失败:", errWrite)
	} else {
		log.Println("[台账] 台账写入成功，记录数：", len(ledger))
	}
	// 构建标准化POI缓冲JSON，增加icon_type字段
	var mergedPois []map[string]interface{}
	iconTypeMap := map[string]string{
		"090100": "icon_general_hospital",
		"090101": "icon_tier3_hospital",
		"090102": "icon_health_center",
		"090200": "icon_special_hospital",
		"090202": "icon_tooth",
		"090203": "icon_small_red_cross_normal",
		"090204": "icon_small_red_cross_normal",
		"090205": "icon_small_red_cross_normal",
		"090206": "icon_small_red_cross_normal",
		"090207": "icon_small_red_cross_normal",
		"090208": "icon_small_red_cross_normal",
		"090209": "icon_small_red_cross_normal",
		"090210": "icon_small_red_cross_normal",
		"090211": "icon_small_red_cross_normal",
		"090300": "icon_clinic",
		"090400": "icon_emergency",
	}
	// POI去重合并逻辑
	for _, v := range poiMap {
		m := v
		tc, _ := m["typecode"].(string)
		if icon, ok := iconTypeMap[tc]; ok {
			m["icon_type"] = icon
		} else {
			m["icon_type"] = "icon_default"
		}
		// 牙科医院不去重，直接加入
		if tc == "090202" {
			mergedPois = append(mergedPois, m)
			continue
		}
		// 其余POI去重
		lat1, _ := parseFloatFromAny(m["location_lat"])
		lng1, _ := parseFloatFromAny(m["location_lng"])
		isDuplicate := false
		for _, exist := range mergedPois {
			tcExist, _ := exist["typecode"].(string)
			if tcExist == "090202" {
				continue // 不与牙科比对
			}
			lat2, _ := parseFloatFromAny(exist["location_lat"])
			lng2, _ := parseFloatFromAny(exist["location_lng"])
			if lat1 != 0 && lng1 != 0 && lat2 != 0 && lng2 != 0 {
				dist := calculateDistance(lat1, lng1, lat2, lng2) * 1000 // km->m
				if dist < duplicateDistanceThreshold {
					isDuplicate = true
					break
				}
			}
		}
		if !isDuplicate {
			mergedPois = append(mergedPois, m)
		}
	}
	// POI合并优化：名称重叠>=4字且距离<350米的合并为一个
	var finalPois []map[string]interface{}
	optOutIds := make(map[string]bool)
	for i, poi1 := range mergedPois {
		if optOutIds[poi1["id"].(string)] {
			continue
		}
		tc1, _ := poi1["typecode"].(string)
		childtype1, _ := poi1["childtype"]
		// 改进：更准确的childtype空值判断
		isChildtype1Empty := false
		if childtype1 == nil || childtype1 == "" {
			isChildtype1Empty = true
		} else if arr, ok := childtype1.([]interface{}); ok && len(arr) == 0 {
			isChildtype1Empty = true
		} else if arr, ok := childtype1.([]string); ok && len(arr) == 0 {
			isChildtype1Empty = true
		}

		// 只对满足如下条件的POI参与合并
		if (tc1 == "090101" && isChildtype1Empty) ||
			strings.HasPrefix(tc1, "0901") ||
			strings.HasPrefix(tc1, "0902") ||
			strings.HasPrefix(tc1, "0903") ||
			strings.HasPrefix(tc1, "0904") {
			// 参与合并
		} else {
			finalPois = append(finalPois, poi1)
			continue
		}
		for j := i + 1; j < len(mergedPois); j++ {
			poi2 := mergedPois[j]
			if optOutIds[poi2["id"].(string)] {
				continue
			}
			tc2, _ := poi2["typecode"].(string)
			childtype2, _ := poi2["childtype"]

			// 改进：更准确的childtype空值判断
			isChildtype2Empty := false
			if childtype2 == nil || childtype2 == "" {
				isChildtype2Empty = true
			} else if arr, ok := childtype2.([]interface{}); ok && len(arr) == 0 {
				isChildtype2Empty = true
			} else if arr, ok := childtype2.([]string); ok && len(arr) == 0 {
				isChildtype2Empty = true
			}

			if tc2 == "090101" && !isChildtype2Empty {
				continue // 090101且childtype不为空的，不参与合并
			}

			// 改进名称重叠判定：优先检查完全相同的名称
			name1, _ := poi1["name"].(string)
			name2, _ := poi2["name"].(string)

			// 如果名称完全相同，直接合并
			if name1 == name2 {
				// 距离判定
				lat1, _ := parseFloatFromAny(poi1["location_lat"])
				lng1, _ := parseFloatFromAny(poi1["location_lng"])
				lat2, _ := parseFloatFromAny(poi2["location_lat"])
				lng2, _ := parseFloatFromAny(poi2["location_lng"])
				if lat1 != 0 && lng1 != 0 && lat2 != 0 && lng2 != 0 {
					dist := calculateDistance(lat1, lng1, lat2, lng2) * 1000
					if dist < duplicateDistanceThreshold {
						// 合并为一个新POI，名称为原名称
						merged := make(map[string]interface{})
						for k, v := range poi1 {
							merged[k] = v
						}
						// 修正：typecode/childtype优先保留主POI的值，如无则补全
						if merged["typecode"] == nil || merged["typecode"] == "" {
							merged["typecode"] = poi2["typecode"]
						}
						if merged["childtype"] == nil || merged["childtype"] == "" {
							merged["childtype"] = poi2["childtype"]
						}
						finalPois = append(finalPois, merged)
						optOutIds[poi2["id"].(string)] = true
						goto NextPoi
					}
				}
			} else {
				// 改进的字符重叠判定逻辑：更智能的医院名称匹配
				name1, _ := poi1["name"].(string)
				name2, _ := poi2["name"].(string)

				// 方法1：检查是否包含相同的医院核心名称（如"东直门医院"）
				coreNames := []string{"医院", "门诊", "诊所", "中心", "院区", "分院"}
				hasCoreOverlap := false
				for _, core := range coreNames {
					if strings.Contains(name1, core) && strings.Contains(name2, core) {
						// 提取核心名称前的部分进行比较
						idx1 := strings.Index(name1, core)
						idx2 := strings.Index(name2, core)
						if idx1 > 0 && idx2 > 0 {
							prefix1 := name1[:idx1]
							prefix2 := name2[:idx2]
							// 检查前缀是否有重叠
							if len(prefix1) >= 2 && len(prefix2) >= 2 {
								// 计算前缀的重叠字符数
								overlapCount := 0
								for _, c := range prefix1 {
									if strings.ContainsRune(prefix2, c) {
										overlapCount++
									}
								}
								if overlapCount >= 2 { // 至少2个字符重叠
									hasCoreOverlap = true
									log.Printf("[合并算法] 核心名称匹配: %s 和 %s 通过 %s 匹配 (重叠字符数: %d)", name1, name2, core, overlapCount)
									break
								}
							}
						}
					}
				}

				// 方法2：检查是否包含相同的医院名称片段（如"东直门"）
				if !hasCoreOverlap {
					hospitalNameFragments := []string{"东直门", "协和", "同仁", "天坛", "安贞", "积水潭", "友谊", "宣武", "朝阳", "海淀", "丰台", "石景山", "门头沟", "房山", "通州", "顺义", "昌平", "大兴", "怀柔", "平谷", "密云", "延庆"}
					for _, fragment := range hospitalNameFragments {
						if strings.Contains(name1, fragment) && strings.Contains(name2, fragment) {
							hasCoreOverlap = true
							log.Printf("[合并算法] 医院名称片段匹配: %s 和 %s 通过 %s 匹配", name1, name2, fragment)
							break
						}
					}
				}

				// 方法3：原有的字符重叠判定逻辑（作为兜底）
				n1 := []rune(name1)
				n2 := []rune(name2)
				overlap := ""
				for _, c := range n1 {
					if strings.ContainsRune(string(n2), c) && !strings.ContainsRune(overlap, c) {
						overlap += string(c)
					}
				}
				charOverlapCount := len([]rune(overlap))

				// 合并条件：核心名称重叠 或 字符重叠>=4
				if hasCoreOverlap || charOverlapCount >= 4 {
					log.Printf("[合并算法] 尝试合并: %s 和 %s (核心重叠: %v, 字符重叠: %d)", name1, name2, hasCoreOverlap, charOverlapCount)
					// 距离判定
					lat1, _ := parseFloatFromAny(poi1["location_lat"])
					lng1, _ := parseFloatFromAny(poi1["location_lng"])
					lat2, _ := parseFloatFromAny(poi2["location_lat"])
					lng2, _ := parseFloatFromAny(poi2["location_lng"])
					if lat1 != 0 && lng1 != 0 && lat2 != 0 && lng2 != 0 {
						dist := calculateDistance(lat1, lng1, lat2, lng2) * 1000
						if dist < duplicateDistanceThreshold {
							// 合并为一个新POI，选择更简洁的名称
							merged := make(map[string]interface{})
							for k, v := range poi1 {
								merged[k] = v
							}

							// 选择更简洁的名称作为合并后的名称
							if len(name1) <= len(name2) {
								merged["name"] = name1
							} else {
								merged["name"] = name2
							}

							// 修正：typecode/childtype优先保留主POI的值，如无则补全
							if merged["typecode"] == nil || merged["typecode"] == "" {
								merged["typecode"] = poi2["typecode"]
							}
							if merged["childtype"] == nil || merged["childtype"] == "" {
								merged["childtype"] = poi2["childtype"]
							}

							// 添加合并日志
							log.Printf("[合并算法] 合并医院: %s + %s -> %s (距离: %.1fm)", name1, name2, merged["name"], dist)

							finalPois = append(finalPois, merged)
							optOutIds[poi2["id"].(string)] = true
							goto NextPoi
						} else {
							log.Printf("[合并算法] 距离过远，不合并: %s 和 %s (距离: %.1fm > %.1fm)", name1, name2, dist, duplicateDistanceThreshold)
						}
					} else {
						log.Printf("[合并算法] 坐标无效，不合并: %s 和 %s", name1, name2)
					}
				} else {
					log.Printf("[合并算法] 名称不匹配，不合并: %s 和 %s (核心重叠: %v, 字符重叠: %d)", name1, name2, hasCoreOverlap, charOverlapCount)
				}
			}
		}
		finalPois = append(finalPois, poi1)
	NextPoi:
	}
	// 其它被合并的POI标记OptOut
	for _, poi := range mergedPois {
		if optOutIds[poi["id"].(string)] {
			if poi["tags"] == nil {
				poi["tags"] = []string{"OptOut"}
			} else {
				poi["tags"] = append(poi["tags"].([]string), "OptOut")
			}
		}
	}

	// 新增：将合并后POI及TAG写入JSON文件，便于前端查看
	mergedResult := map[string]interface{}{
		"status": "1",
		"count":  len(finalPois),
		"pois":   finalPois,
	}
	mergedBytes, _ := json.MarshalIndent(mergedResult, "", "  ")
	errMerged := ioutil.WriteFile("backend/cache/merged_poi_result.json", mergedBytes, 0644)
	if errMerged != nil {
		log.Println("[合并结果] 写入合并POI结果JSON失败:", errMerged)
	} else {
		log.Println("[合并结果] 合并POI结果写入backend/cache/merged_poi_result.json成功")
	}

	// 修正医院类别、ICON、排序判定逻辑
	for _, poi := range finalPois {
		cat, icon, order := classifyHospital(poi)
		poi["algo_hospital_category"] = cat
		poi["algo_icon_type"] = icon
		poi["algo_display_order"] = order
	}

	c.JSON(http.StatusOK, mergedResult)
	return
}

// 合并090100/090101医院POI，按名称最长公共子串≥4且距离<300米原则，生成新POI
func merge0901xxHospitals(poiMap map[string]map[string]interface{}) {
	// 1. 收集所有0901xx POI（typecode前4位为0901）
	var poiList []map[string]interface{}
	for _, poi := range poiMap {
		tc, _ := poi["typecode"].(string)
		if len(tc) >= 4 && tc[:4] == "0901" {
			poiList = append(poiList, poi)
		}
	}
	merged := make([]bool, len(poiList))
	for i := 0; i < len(poiList); i++ {
		if merged[i] {
			continue
		}
		name1, _ := poiList[i]["name"].(string)
		lat1, _ := parseFloatFromAny(poiList[i]["location_lat"])
		lng1, _ := parseFloatFromAny(poiList[i]["location_lng"])
		if lat1 == 0 || lng1 == 0 {
			log.Printf("[合并算法] 坐标无效，不合并: %s", name1)
			continue
		}
		group := []int{i}
		for j := i + 1; j < len(poiList); j++ {
			if merged[j] {
				continue
			}
			name2, _ := poiList[j]["name"].(string)
			lat2, _ := parseFloatFromAny(poiList[j]["location_lat"])
			lng2, _ := parseFloatFromAny(poiList[j]["location_lng"])
			if lat2 == 0 || lng2 == 0 {
				log.Printf("[合并算法] 坐标无效，不合并: %s 和 %s", name1, name2)
				continue
			}
			common := lcs(name1, name2)
			dist := calculateDistance(lat1, lng1, lat2, lng2) * 1000
			if len([]rune(common)) >= 4 && dist < 300 {
				log.Printf("[合并算法] LCS和距离均满足，合并: %s 和 %s (LCS: %s, 距离: %.2fm)", name1, name2, common, dist)
				group = append(group, j)
				merged[j] = true
			} else {
				log.Printf("[合并算法] 不合并: %s 和 %s (LCS: %s, 距离: %.2fm)", name1, name2, common, dist)
			}
		}
		if len(group) > 1 {
			// 生成新POI
			commonName := poiList[group[0]]["name"].(string)
			for _, idx := range group[1:] {
				commonName = lcs(commonName, poiList[idx]["name"].(string))
			}
			commonAddr := poiList[group[0]]["address"].(string)
			for _, idx := range group[1:] {
				commonAddr = lcs(commonAddr, poiList[idx]["address"].(string))
			}
			newType := "090100"
			for _, idx := range group {
				if poiList[idx]["typecode"] == "090101" {
					newType = "090101"
					break
				}
			}
			newPOI := map[string]interface{}{
				"name":      commonName,
				"address":   commonAddr,
				"typecode":  newType,
				"childtype": []string{},
				"tel":       "",
			}
			for _, idx := range group {
				poiList[idx]["category"] = nil // 原POI分类置为null
			}
			poiMap[commonName+commonAddr] = newPOI
			log.Printf("[合并算法] 生成新合并POI: %s, 地址: %s, 类别: %s", commonName, commonAddr, newType)
		}
	}
}

// 辅助函数：从interface{}安全解析float64
func parseFloatFromAny(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return f, true
		}
	}
	return 0, false
}

func checkAmapKeyHealth() {
	key := os.Getenv("AMAP_KEY")
	if key == "" {
		log.Println("[健康检查] AMAP_KEY未设置")
		return
	}
	testUrl := "https://restapi.amap.com/v3/geocode/geo?address=北京&key=" + key
	resp, err := http.Get(testUrl)
	if err != nil {
		log.Println("[健康检查] 高德API请求失败:", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	log.Printf("[健康检查] 高德API响应: %s", string(body))
}

// 台账结构体和全局变量

type RawPOIRecord struct {
	Typecode string        `json:"typecode"`
	POIs     []interface{} `json:"pois"`
}

var allRawPois []RawPOIRecord

// 相似医院去重距离阈值由350米
const duplicateDistanceThreshold = 350.0 // 单位：米

// 新增：合并POI结果JSON文件的API接口
func getMergedPois(c *gin.Context) {
	// 1. 直接复用AmapAroundProxy的实时抓取与合并逻辑
	// 默认北京中心点与半径（可根据前端传参扩展）
	location := c.DefaultQuery("location", "116.407387,39.904179")
	radius := c.DefaultQuery("radius", "5000")
	key := os.Getenv("AMAP_KEY")
	if key == "" {
		c.JSON(500, gin.H{"error": "AMAP_KEY not set in backend env"})
		return
	}
	typecodes := []string{
		"090100", // 综合医院
		"090101", // 三级甲等医院
		"090102", // 社区医院
		"090200", // 专科医院
		"090202", // 牙科
		"090203", // 眼科
		"090204", // 耳鼻喉
		"090205", // 胸科
		"090206", // 骨科
		"090207", // 肿瘤
		"090208", // 脑科
		"090209", // 妇科
		"090210", // 精神
		"090211", // 传染病
		"090300", // 诊所
		"090400", // 急救中心
	}
	typecodeCategory := map[string]string{
		"090100": "综合医院",
		"090101": "三级甲等医院",
		"090102": "社区医院",
		"090200": "专科医院",
		"090202": "牙科医院",
		"090203": "眼科医院",
		"090204": "耳鼻喉医院",
		"090205": "胸科医院",
		"090206": "骨科医院",
		"090207": "肿瘤医院",
		"090208": "脑科医院",
		"090209": "妇科医院",
		"090210": "精神医院",
		"090211": "传染病医院",
		"090300": "诊所",
		"090400": "急救中心",
	}
	poiMap := make(map[string]map[string]interface{})
	var ledger []RawPOIRecord
	for _, tc := range typecodes {
		url := fmt.Sprintf("https://restapi.amap.com/v3/place/around?key=%s&location=%s&radius=%s&types=%s", key, location, radius, tc)
		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		var amapResp map[string]interface{}
		json.Unmarshal(body, &amapResp)
		pois, _ := amapResp["pois"].([]interface{})
		// 追加到台账
		ledger = append(ledger, RawPOIRecord{Typecode: tc, POIs: pois})
		for _, poi := range pois {
			m, ok := poi.(map[string]interface{})
			if !ok {
				continue
			}
			id, _ := m["id"].(string)
			if cat, ok := typecodeCategory[tc]; ok {
				m["hospital_category"] = cat
			}
			poiMap[id] = m
		}
	}

	// 写入amap_query_ledger.json到backend目录，增强健壮性
	log.Println("[台账] 准备写入台账文件 backend/cache/amap_query_ledger.json ...")
	if _, err := os.Stat("backend"); os.IsNotExist(err) {
		errMk := os.Mkdir("backend", 0755)
		if errMk != nil {
			log.Println("[台账] 创建backend目录失败:", errMk)
		}
	}
	ledgerBytes, _ := json.MarshalIndent(ledger, "", "  ")
	errWrite := ioutil.WriteFile("backend/cache/amap_query_ledger.json", ledgerBytes, 0644)
	if errWrite != nil {
		log.Println("[台账] 写台账失败:", errWrite)
	} else {
		log.Println("[台账] 台账写入成功，记录数：", len(ledger))
	}

	// 构建标准化POI缓冲JSON，增加icon_type字段
	var mergedPois []map[string]interface{}
	iconTypeMap := map[string]string{
		"090100": "icon_general_hospital",
		"090101": "icon_tier3_hospital",
		"090102": "icon_health_center",
		"090200": "icon_special_hospital",
		"090202": "icon_tooth",
		"090203": "icon_small_red_cross_normal",
		"090204": "icon_small_red_cross_normal",
		"090205": "icon_small_red_cross_normal",
		"090206": "icon_small_red_cross_normal",
		"090207": "icon_small_red_cross_normal",
		"090208": "icon_small_red_cross_normal",
		"090209": "icon_small_red_cross_normal",
		"090210": "icon_small_red_cross_normal",
		"090211": "icon_small_red_cross_normal",
		"090300": "icon_clinic",
		"090400": "icon_emergency",
	}
	// POI去重合并逻辑
	for _, v := range poiMap {
		m := v
		tc, _ := m["typecode"].(string)
		if icon, ok := iconTypeMap[tc]; ok {
			m["icon_type"] = icon
		} else {
			m["icon_type"] = "icon_default"
		}
		// 牙科医院不去重，直接加入
		if tc == "090202" {
			mergedPois = append(mergedPois, m)
			continue
		}
		// 其余POI去重
		lat1, _ := parseFloatFromAny(m["location_lat"])
		lng1, _ := parseFloatFromAny(m["location_lng"])
		isDuplicate := false
		for _, exist := range mergedPois {
			tcExist, _ := exist["typecode"].(string)
			if tcExist == "090202" {
				continue // 不与牙科比对
			}
			lat2, _ := parseFloatFromAny(exist["location_lat"])
			lng2, _ := parseFloatFromAny(exist["location_lng"])
			if lat1 != 0 && lng1 != 0 && lat2 != 0 && lng2 != 0 {
				dist := calculateDistance(lat1, lng1, lat2, lng2) * 1000 // km->m
				if dist < duplicateDistanceThreshold {
					isDuplicate = true
					break
				}
			}
		}
		if !isDuplicate {
			mergedPois = append(mergedPois, m)
		}
	}

	// POI合并优化：名称重叠>=4字且距离<350米的合并为一个
	var finalPois []map[string]interface{}
	optOutIds := make(map[string]bool)
	for i, poi1 := range mergedPois {
		if optOutIds[poi1["id"].(string)] {
			continue
		}
		tc1, _ := poi1["typecode"].(string)
		childtype1, _ := poi1["childtype"]
		// 改进：更准确的childtype空值判断
		isChildtype1Empty := false
		if childtype1 == nil || childtype1 == "" {
			isChildtype1Empty = true
		} else if arr, ok := childtype1.([]interface{}); ok && len(arr) == 0 {
			isChildtype1Empty = true
		} else if arr, ok := childtype1.([]string); ok && len(arr) == 0 {
			isChildtype1Empty = true
		}

		// 只对满足如下条件的POI参与合并
		if (tc1 == "090101" && isChildtype1Empty) ||
			strings.HasPrefix(tc1, "0901") ||
			strings.HasPrefix(tc1, "0902") ||
			strings.HasPrefix(tc1, "0903") ||
			strings.HasPrefix(tc1, "0904") {
			// 参与合并
		} else {
			finalPois = append(finalPois, poi1)
			continue
		}
		for j := i + 1; j < len(mergedPois); j++ {
			poi2 := mergedPois[j]
			if optOutIds[poi2["id"].(string)] {
				continue
			}
			tc2, _ := poi2["typecode"].(string)
			childtype2, _ := poi2["childtype"]

			// 改进：更准确的childtype空值判断
			isChildtype2Empty := false
			if childtype2 == nil || childtype2 == "" {
				isChildtype2Empty = true
			} else if arr, ok := childtype2.([]interface{}); ok && len(arr) == 0 {
				isChildtype2Empty = true
			} else if arr, ok := childtype2.([]string); ok && len(arr) == 0 {
				isChildtype2Empty = true
			}

			if tc2 == "090101" && !isChildtype2Empty {
				continue // 090101且childtype不为空的，不参与合并
			}

			// 改进名称重叠判定：优先检查完全相同的名称
			name1, _ := poi1["name"].(string)
			name2, _ := poi2["name"].(string)

			// 如果名称完全相同，直接合并
			if name1 == name2 {
				// 距离判定
				lat1, _ := parseFloatFromAny(poi1["location_lat"])
				lng1, _ := parseFloatFromAny(poi1["location_lng"])
				lat2, _ := parseFloatFromAny(poi2["location_lat"])
				lng2, _ := parseFloatFromAny(poi2["location_lng"])
				if lat1 != 0 && lng1 != 0 && lat2 != 0 && lng2 != 0 {
					dist := calculateDistance(lat1, lng1, lat2, lng2) * 1000
					if dist < duplicateDistanceThreshold {
						// 合并为一个新POI，名称为原名称
						merged := make(map[string]interface{})
						for k, v := range poi1 {
							merged[k] = v
						}
						// 修正：typecode/childtype优先保留主POI的值，如无则补全
						if merged["typecode"] == nil || merged["typecode"] == "" {
							merged["typecode"] = poi2["typecode"]
						}
						if merged["childtype"] == nil || merged["childtype"] == "" {
							merged["childtype"] = poi2["childtype"]
						}
						finalPois = append(finalPois, merged)
						optOutIds[poi2["id"].(string)] = true
						goto NextPoi
					}
				}
			} else {
				// 改进的字符重叠判定逻辑：更智能的医院名称匹配
				name1, _ := poi1["name"].(string)
				name2, _ := poi2["name"].(string)

				// 方法1：检查是否包含相同的医院核心名称（如"东直门医院"）
				coreNames := []string{"医院", "门诊", "诊所", "中心", "院区", "分院"}
				hasCoreOverlap := false
				for _, core := range coreNames {
					if strings.Contains(name1, core) && strings.Contains(name2, core) {
						// 提取核心名称前的部分进行比较
						idx1 := strings.Index(name1, core)
						idx2 := strings.Index(name2, core)
						if idx1 > 0 && idx2 > 0 {
							prefix1 := name1[:idx1]
							prefix2 := name2[:idx2]
							// 检查前缀是否有重叠
							if len(prefix1) >= 2 && len(prefix2) >= 2 {
								// 计算前缀的重叠字符数
								overlapCount := 0
								for _, c := range prefix1 {
									if strings.ContainsRune(prefix2, c) {
										overlapCount++
									}
								}
								if overlapCount >= 2 { // 至少2个字符重叠
									hasCoreOverlap = true
									log.Printf("[合并算法] 核心名称匹配: %s 和 %s 通过 %s 匹配 (重叠字符数: %d)", name1, name2, core, overlapCount)
									break
								}
							}
						}
					}
				}

				// 方法2：检查是否包含相同的医院名称片段（如"东直门"）
				if !hasCoreOverlap {
					hospitalNameFragments := []string{"东直门", "协和", "同仁", "天坛", "安贞", "积水潭", "友谊", "宣武", "朝阳", "海淀", "丰台", "石景山", "门头沟", "房山", "通州", "顺义", "昌平", "大兴", "怀柔", "平谷", "密云", "延庆"}
					for _, fragment := range hospitalNameFragments {
						if strings.Contains(name1, fragment) && strings.Contains(name2, fragment) {
							hasCoreOverlap = true
							log.Printf("[合并算法] 医院名称片段匹配: %s 和 %s 通过 %s 匹配", name1, name2, fragment)
							break
						}
					}
				}

				// 方法3：原有的字符重叠判定逻辑（作为兜底）
				n1 := []rune(name1)
				n2 := []rune(name2)
				overlap := ""
				for _, c := range n1 {
					if strings.ContainsRune(string(n2), c) && !strings.ContainsRune(overlap, c) {
						overlap += string(c)
					}
				}
				charOverlapCount := len([]rune(overlap))

				// 合并条件：核心名称重叠 或 字符重叠>=4
				if hasCoreOverlap || charOverlapCount >= 4 {
					log.Printf("[合并算法] 尝试合并: %s 和 %s (核心重叠: %v, 字符重叠: %d)", name1, name2, hasCoreOverlap, charOverlapCount)
					// 距离判定
					lat1, _ := parseFloatFromAny(poi1["location_lat"])
					lng1, _ := parseFloatFromAny(poi1["location_lng"])
					lat2, _ := parseFloatFromAny(poi2["location_lat"])
					lng2, _ := parseFloatFromAny(poi2["location_lng"])
					if lat1 != 0 && lng1 != 0 && lat2 != 0 && lng2 != 0 {
						dist := calculateDistance(lat1, lng1, lat2, lng2) * 1000
						if dist < duplicateDistanceThreshold {
							// 合并为一个新POI，选择更简洁的名称
							merged := make(map[string]interface{})
							for k, v := range poi1 {
								merged[k] = v
							}

							// 选择更简洁的名称作为合并后的名称
							if len(name1) <= len(name2) {
								merged["name"] = name1
							} else {
								merged["name"] = name2
							}

							// 修正：typecode/childtype优先保留主POI的值，如无则补全
							if merged["typecode"] == nil || merged["typecode"] == "" {
								merged["typecode"] = poi2["typecode"]
							}
							if merged["childtype"] == nil || merged["childtype"] == "" {
								merged["childtype"] = poi2["childtype"]
							}

							// 添加合并日志
							log.Printf("[合并算法] 合并医院: %s + %s -> %s (距离: %.1fm)", name1, name2, merged["name"], dist)

							finalPois = append(finalPois, merged)
							optOutIds[poi2["id"].(string)] = true
							goto NextPoi
						} else {
							log.Printf("[合并算法] 距离过远，不合并: %s 和 %s (距离: %.1fm > %.1fm)", name1, name2, dist, duplicateDistanceThreshold)
						}
					} else {
						log.Printf("[合并算法] 坐标无效，不合并: %s 和 %s", name1, name2)
					}
				} else {
					log.Printf("[合并算法] 名称不匹配，不合并: %s 和 %s (核心重叠: %v, 字符重叠: %d)", name1, name2, hasCoreOverlap, charOverlapCount)
				}
			}
		}
		finalPois = append(finalPois, poi1)
	NextPoi:
	}
	// 其它被合并的POI标记OptOut
	for _, poi := range mergedPois {
		if optOutIds[poi["id"].(string)] {
			if poi["tags"] == nil {
				poi["tags"] = []string{"OptOut"}
			} else {
				poi["tags"] = append(poi["tags"].([]string), "OptOut")
			}
		}
	}

	// 新增：将合并后POI及TAG写入JSON文件，便于前端查看
	mergedResult := map[string]interface{}{
		"status": "1",
		"count":  len(finalPois),
		"pois":   finalPois,
	}
	mergedBytes, _ := json.MarshalIndent(mergedResult, "", "  ")
	errMerged := ioutil.WriteFile("backend/cache/merged_poi_result.json", mergedBytes, 0644)
	if errMerged != nil {
		log.Println("[合并结果] 写入合并POI结果JSON失败:", errMerged)
	} else {
		log.Println("[合并结果] 合并POI结果写入backend/cache/merged_poi_result.json成功")
	}

	// 修正医院类别、ICON、排序判定逻辑
	for _, poi := range finalPois {
		cat, icon, order := classifyHospital(poi)
		poi["algo_hospital_category"] = cat
		poi["algo_icon_type"] = icon
		poi["algo_display_order"] = order
	}

	c.JSON(http.StatusOK, mergedResult)
	return
}

// 修正医院类别、ICON、排序判定逻辑
func classifyHospital(poi map[string]interface{}) (string, string, int) {
	typecode, _ := poi["typecode"].(string)
	childtype, childtypeExists := poi["childtype"]
	childtypeStr := ""
	if childtypeExists && childtype != nil {
		switch t := childtype.(type) {
		case string:
			childtypeStr = t
		case float64, int:
			childtypeStr = fmt.Sprintf("%v", t)
		default:
			childtypeStr = fmt.Sprintf("%v", t)
		}
	}

	// 处理多个typecode的情况（如：090202|090300），只取前6位作为主要typecode
	primaryTypecode := typecode
	if strings.Contains(typecode, "|") {
		parts := strings.Split(typecode, "|")
		if len(parts) > 0 && len(parts[0]) >= 6 {
			primaryTypecode = parts[0][:6]
			log.Printf("[分类算法] 多typecode处理: %s -> 取前6位: %s", typecode, primaryTypecode)
		}
	}

	// 添加调试日志
	name, _ := poi["name"].(string)
	log.Printf("[分类算法] 医院: %s, 原始typecode: %s, 主要typecode: %s, childtype: %s", name, typecode, primaryTypecode, childtypeStr)

	// 判断childtype是否为空（包括空字符串、null、[]等）
	isChildtypeEmpty := childtypeStr == "" || childtypeStr == "[]" || childtypeStr == "null" || childtypeStr == "0"
	isChildtypeNotEmpty := childtypeStr != "" && childtypeStr != "[]" && childtypeStr != "null" && childtypeStr != "0"

	// 严格按照规则表进行精细化判断，使用主要typecode
	// 1. (POI=090101) and (childtype=[]) → 三甲，大红十字BOLD，排序1
	if primaryTypecode == "090101" && isChildtypeEmpty {
		log.Printf("[分类算法] %s → 三甲 (090101且childtype为空)", name)
		return "三甲", "icon_tier3_hospital_bold", 1
	}
	// 2. (POI=090101) and (childtype<>[]) → null，不显示
	if primaryTypecode == "090101" && isChildtypeNotEmpty {
		log.Printf("[分类算法] %s → null (090101且childtype非空: %s)", name, childtypeStr)
		return "null", "null", 0
	}
	// 3. (POI=090100) and (childtype=[]) → 综合医院，中红十字BOLD，排序2
	if primaryTypecode == "090100" && isChildtypeEmpty {
		log.Printf("[分类算法] %s → 综合医院 (090100且childtype为空)", name)
		return "综合医院", "icon_general_hospital_bold", 2
	}
	// 4. (POI=090100) and (childtype<>[]) → null，不显示
	if primaryTypecode == "090100" && isChildtypeNotEmpty {
		log.Printf("[分类算法] %s → null (090100且childtype非空: %s)", name, childtypeStr)
		return "null", "null", 0
	}
	// 5. POI=090201 → null，不显示
	if primaryTypecode == "090201" {
		log.Printf("[分类算法] %s → null (090201)", name)
		return "null", "null", 0
	}
	// 6. POI=090102 → 社区医院，小红十字BOLD，排序3
	if primaryTypecode == "090102" {
		log.Printf("[分类算法] %s → 社区医院", name)
		return "社区医院", "icon_small_red_cross_bold", 3
	}
	// 7. POI=090200 → 专科，小红十字BOLD，排序4
	if primaryTypecode == "090200" {
		log.Printf("[分类算法] %s → 专科", name)
		return "专科", "icon_small_red_cross_bold", 4
	}
	// 8. POI=090202 → 牙科，Tooth，排序5
	if primaryTypecode == "090202" {
		log.Printf("[分类算法] %s → 牙科", name)
		return "牙科", "icon_tooth", 5
	}
	// 9. POI=090203 → 眼科，小红十字Normal，排序6
	if primaryTypecode == "090203" {
		log.Printf("[分类算法] %s → 眼科", name)
		return "眼科", "icon_small_red_cross_normal", 6
	}
	// 10. POI=090204 → 耳鼻喉，小红十字Normal，排序7
	if primaryTypecode == "090204" {
		log.Printf("[分类算法] %s → 耳鼻喉", name)
		return "耳鼻喉", "icon_small_red_cross_normal", 7
	}
	// 11. POI=090205 → 胸科，小红十字Normal，排序8
	if primaryTypecode == "090205" {
		log.Printf("[分类算法] %s → 胸科", name)
		return "胸科", "icon_small_red_cross_normal", 8
	}
	// 12. POI=090206 → 骨科，小红十字Normal，排序9
	if primaryTypecode == "090206" {
		log.Printf("[分类算法] %s → 骨科", name)
		return "骨科", "icon_small_red_cross_normal", 9
	}
	// 13. POI=090207 → 肿瘤，小红十字Normal，排序10
	if primaryTypecode == "090207" {
		log.Printf("[分类算法] %s → 肿瘤", name)
		return "肿瘤", "icon_small_red_cross_normal", 10
	}
	// 14. POI=090208 → 脑科，小红十字Normal，排序11
	if primaryTypecode == "090208" {
		log.Printf("[分类算法] %s → 脑科", name)
		return "脑科", "icon_small_red_cross_normal", 11
	}
	// 15. POI=090209 → 妇科，小红十字Normal，排序12
	if primaryTypecode == "090209" {
		log.Printf("[分类算法] %s → 妇科", name)
		return "妇科", "icon_small_red_cross_normal", 12
	}
	// 16. POI=090210 → 精神，小红十字Normal，排序13
	if primaryTypecode == "090210" {
		log.Printf("[分类算法] %s → 精神", name)
		return "精神", "icon_small_red_cross_normal", 13
	}
	// 17. POI=090211 → 传染病，小红十字Normal，排序14
	if primaryTypecode == "090211" {
		log.Printf("[分类算法] %s → 传染病", name)
		return "传染病", "icon_small_red_cross_normal", 14
	}
	// 18. POI=090300 → 诊所，小红十字Normal，排序15
	if primaryTypecode == "090300" {
		log.Printf("[分类算法] %s → 诊所", name)
		return "诊所", "icon_small_red_cross_normal", 15
	}
	// 19. POI=090400 → 急救中心，ER.png，排序16
	if primaryTypecode == "090400" {
		log.Printf("[分类算法] %s → 急救中心", name)
		return "急救中心", "icon_er", 16
	}
	// 其它默认
	log.Printf("[分类算法] %s → 其他 (主要typecode: %s)", name, primaryTypecode)
	return "其他", "icon_default", 99
}

// 地址最大交集
func maxCommonAddress(addrs []string) string {
	if len(addrs) == 0 {
		return ""
	}
	res := addrs[0]
	for i := 1; i < len(addrs); i++ {
		res = lcs(res, addrs[i])
		if res == "" {
			break
		}
	}
	return res
}

// 最长公共子串
func lcs(s1, s2 string) string {
	m, n := len(s1), len(s2)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	maxLen, end := 0, 0
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if s1[i-1] == s2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
				if dp[i][j] > maxLen {
					maxLen = dp[i][j]
					end = i
				}
			}
		}
	}
	if maxLen >= 4 {
		return s1[end-maxLen : end]
	}
	return ""
}

// 计算两点经纬度距离（单位：米）
func haversine(lng1, lat1, lng2, lat2 float64) float64 {
	const R = 6371000 // 地球半径，单位：米
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lng2 - lng1) * math.Pi / 180.0
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// 获取医院评级API
func getHospitalRatingsAPI(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hospital ID"})
		return
	}

	// 查询医院评级
	rows, err := db.Query(`
		SELECT id, hospital_id, source, rating_value, confidence, rating_date, created_at
		FROM ratings
		WHERE hospital_id = ?
		ORDER BY created_at DESC
	`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var ratings []Rating
	for rows.Next() {
		var r Rating
		err := rows.Scan(&r.ID, &r.HospitalID, &r.Source, &r.RatingValue, &r.Confidence, &r.RatingDate, &r.CreatedAt)
		if err != nil {
			continue
		}
		ratings = append(ratings, r)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  len(ratings),
		"data":   ratings,
	})
}

// 获取医院评论
func getHospitalReviews(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hospital ID"})
		return
	}

	// 查询医院评论
	rows, err := db.Query(`
		SELECT id, hospital_id, source, user_name, rating, review_text, review_date, sentiment_score, created_at
		FROM reviews
		WHERE hospital_id = ?
		ORDER BY created_at DESC
		LIMIT 50
	`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var r Review
		err := rows.Scan(&r.ID, &r.HospitalID, &r.Source, &r.UserName, &r.Rating, &r.ReviewText, &r.ReviewDate, &r.SentimentScore, &r.CreatedAt)
		if err != nil {
			continue
		}
		reviews = append(reviews, r)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  len(reviews),
		"data":   reviews,
	})
}

// 获取医院用户反馈
func getHospitalFeedback(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hospital ID"})
		return
	}

	// 查询用户反馈
	rows, err := db.Query(`
		SELECT id, hospital_id, user_ip, rating, comment, created_at
		FROM user_feedback
		WHERE hospital_id = ?
		ORDER BY created_at DESC
		LIMIT 50
	`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var feedbacks []UserFeedback
	for rows.Next() {
		var f UserFeedback
		err := rows.Scan(&f.ID, &f.HospitalID, &f.UserIP, &f.Rating, &f.Comment, &f.CreatedAt)
		if err != nil {
			continue
		}
		feedbacks = append(feedbacks, f)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  len(feedbacks),
		"data":   feedbacks,
	})
}
