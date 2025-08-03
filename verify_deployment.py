#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Hospital Spider 部署验证脚本
用于检查 Render 部署是否成功
"""

import requests
import time
import json
import sys

def check_backend_health(backend_url):
    """检查后端服务健康状态"""
    try:
        response = requests.get(f"{backend_url}/api/hospitals", timeout=10)
        if response.status_code == 200:
            print(f"✅ 后端服务正常: {backend_url}")
            return True
        else:
            print(f"❌ 后端服务异常 (状态码: {response.status_code}): {backend_url}")
            return False
    except Exception as e:
        print(f"❌ 后端服务连接失败: {e}")
        return False

def check_frontend_health(frontend_url):
    """检查前端服务健康状态"""
    try:
        response = requests.get(frontend_url, timeout=10)
        if response.status_code == 200:
            print(f"✅ 前端服务正常: {frontend_url}")
            return True
        else:
            print(f"❌ 前端服务异常 (状态码: {response.status_code}): {frontend_url}")
            return False
    except Exception as e:
        print(f"❌ 前端服务连接失败: {e}")
        return False

def test_api_endpoints(backend_url):
    """测试API端点"""
    endpoints = [
        "/api/hospitals",
        "/api/hospitals/search?lat=39.9042&lng=116.4074&radius=5"
    ]
    
    print("\n🔍 测试API端点:")
    for endpoint in endpoints:
        try:
            response = requests.get(f"{backend_url}{endpoint}", timeout=10)
            if response.status_code == 200:
                print(f"✅ {endpoint} - 正常")
            else:
                print(f"❌ {endpoint} - 异常 (状态码: {response.status_code})")
        except Exception as e:
            print(f"❌ {endpoint} - 连接失败: {e}")

def main():
    print("🏥 Hospital Spider 部署验证工具")
    print("=" * 50)
    
    # 获取服务URL（用户需要输入）
    print("\n📝 请输入您的服务URL:")
    backend_url = input("后端服务URL (例如: https://hospital-spider-backend.onrender.com): ").strip()
    frontend_url = input("前端服务URL (例如: https://hospital-spider-frontend.onrender.com): ").strip()
    
    if not backend_url or not frontend_url:
        print("❌ 请提供有效的服务URL")
        return
    
    print(f"\n🔍 开始验证部署...")
    print(f"后端: {backend_url}")
    print(f"前端: {frontend_url}")
    
    # 等待服务启动
    print("\n⏳ 等待服务启动 (30秒)...")
    time.sleep(30)
    
    # 检查服务健康状态
    backend_ok = check_backend_health(backend_url)
    frontend_ok = check_frontend_health(frontend_url)
    
    # 测试API端点
    if backend_ok:
        test_api_endpoints(backend_url)
    
    # 总结
    print("\n" + "=" * 50)
    print("📊 部署验证结果:")
    print(f"后端服务: {'✅ 正常' if backend_ok else '❌ 异常'}")
    print(f"前端服务: {'✅ 正常' if frontend_ok else '❌ 异常'}")
    
    if backend_ok and frontend_ok:
        print("\n🎉 恭喜！部署成功！")
        print(f"🌐 访问您的应用: {frontend_url}")
    else:
        print("\n⚠️ 部署存在问题，请检查:")
        print("1. 环境变量是否正确设置")
        print("2. 服务是否正在构建中")
        print("3. 网络连接是否正常")

if __name__ == "__main__":
    main() 