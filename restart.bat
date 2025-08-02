@echo off
chcp 65001 >nul
echo ========================================
echo 医院爬虫项目 - 服务重启脚本
echo ========================================
echo.

echo 清理端口...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8080') do taskkill /PID %%a /F >nul 2>&1
for /f "tokens=5" %%a in ('netstat -ano ^| findstr :3000') do taskkill /PID %%a /F >nul 2>&1
for /f "tokens=5" %%a in ('netstat -ano ^| findstr :3001') do taskkill /PID %%a /F >nul 2>&1
for /f "tokens=5" %%a in ('netstat -ano ^| findstr :3002') do taskkill /PID %%a /F >nul 2>&1

echo 等待端口释放...
timeout /t 3 /nobreak >nul

echo.
echo 启动后端服务...
if exist "backend\main.go" (
    cd backend
    start "后端服务" cmd /k "go run main.go"
    cd ..
) else (
    echo 错误: 未找到 backend\main.go
    pause
    exit /b 1
)

echo 等待后端启动...
timeout /t 5 /nobreak >nul

echo.
echo 启动前端服务...
if exist "frontend\package.json" (
    cd frontend
    start "前端服务" cmd /k "npm start"
    cd ..
) else (
    echo 错误: 未找到 frontend\package.json
    pause
    exit /b 1
)

echo.
echo 等待前端启动...
timeout /t 10 /nobreak >nul

echo.
echo ========================================
echo 服务启动完成！
echo ========================================
echo 后端服务: http://localhost:8080
echo 前端服务: http://localhost:3000
echo.
echo 如果服务未正常启动，请检查:
echo 1. 环境变量AMAP_KEY是否设置
echo 2. Go环境是否正确安装
echo 3. Node.js环境是否正确安装
echo.
pause 