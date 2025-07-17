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

func main() {
	// 自动加载.env文件
	_ = godotenv.Load(".env")
	fmt.Println("AMAP_KEY from env:", os.Getenv("AMAP_KEY"))

	// 启动时健康检查AMAP_KEY
	checkAmapKeyHealth()

	// 初始化数据库
	initDB()
	defer db.Close()

	fmt.Println("Database initialized successfully")

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

		// 用户反馈 API
		api.POST("/hospitals/:id/feedback", submitFeedback)
		api.GET("/places/hospitals", getNearbyHospitals)
	}

	// 启动服务器
	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("Press Ctrl+C to stop the server")
	r.GET("/api/amap/geo", AmapGeoProxy)
	r.GET("/api/amap/around", AmapAroundProxy)
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
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
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
	lat := 13.7563 // 曼谷默认坐标
	lng := 100.5018
	_ = 10.0 // radius 暂时未使用
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
			_ = r // radius 暂时未使用
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	// 简化版距离计算（实际项目中应使用更精确的地理计算）
	rows, err := db.Query(`
		SELECT id, name, address, latitude, longitude, phone, hospital_type, main_departments, business_hours, qualifications, created_at, updated_at
		FROM hospitals
		LIMIT ?
	`, limit)

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

		// 计算距离（简化版）
		h.Distance = calculateDistance(lat, lng, h.Latitude, h.Longitude)

		// 获取平均评分
		h.Rating, h.Confidence = getHospitalRating(h.ID)

		hospitals = append(hospitals, h)
	}

	// 按距离排序
	// 这里可以添加更复杂的排序算法

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
	key := os.Getenv("AMAP_KEY")
	if key == "" {
		log.Println("[AmapGeoProxy] AMAP_KEY not set in backend env")
		c.JSON(500, gin.H{"error": "AMAP_KEY not set in backend env"})
		return
	}
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
		log.Printf("[AmapAroundProxy][参数异常] location参数缺失或格式错误: %s", location)
		c.JSON(400, gin.H{"error": "location参数缺失或格式错误"})
		return
	}
	radius := c.DefaultQuery("radius", "5000")
	types := c.DefaultQuery("types", "090000")
	key := os.Getenv("AMAP_KEY")
	if key == "" {
		log.Println("[AmapAroundProxy] AMAP_KEY not set in backend env")
		c.JSON(500, gin.H{"error": "AMAP_KEY not set in backend env"})
		return
	}
	// 修复：将逗号分隔的types转为竖线分隔
	types = strings.ReplaceAll(types, ",", "|")
	log.Println("[AmapAroundProxy] 参数:", "location=", location, "radius=", radius, "types=", types)
	amapUrl := "https://restapi.amap.com/v3/place/around?key=" + key + "&location=" + url.QueryEscape(location) + "&radius=" + url.QueryEscape(radius)
	if types != "" {
		amapUrl += "&types=" + url.QueryEscape(types)
	}
	log.Println("[AmapAroundProxy] 请求URL:", amapUrl)
	resp, err := http.Get(amapUrl)
	if err != nil {
		log.Println("[AmapAroundProxy] amap request failed:", err)
		c.JSON(500, gin.H{"error": "amap request failed", "detail": err.Error()})
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[AmapAroundProxy] 读取响应体失败: %v", err)
		c.JSON(500, gin.H{"error": "amap response body read failed", "detail": err.Error()})
		return
	}
	log.Printf("[AmapAroundProxy] HTTP状态码: %d, 响应体: %s", resp.StatusCode, string(body))
	// 自动解析高德API响应码
	var amapResp map[string]interface{}
	json.Unmarshal(body, &amapResp)
	if status, ok := amapResp["status"]; ok && status != "1" {
		log.Printf("[AmapAroundProxy][高德API异常] status: %v, info: %v, infocode: %v, body: %s", status, amapResp["info"], amapResp["infocode"], string(body))
	}
	if resp.StatusCode != 200 {
		c.JSON(500, gin.H{"error": "amap api status not 200", "status": resp.StatusCode, "body": string(body)})
		return
	}
	log.Println("[AmapAroundProxy] 高德原始响应:", string(body))
	c.DataFromReader(resp.StatusCode, int64(len(body)), resp.Header.Get("Content-Type"), io.NopCloser(bytes.NewReader(body)), nil)
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
