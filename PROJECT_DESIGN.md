# Hospital_Spider 项目设计文档

## 系统架构设计

### 整体架构
```
+------------------+     +------------------+     +------------------+
|                  |     |                  |     |                  |
|  前端展示层      |<--->|  后端API服务     |<--->|  数据采集层      |
|                  |     |                  |     |                  |
+------------------+     +------------------+     +------------------+
         ^                       ^                       ^
         |                       |                       |
         v                       v                       v
+------------------+     +------------------+     +------------------+
|                  |     |                  |     |                  |
|  用户交互模块    |     |  数据处理模块    |     |  外部API对接    |
|                  |     |                  |     |                  |
+------------------+     +------------------+     +------------------+
```

### 模块划分

1. **数据采集层**
   - Google Maps爬虫模块
   - 医院评级数据爬虫模块
   - 用户评论爬虫模块
   - 数据更新与缓存管理模块

2. **后端API服务**
   - 医院查询API
   - 地理位置服务API
   - 评级与评论聚合API
   - 用户反馈API

3. **数据处理模块**
   - 地理距离计算引擎
   - 评级置信度模型
   - 综合排名算法
   - 数据清洗与标准化

4. **前端展示层**
   - 搜索界面
   - 医院列表展示
   - 医院详情页面
   - 地图交互组件

5. **用户交互模块**
   - 用户评论系统
   - 评分反馈系统
   - 位置选择器

## 数据模型设计

### 实体关系图
```
+--------------+       +--------------+       +--------------+
|    医院      |<------|    地理位置   |------>|    区域      |
+--------------+       +--------------+       +--------------+
      |                       |                    |
      |                       |                    |
      v                       v                    v
+--------------+       +--------------+       +--------------+
|    评级      |       |    距离数据   |       |   地标信息   |
+--------------+       +--------------+       +--------------+
      |
      |
      v
+--------------+       +--------------+
|    评论      |------>|    用户      |
+--------------+       +--------------+
```

### 数据表设计

1. **hospitals表** - 存储医院基本信息
```sql
CREATE TABLE hospitals (
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
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

2. **ratings表** - 存储医院评级信息
```sql
CREATE TABLE ratings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hospital_id INTEGER NOT NULL,
    source TEXT NOT NULL,
    rating_value REAL NOT NULL,
    confidence REAL DEFAULT 0.5,
    rating_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (hospital_id) REFERENCES hospitals(id)
);
```

3. **reviews表** - 存储用户评论
```sql
CREATE TABLE reviews (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hospital_id INTEGER NOT NULL,
    source TEXT NOT NULL,
    user_name TEXT,
    rating REAL,
    review_text TEXT,
    review_date TIMESTAMP,
    sentiment_score REAL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (hospital_id) REFERENCES hospitals(id)
);
```

4. **landmarks表** - 存储地标信息
```sql
CREATE TABLE landmarks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    address TEXT NOT NULL,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL,
    search_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

5. **hospital_distances表** - 存储医院与地标的距离关系
```sql
CREATE TABLE hospital_distances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hospital_id INTEGER NOT NULL,
    landmark_id INTEGER NOT NULL,
    distance_km REAL NOT NULL,
    travel_time_minutes INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (hospital_id) REFERENCES hospitals(id),
    FOREIGN KEY (landmark_id) REFERENCES landmarks(id)
);
```

6. **user_feedback表** - 存储用户提供的反馈
```sql
CREATE TABLE user_feedback (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hospital_id INTEGER NOT NULL,
    user_ip TEXT,
    rating REAL,
    comment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (hospital_id) REFERENCES hospitals(id)
);
```

7. **cache表** - 存储API请求缓存
```sql
CREATE TABLE cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    query_hash TEXT NOT NULL UNIQUE,
    data TEXT NOT NULL,
    expiry TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## API接口设计

### 1. 搜索API

#### 根据地标搜索医院
```
GET /api/hospitals/search
参数:
- landmark: 地标名称或地址
- lat: 纬度(可选)
- lng: 经度(可选)
- radius: 搜索半径，默认为10km
- limit: 返回结果数量，默认为10
- type: 医院类型筛选(可选)

返回:
{
  "status": "success",
  "count": 10,
  "data": [
    {
      "id": 1,
      "name": "医院名称",
      "address": "详细地址",
      "distance": 1.5,
      "rating": 4.5,
      "confidence": 0.85,
      "hospital_type": "综合医院",
      "main_departments": ["内科", "外科", "妇产科"]
    },
    ...
  ]
}
```

### 2. 医院详情API

#### 获取医院详细信息
```
GET /api/hospitals/{id}

