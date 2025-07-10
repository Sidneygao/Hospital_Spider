package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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
	fmt.Println("Starting Hospital Spider Backend...")

	// 初始化数据库
	initDB()
	defer db.Close()

	fmt.Println("Database initialized successfully")

	// 创建 Gin 路由
	r := gin.Default()

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
	}

	// 启动服务器
	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("Press Ctrl+C to stop the server")
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

	// 插入示例医院数据
	hospitals := []Hospital{
		{
			Name:            "曼谷医院",
			Address:         "2 Soi Soonvijai 7, New Petchburi Rd, Bangkok 10310",
			Latitude:        13.7563,
			Longitude:       100.5018,
			Phone:           "+66 2 310 3000",
			HospitalType:    "综合医院",
			MainDepartments: "内科,外科,妇产科,儿科",
			BusinessHours:   "24小时营业",
			Qualifications:  "JCI认证,ISO认证",
		},
		{
			Name:            "康民国际医院",
			Address:         "33 Sukhumvit 3, Khlong Toei Nuea, Watthana, Bangkok 10110",
			Latitude:        13.7383,
			Longitude:       100.5608,
			Phone:           "+66 2 066 8888",
			HospitalType:    "综合医院",
			MainDepartments: "心脏科,神经科,肿瘤科",
			BusinessHours:   "24小时营业",
			Qualifications:  "JCI认证,泰国卫生部认证",
		},
		{
			Name:            "三美泰医院",
			Address:         "488 Silom Road, Bangkok 10500",
			Latitude:        13.7246,
			Longitude:       100.5276,
			Phone:           "+66 2 711 8000",
			HospitalType:    "专科医院",
			MainDepartments: "妇产科,儿科,整形外科",
			BusinessHours:   "周一至周日 8:00-20:00",
			Qualifications:  "ISO认证,泰国卫生部认证",
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

	// 插入示例评分数据
	ratings := []Rating{
		{HospitalID: 1, Source: "官方评级", RatingValue: 4.8, Confidence: 0.95},
		{HospitalID: 1, Source: "用户评分", RatingValue: 4.5, Confidence: 0.85},
		{HospitalID: 2, Source: "官方评级", RatingValue: 4.9, Confidence: 0.98},
		{HospitalID: 2, Source: "用户评分", RatingValue: 4.7, Confidence: 0.90},
		{HospitalID: 3, Source: "官方评级", RatingValue: 4.6, Confidence: 0.92},
		{HospitalID: 3, Source: "用户评分", RatingValue: 4.3, Confidence: 0.80},
	}

	for _, rating := range ratings {
		_, err := db.Exec(`
			INSERT INTO ratings (hospital_id, source, rating_value, confidence)
			VALUES (?, ?, ?, ?)
		`, rating.HospitalID, rating.Source, rating.RatingValue, rating.Confidence)

		if err != nil {
			log.Printf("Error inserting rating: %v", err)
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
 