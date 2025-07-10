# Hospital Spider - 医院智能推荐系统

基于地理位置的医院智能推荐服务，为用户提供多维度聚合的医院推荐。

## 项目概述

本项目旨在为用户提供基于地理位置的医院智能推荐服务。用户可通过定位或输入关键地标，系统自动按地理便利性、官方评级、用户评论等多维度聚合，推荐附近优质医院，并展示详细信息，提升就医决策效率。

## 技术栈

### 后端
- **语言**: Go 1.21+
- **Web框架**: Gin
- **数据库**: SQLite
- **爬虫**: 自定义爬虫模块
- **算法**: 地理距离计算、置信度算法、综合排名算法

### 前端
- **框架**: React 18
- **UI库**: Ant Design
- **地图**: Google Maps API
- **状态管理**: Redux
- **数据可视化**: ECharts

## 核心功能

1. **地理位置医院检索**
   - 支持用户输入地标或自动定位
   - 按1KM步进（至10KM）检索附近医院
   - 拉取医院基础信息（名称、地址、经纬度、类型、电话、重点科室、营业时间、资质等）

2. **多源评级与评论聚合**
   - 集成官方评级机构数据，支持多来源置信度加权
   - 聚合消费评分网站评论，采用正态分布分析法评估置信度
   - 综合地理便利性、官方评级、用户评论等因素，输出前10优选医院

3. **医院详情与地图展示**
   - 医院卡片展示基础信息、聚合评分/评级、资质
   - 详情页展示地图、评分/评级/资质详情

4. **用户与管理员交互**
   - 支持用户和管理员在详情页评论、评级、备注，信息实时反映

5. **数据缓存与更新**
   - 初期以缓存为主，支持API请求结果缓存
   - 后续支持在线搜索与即时计算

## 快速开始

### 环境要求
- Go 1.21+
- Node.js 16+
- Google Maps API Key

### 后端启动

1. **安装依赖**
   ```bash
   go mod tidy
   ```

2. **配置环境变量**
   创建 `.env` 文件：
   ```
   GOOGLE_MAPS_API_KEY=your_google_maps_api_key
   PORT=8080
   ```

3. **启动后端服务**
   ```bash
   go run backend/*.go
   ```
   或
   ```bash
   go build -o hospital-spider backend/*.go
   ./hospital-spider
   ```

4. **访问API文档**
   - 服务启动后访问: http://localhost:8080
   - API端点:
     - `GET /api/hospitals` - 获取医院列表
     - `GET /api/hospitals/search` - 搜索医院
     - `GET /api/hospitals/:id` - 获取医院详情
     - `POST /api/hospitals/:id/feedback` - 提交用户反馈

### 前端启动

1. **安装依赖**
   ```bash
   cd frontend
   npm install
   ```

2. **启动开发服务器**
   ```bash
   npm start
   ```

3. **访问前端**
   - 浏览器访问: http://localhost:3000

## 项目结构

```
Hospital_Spider/
├── backend/
│   ├── main.go          # 主程序入口
│   ├── spider.go        # 爬虫模块
│   └── algorithm.go     # 算法模块
├── frontend/
│   ├── package.json
│   └── src/
├── go.mod               # Go模块文件
├── requirements.txt     # Python依赖（已弃用）
└── README.md
```

## API接口

### 医院搜索
```
GET /api/hospitals/search?lat=13.7563&lng=100.5018&radius=10&limit=10
```

### 医院详情
```
GET /api/hospitals/1
```

### 用户反馈
```
POST /api/hospitals/1/feedback
Content-Type: application/json

{
  "rating": 4.5,
  "comment": "服务很好，医生专业"
}
```

## 核心算法

### 1. 1KM步进搜索算法
基于给定地标，按照1KM递增半径搜索医院，确保覆盖全面。

### 2. 评价置信度算法
基于正态分布计算评论的置信度，分布越集中置信度越高。

### 3. 综合排名算法
融合距离、评级和评论数据的综合排名算法，支持用户自定义权重。

## 部署

### 开发环境
- Go 1.21+
- SQLite数据库
- 前端: React + Ant Design

### 生产环境
- 推荐使用 Docker 容器化部署
- 可部署到 Render、Heroku 等云平台
- 数据库可升级为 PostgreSQL

## 扩展功能

- 支持更多国家和地区的数据源接入
- 实现在线搜索与即时计算，提升实时性和准确性
- 丰富前端交互与可视化体验
- 添加用户认证和权限管理
- 支持多语言国际化

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License 