返回:
{
  "status": "success",
  "data": {
    "id": 1,
    "name": "医院名称",
    "address": "详细地址",
    "location": {
      "lat": 13.7563,
      "lng": 100.5018
    },
    "phone": "+66 2 123 4567",
    "hospital_type": "综合医院",
    "main_departments": ["内科", "外科", "妇产科"],
    "business_hours": "周一至周五 8:00-17:00",
    "qualifications": ["JCI认证", "国家三甲"],
    "ratings": [
      {
        "source": "官方评级",
        "value": 5.0,
        "confidence": 0.95
      },
      {
        "source": "用户评分",
        "value": 4.3,
        "confidence": 0.75
      }
    ],
    "reviews": [
      {
        "user": "匿名用户",
        "rating": 4,
        "text": "服务很好，医生专业",
        "date": "2023-05-15"
      },
      ...
    ]
  }
}
```

### 3. 用户反馈API

#### 提交用户评价
```
POST /api/hospitals/{id}/feedback
参数:
{
  "rating": 4.5,
  "comment": "用户评论内容"
}

返回:
{
  "status": "success",
  "message": "反馈已提交",
  "data": {
    "id": 123,
    "hospital_id": 1,
    "rating": 4.5,
    "comment": "用户评论内容",
    "created_at": "2023-07-20T14:30:15Z"
  }
}
```

## 算法设计

### 1. 1KM步进搜索算法

基于给定地标，按照1KM递增半径搜索医院：

```python
def step_search(landmark_lat, landmark_lng, max_distance=10):
    """
    1KM步进搜索算法
    
    参数:
        landmark_lat: 地标纬度
        landmark_lng: 地标经度
        max_distance: 最大搜索距离(km)
        
    返回:
        hospitals: 按距离排序的医院列表
    """
    all_hospitals = []
    
    for distance in range(1, max_distance + 1):
        # 搜索当前步进圈内的医院
        hospitals = search_hospitals_in_radius(
            landmark_lat, landmark_lng, 
            min_radius=(distance - 1), 
            max_radius=distance
        )
        
        # 添加到结果列表
        all_hospitals.extend(hospitals)
        
        # 如果已找到足够多的医院，可提前结束搜索
        if len(all_hospitals) >= 30:
            break
            
    # 返回按距离排序的医院
    return sorted(all_hospitals, key=lambda x: x['distance'])
```

### 2. 评价置信度算法

基于正态分布计算评论的置信度：

```python
def calculate_confidence(ratings, source_weights):
    """
    评价置信度算法
    
    参数:
        ratings: 评分列表
        source_weights: 来源权重字典
        
    返回:
        confidence: 置信度评分(0-1)
    """
    if not ratings:
        return 0.0
    
    # 计算加权平均分和标准差
    weighted_ratings = []
    weights = []
    
    for rating in ratings:
        source = rating['source']
        value = rating['value']
        weight = source_weights.get(source, 0.5)
        
        weighted_ratings.append(value * weight)
        weights.append(weight)
    
    # 计算加权平均值
    if sum(weights) > 0:
        weighted_avg = sum(weighted_ratings) / sum(weights)
    else:
        return 0.0
    
    # 计算标准差
    variance = sum([(r - weighted_avg) ** 2 for r in weighted_ratings]) / len(weighted_ratings)
    std_dev = math.sqrt(variance)
    
    # 计算置信度 - 分布越集中，置信度越高
    if std_dev > 0:
        confidence = 1.0 / (1.0 + std_dev)
    else:
        confidence = 1.0  # 完全一致的评价
        
    # 根据样本数量调整置信度
    sample_factor = min(1.0, len(ratings) / 10.0)
    confidence = confidence * sample_factor
    
    return min(1.0, confidence)
```

### 3. 综合排名算法

融合距离、评级和评论数据的综合排名算法：

```python
def rank_hospitals(hospitals, user_preferences=None):
    """
    医院综合排名算法
    
    参数:
        hospitals: 医院列表
        user_preferences: 用户偏好设置(可选)
        
    返回:
        ranked_hospitals: 排序后的医院列表
    """
    # 默认权重
    weights = {
        'distance': 0.4,
        'rating': 0.4,
        'confidence': 0.2
    }
    
    # 应用用户自定义权重(如果有)
    if user_preferences:
        weights.update(user_preferences)
    
    # 归一化各项指标
    max_distance = max([h['distance'] for h in hospitals]) if hospitals else 1.0
    
    for hospital in hospitals:
        # 距离分数(越近越好)
        distance_score = 1.0 - (hospital['distance'] / max_distance)
        
        # 评分分数
        rating_score = hospital['rating'] / 5.0  # 假设评分满分为5
        
        # 置信度分数
        confidence_score = hospital['confidence']
        
        # 计算综合得分
        hospital['score'] = (
            weights['distance'] * distance_score +
            weights['rating'] * rating_score +
            weights['confidence'] * confidence_score
        )
    
    # 按综合得分排序
    ranked_hospitals = sorted(hospitals, key=lambda x: x['score'], reverse=True)
    
    return ranked_hospitals[:10]  # 返回前10名
```

## 部署架构

### 开发环境
- Python 3.9+
- Flask Web框架
- SQLite数据库
- 前端: HTML5 + CSS3 + JavaScript

### 生产环境
- Render托管服务
- PostgreSQL数据库
- Nginx反向代理
- 定时任务: 数据更新与缓存清理

### 扩展性考虑
- 数据库分片: 按地理区域划分
- API缓存层: Redis缓存常用查询
- 负载均衡: 多实例部署 