# 🔧 GitHub 连接问题解决指南

## 问题：在 Render 中看不到 Hospital_Spider 仓库

### 📋 快速检查清单

#### 1. 确认仓库信息
- **仓库名称**: `Hospital_Spider`
- **GitHub 账户**: `Sidneygao`
- **完整路径**: `https://github.com/Sidneygao/Hospital_Spider`

#### 2. 检查仓库可见性
1. 访问：https://github.com/Sidneygao/Hospital_Spider
2. 查看页面右上角是否显示 "Public" 标签
3. 如果没有显示 "Public"，需要将仓库设为公开

### 🚀 详细解决步骤

#### 步骤 1：将仓库设为公开（如果还没有）

1. **访问仓库设置页面**
   - 打开：https://github.com/Sidneygao/Hospital_Spider
   - 点击顶部的 "Settings" 标签

2. **更改可见性**
   - 向下滚动到 "Danger Zone"
   - 点击 "Change repository visibility"
   - 选择 "Make public"
   - 输入仓库名称 `Hospital_Spider` 确认

3. **确认更改**
   - 页面应该显示 "Public" 标签
   - 等待几分钟让更改生效

#### 步骤 2：重新连接 GitHub 账户

1. **断开当前连接**
   - 访问：https://dashboard.render.com
   - 点击右上角头像
   - 选择 "Account Settings"
   - 找到 "GitHub" 部分
   - 点击 "Disconnect"

2. **重新连接 GitHub**
   - 点击 "Connect GitHub"
   - 使用您的 GitHub 账户登录
   - 授权 Render 访问您的仓库

#### 步骤 3：检查 GitHub 应用权限

1. **访问 GitHub 应用设置**
   - 打开：https://github.com/settings/connections/applications
   - 找到 "Render" 应用

2. **检查权限设置**
   - 确保 Render 有访问仓库的权限
   - 如果没有，点击 "Configure" 重新授权

#### 步骤 4：在 Render 中搜索仓库

1. **创建 Blueprint**
   - 在 Render 控制台点击 "New" → "Blueprint"
   - 在仓库搜索框中输入：`Hospital_Spider`

2. **如果仍然看不到仓库**
   - 尝试搜索：`Sidneygao/Hospital_Spider`
   - 或者搜索：`hospital-spider`

### 🔍 替代解决方案

#### 方案 A：手动创建服务（如果 Blueprint 不工作）

1. **创建后端服务**
   - 点击 "New" → "Web Service"
   - 连接 GitHub 仓库
   - 配置：
     - **Name**: `hospital-spider-backend`
     - **Environment**: `Go`
     - **Build Command**: `cd backend && go build -o main .`
     - **Start Command**: `cd backend && ./main`

2. **创建前端服务**
   - 点击 "New" → "Static Site"
   - 连接同一个 GitHub 仓库
   - 配置：
     - **Name**: `hospital-spider-frontend`
     - **Build Command**: `cd frontend && npm install && npm run build`
     - **Publish Directory**: `frontend/build`

#### 方案 B：使用 GitHub 链接

1. **直接使用仓库 URL**
   - 在 Render 中手动输入仓库 URL：
   - `https://github.com/Sidneygao/Hospital_Spider`

### 🐛 常见错误及解决方案

#### 错误 1：仓库未找到
**原因**: 仓库是私有的或权限不足
**解决**: 将仓库设为公开

#### 错误 2：权限被拒绝
**原因**: GitHub 应用权限不足
**解决**: 重新授权 Render 应用

#### 错误 3：仓库名称不匹配
**原因**: 大小写或拼写错误
**解决**: 确认仓库名称完全正确

### 📞 获取帮助

如果以上步骤都无法解决问题：

1. **检查 GitHub 账户**
   - 确认您登录的是正确的 GitHub 账户
   - 确认您是该仓库的所有者

2. **联系 Render 支持**
   - 访问：https://render.com/docs/help
   - 提供详细的错误信息

3. **检查网络连接**
   - 确保可以正常访问 GitHub
   - 尝试清除浏览器缓存

### ✅ 成功标志

当您看到以下内容时，说明问题已解决：

- ✅ 在 Render 中可以看到 `Hospital_Spider` 仓库
- ✅ 可以成功创建 Blueprint 或手动服务
- ✅ 服务开始构建和部署

### 🎯 下一步

一旦仓库可见，继续按照 `部署检查清单.md` 中的步骤完成部署。 