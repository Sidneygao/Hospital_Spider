package main

import (
	"math"
	"sort"
)

// 地理距离计算 - Haversine 公式
func calculateHaversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371 // 地球半径（公里）

	// 转换为弧度
	lat1Rad := lat1 * math.Pi / 180
	lng1Rad := lng1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lng2Rad := lng2 * math.Pi / 180

	// Haversine 公式
	dlat := lat2Rad - lat1Rad
	dlng := lng2Rad - lng1Rad

	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dlng/2)*math.Sin(dlng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// 1KM 步进搜索算法
func stepSearch(landmarkLat, landmarkLng float64, maxDistance int) []Hospital {
	var allHospitals []Hospital

	// 按 1KM 步进搜索
	for distance := 1; distance <= maxDistance; distance++ {
		hospitals := searchHospitalsInRadius(landmarkLat, landmarkLng, distance)
		allHospitals = append(allHospitals, hospitals...)

		// 如果已找到足够多的医院，可提前结束搜索
		if len(allHospitals) >= 30 {
			break
		}
	}

	// 按距离排序
	sort.Slice(allHospitals, func(i, j int) bool {
		return allHospitals[i].Distance < allHospitals[j].Distance
	})

	return allHospitals
}

// 在指定半径内搜索医院（简化版）
func searchHospitalsInRadius(landmarkLat, landmarkLng float64, radiusKm int) []Hospital {
	var hospitals []Hospital

	// 查询数据库中的医院
	rows, err := db.Query(`
		SELECT id, name, address, latitude, longitude, phone, hospital_type, main_departments, business_hours, qualifications, created_at, updated_at
		FROM hospitals
	`)
	if err != nil {
		return hospitals
	}
	defer rows.Close()

	for rows.Next() {
		var h Hospital
		err := rows.Scan(&h.ID, &h.Name, &h.Address, &h.Latitude, &h.Longitude, &h.Phone, &h.HospitalType, &h.MainDepartments, &h.BusinessHours, &h.Qualifications, &h.CreatedAt, &h.UpdatedAt)
		if err != nil {
			continue
		}

		// 计算距离
		distance := calculateHaversineDistance(landmarkLat, landmarkLng, h.Latitude, h.Longitude)

		// 检查是否在指定半径内
		if distance <= float64(radiusKm) {
			h.Distance = distance
			hospitals = append(hospitals, h)
		}
	}

	return hospitals
}

// 评价置信度算法
func calculateConfidence(ratings []Rating, sourceWeights map[string]float64) float64 {
	if len(ratings) == 0 {
		return 0.0
	}

	// 计算加权平均分和标准差
	var weightedRatings []float64
	var weights []float64

	for _, rating := range ratings {
		weight := sourceWeights[rating.Source]
		if weight == 0 {
			weight = 0.5 // 默认权重
		}

		weightedRatings = append(weightedRatings, rating.RatingValue*weight)
		weights = append(weights, weight)
	}

	// 计算加权平均值
	var sumWeighted float64
	var sumWeights float64
	for i, weightedRating := range weightedRatings {
		sumWeighted += weightedRating
		sumWeights += weights[i]
	}

	if sumWeights == 0 {
		return 0.0
	}

	weightedAvg := sumWeighted / sumWeights

	// 计算标准差
	var variance float64
	for _, weightedRating := range weightedRatings {
		variance += math.Pow(weightedRating-weightedAvg, 2)
	}
	variance /= float64(len(weightedRatings))
	stdDev := math.Sqrt(variance)

	// 计算置信度 - 分布越集中，置信度越高
	var confidence float64
	if stdDev > 0 {
		confidence = 1.0 / (1.0 + stdDev)
	} else {
		confidence = 1.0 // 完全一致的评价
	}

	// 根据样本数量调整置信度
	sampleFactor := math.Min(1.0, float64(len(ratings))/10.0)
	confidence = confidence * sampleFactor

	return math.Min(1.0, confidence)
}

// 综合排名算法
func rankHospitals(hospitals []Hospital, userPreferences map[string]float64) []Hospital {
	// 默认权重
	weights := map[string]float64{
		"distance":   0.4,
		"rating":     0.4,
		"confidence": 0.2,
	}

	// 应用用户自定义权重
	if userPreferences != nil {
		for key, value := range userPreferences {
			weights[key] = value
		}
	}

	// 归一化各项指标
	var maxDistance float64
	for _, hospital := range hospitals {
		if hospital.Distance > maxDistance {
			maxDistance = hospital.Distance
		}
	}

	if maxDistance == 0 {
		maxDistance = 1.0
	}

	// 计算综合得分
	for i := range hospitals {
		// 距离分数（越近越好）
		distanceScore := 1.0 - (hospitals[i].Distance / maxDistance)

		// 评分分数
		ratingScore := hospitals[i].Rating / 5.0 // 假设评分满分为5

		// 置信度分数
		confidenceScore := hospitals[i].Confidence

		// 计算综合得分
		hospitals[i].Rating = weights["distance"]*distanceScore +
			weights["rating"]*ratingScore +
			weights["confidence"]*confidenceScore
	}

	// 按综合得分排序
	sort.Slice(hospitals, func(i, j int) bool {
		return hospitals[i].Rating > hospitals[j].Rating
	})

	// 返回前10名
	if len(hospitals) > 10 {
		return hospitals[:10]
	}

	return hospitals
}

// 正态分布分析法评估置信度
func calculateNormalDistributionConfidence(ratings []float64) float64 {
	if len(ratings) == 0 {
		return 0.0
	}

	// 计算平均值
	var sum float64
	for _, rating := range ratings {
		sum += rating
	}
	mean := sum / float64(len(ratings))

	// 计算标准差
	var variance float64
	for _, rating := range ratings {
		variance += math.Pow(rating-mean, 2)
	}
	variance /= float64(len(ratings))
	stdDev := math.Sqrt(variance)

	// 计算置信度 - 标准差越小，置信度越高
	if stdDev > 0 {
		confidence := 1.0 / (1.0 + stdDev)
		return math.Min(1.0, confidence)
	}

	return 1.0 // 完全一致
}

// 多来源置信度加权算法
func calculateMultiSourceConfidence(ratings []Rating) float64 {
	// 定义不同来源的权重
	sourceWeights := map[string]float64{
		"官方评级":        0.9,
		"用户评分":        0.7,
		"Google Maps": 0.8,
		"其他":          0.5,
	}

	return calculateConfidence(ratings, sourceWeights)
}

// 地理便利性评分算法
func calculateGeographicConvenience(lat, lng float64, hospital Hospital) float64 {
	// 计算距离
	distance := calculateHaversineDistance(lat, lng, hospital.Latitude, hospital.Longitude)

	// 距离越近，便利性越高
	// 使用指数衰减函数
	convenience := math.Exp(-distance / 5.0) // 5km 为基准距离

	return math.Min(1.0, convenience)
}

// 综合评分算法
func calculateComprehensiveScore(hospital Hospital, userLat, userLng float64, userPreferences map[string]float64) float64 {
	// 获取医院的所有评分
	ratings := getHospitalRatings(hospital.ID)

	// 计算多来源置信度
	confidence := calculateMultiSourceConfidence(ratings)

	// 计算地理便利性
	geographicScore := calculateGeographicConvenience(userLat, userLng, hospital)

	// 计算平均评分
	var avgRating float64
	if len(ratings) > 0 {
		var sum float64
		for _, rating := range ratings {
			sum += rating.RatingValue
		}
		avgRating = sum / float64(len(ratings))
	}

	// 默认权重
	weights := map[string]float64{
		"geographic": 0.3,
		"rating":     0.4,
		"confidence": 0.3,
	}

	// 应用用户偏好
	if userPreferences != nil {
		for key, value := range userPreferences {
			weights[key] = value
		}
	}

	// 计算综合得分
	score := weights["geographic"]*geographicScore +
		weights["rating"]*(avgRating/5.0) +
		weights["confidence"]*confidence

	return score
}

// 获取医院的所有评分
func getHospitalRatings(hospitalID int) []Rating {
	var ratings []Rating

	rows, err := db.Query(`
		SELECT id, hospital_id, source, rating_value, confidence, rating_date, created_at
		FROM ratings
		WHERE hospital_id = ?
	`, hospitalID)
	if err != nil {
		return ratings
	}
	defer rows.Close()

	for rows.Next() {
		var rating Rating
		err := rows.Scan(&rating.ID, &rating.HospitalID, &rating.Source, &rating.RatingValue, &rating.Confidence, &rating.RatingDate, &rating.CreatedAt)
		if err != nil {
			continue
		}
		ratings = append(ratings, rating)
	}

	return ratings
}

// 智能推荐算法
func intelligentRecommendation(userLat, userLng float64, radius int, userPreferences map[string]float64) []Hospital {
	// 1KM 步进搜索
	hospitals := stepSearch(userLat, userLng, radius)

	// 为每个医院计算综合得分
	for i := range hospitals {
		hospitals[i].Rating = calculateComprehensiveScore(hospitals[i], userLat, userLng, userPreferences)
	}

	// 综合排名
	rankedHospitals := rankHospitals(hospitals, userPreferences)

	return rankedHospitals
}

// 缓存评分计算
func calculateCachedScore(hospitalID int) (float64, float64) {
	// 从缓存或数据库获取评分
	ratings := getHospitalRatings(hospitalID)

	if len(ratings) == 0 {
		return 0.0, 0.0
	}

	// 计算平均评分
	var sumRating float64
	var sumConfidence float64
	for _, rating := range ratings {
		sumRating += rating.RatingValue
		sumConfidence += rating.Confidence
	}

	avgRating := sumRating / float64(len(ratings))
	avgConfidence := sumConfidence / float64(len(ratings))

	return avgRating, avgConfidence
}
