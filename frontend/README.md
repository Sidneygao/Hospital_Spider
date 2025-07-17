# 医院智能推荐系统 - 前端

基于 React + TypeScript + Ant Design 的现代化医院搜索和推荐系统前端。

## 🎨 设计特色

- **淡雅色调**: 采用 `#8B9DC3` 主色调，营造专业、舒适的视觉体验
- **响应式设计**: 完美适配桌面端、平板和移动端
- **卡片式布局**: 医院信息以圆角阴影卡片形式展示
- **实时搜索**: 支持真实地址输入和地理位置搜索
- **交互友好**: 丰富的动画效果和用户反馈

## 🚀 快速开始

### 环境要求

- Node.js >= 16.0.0
- npm >= 8.0.0

### 安装依赖

```bash
cd frontend
npm install
```

### 环境配置

创建 `.env` 文件并配置以下环境变量：

```env
# API 配置
REACT_APP_API_BASE_URL=http://localhost:3000/api

# Google Maps API 配置
REACT_APP_GOOGLE_MAPS_API_KEY=your_google_maps_api_key_here  # 必须通过环境变量配置，严禁在代码和配置文件中明文出现API Key

# 应用配置
REACT_APP_NAME=医院智能推荐系统
REACT_APP_VERSION=1.0.0
REACT_APP_ENVIRONMENT=development

# 功能开关
REACT_APP_ENABLE_ANALYTICS=false
REACT_APP_ENABLE_DEBUG=true
REACT_APP_ENABLE_MOCK_DATA=true
```

### 启动开发服务器

```bash
npm start
```

应用将在 http://localhost:3001 启动

### 构建生产版本

```bash
npm run build
```

## 📁 项目结构

```
src/
├── components/          # 组件目录
│   ├── SearchBar/      # 搜索栏组件
│   ├── HospitalMap/    # 医院地图组件
│   ├── HospitalList/   # 医院列表组件
│   └── ...
├── styles/             # 样式文件
│   ├── App.css         # 主应用样式
│   ├── index.css       # 全局样式
│   └── ...
├── services/           # API 服务
│   └── api.ts          # API 接口定义
├── types/              # TypeScript 类型定义
│   └── hospital.ts     # 医院相关类型
├── utils/              # 工具函数
├── hooks/              # 自定义 Hooks
└── constants/          # 常量定义
```

## 🎯 核心功能

### 1. 智能搜索
- 支持真实地址输入
- 地理位置自动定位
- 搜索建议和自动完成
- 历史搜索记录

### 2. 地图展示
- Google Maps 集成
- 医院位置标记
- 搜索位置高亮
- 地图交互操作

### 3. 医院列表
- 卡片式布局展示
- 评分和标签显示
- 距离和联系方式
- 收藏和分享功能

### 4. 响应式设计
- 移动端优化
- 触摸友好交互
- 自适应布局
- 性能优化

## 🎨 设计系统

### 色彩方案

```css
/* 主色调 */
--primary-color: #8B9DC3;      /* 淡雅蓝灰 */
--secondary-color: #6B7A99;    /* 深蓝灰 */
--accent-color: #F5F7FA;       /* 浅灰白 */

/* 功能色 */
--success-color: #52C41A;      /* 成功绿 */
--warning-color: #FAAD14;      /* 警告橙 */
--error-color: #FF4D4F;        /* 错误红 */
--info-color: #1890FF;         /* 信息蓝 */

/* 中性色 */
--text-primary: #262626;       /* 主文字 */
--text-secondary: #595959;     /* 次文字 */
--text-muted: #8C8C8C;         /* 辅助文字 */
--border-color: #D9D9D9;       /* 边框色 */
--bg-color: #FAFAFA;           /* 背景色 */
```

### 组件设计

- **圆角设计**: 统一使用 12px-16px 圆角
- **阴影效果**: 多层次阴影营造立体感
- **动画过渡**: 0.3s ease-in-out 统一过渡
- **间距规范**: 8px 基础间距单位

## 🔧 开发指南

### 代码规范

- 使用 TypeScript 进行类型检查
- 遵循 ESLint 规则
- 使用函数式组件和 Hooks
- 组件采用 PascalCase 命名

### 样式规范

- 使用 CSS Modules 或内联样式
- 遵循 BEM 命名规范
- 响应式设计优先
- 性能优化考虑

### 组件开发

```typescript
// 组件模板
import React from 'react';
import { ComponentProps } from './types';
import './Component.css';

interface ComponentProps {
  // 属性定义
}

const Component: React.FC<ComponentProps> = ({ ...props }) => {
  // 组件逻辑
  
  return (
    <div className="component">
      {/* 组件内容 */}
    </div>
  );
};

export default Component;
```

## 📱 移动端适配

### 断点设置

```css
/* 移动端 */
@media (max-width: 768px) {
  /* 移动端样式 */
}

/* 平板端 */
@media (min-width: 769px) and (max-width: 1024px) {
  /* 平板端样式 */
}

/* 桌面端 */
@media (min-width: 1025px) {
  /* 桌面端样式 */
}
```

### 触摸优化

- 按钮最小触摸区域 44px
- 卡片间距适配手指操作
- 滑动和手势支持
- 键盘友好设计

## 🚀 性能优化

### 代码分割

```typescript
// 懒加载组件
const LazyComponent = React.lazy(() => import('./LazyComponent'));

// 路由懒加载
const HomePage = React.lazy(() => import('./pages/HomePage'));
```

### 缓存策略

- API 请求缓存
- 图片懒加载
- 组件记忆化
- 本地存储优化

## 🧪 测试

```bash
# 运行测试
npm test

# 类型检查
npm run type-check

# 代码检查
npm run lint
```

## 📦 构建部署

### 生产构建

```bash
npm run build
```

### 部署配置

- 支持静态文件部署
- CDN 加速配置
- 环境变量管理
- 错误监控集成

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交代码变更
4. 发起 Pull Request

## 📄 许可证

MIT License

## 📞 联系方式

- 项目维护者: Hospital Spider Team
- 邮箱: support@hospital-spider.com
- 项目地址: https://github.com/Sidneygao/Hospital_Spider 