#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
测试本地数据使用情况
验证后端是否正确使用本地JSON数据而不是API调用
"""

import requests
import json
import time

def test_local_search():
    """测试本地搜索功能"""
    print("🔍 测试本地搜索功能...")
    
    # 测试参数
    test_params = [
        {"lat": 39.9042, "lng": 116.4074, "radius": 5, "limit": 5},  # 北京
        {"lat": 31.2304, "lng": 121.4737, "radius": 3, "limit": 3},  # 上海
        {"lat": 23.1291, "lng": 113.2644, "radius": 2, "limit": 2},  # 广州
    ]
    
    base_url = "http://localhost:8080"
    
    for i, params in enumerate(test_params, 1):
        print(f"\n📋 测试 {i}: {params}")
        
        try:
            response = requests.get(f"{base_url}/api/hospitals/search", params=params, timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                print(f"✅ 搜索成功")
                print(f"   找到医院数量: {data.get('count', 0)}")
                
                if data.get('data'):
                    for j, hospital in enumerate(data['data'][:3], 1):  # 只显示前3个
                        print(f"   {j}. {hospital.get('name', 'N/A')} - 距离: {hospital.get('distance', 0):.2f}km")
            else:
                print(f"❌ 搜索失败: {response.status_code}")
                print(f"   响应: {response.text}")
                
        except Exception as e:
            print(f"❌ 请求失败: {e}")

def test_local_geocode():
    """测试本地地理编码功能"""
    print("\n🔍 测试本地地理编码功能...")
    
    # 测试地址
    test_addresses = [
        "北京市朝阳区建国门外大街1号",
        "上海市浦东新区陆家嘴环路1000号",
        "广州市天河区珠江新城花城大道85号",
        "深圳市南山区深南大道10000号",
        "未知地址测试",  # 这个应该会调用API
    ]
    
    base_url = "http://localhost:8080"
    
    for i, address in enumerate(test_addresses, 1):
        print(f"\n📋 测试 {i}: {address}")
        
        try:
            response = requests.get(f"{base_url}/api/amap/geo", params={"address": address}, timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                print(f"✅ 地理编码成功")
                print(f"   状态: {data.get('status')}")
                print(f"   信息: {data.get('info')}")
                
                if data.get('geocodes'):
                    geocode = data['geocodes'][0]
                    print(f"   地址: {geocode.get('formatted_address')}")
                    print(f"   坐标: {geocode.get('location')}")
            else:
                print(f"❌ 地理编码失败: {response.status_code}")
                print(f"   响应: {response.text}")
                
        except Exception as e:
            print(f"❌ 请求失败: {e}")

def test_api_usage():
    """测试API使用情况"""
    print("\n🔍 测试API使用情况...")
    
    base_url = "http://localhost:8080"
    
    # 测试不需要API调用的功能
    print("📋 测试本地数据功能...")
    
    try:
        # 测试获取所有医院（应该使用本地数据）
        response = requests.get(f"{base_url}/api/hospitals", timeout=10)
        
        if response.status_code == 200:
            data = response.json()
            print(f"✅ 获取医院列表成功")
            print(f"   医院数量: {len(data) if isinstance(data, list) else 'N/A'}")
        else:
            print(f"❌ 获取医院列表失败: {response.status_code}")
            
    except Exception as e:
        print(f"❌ 请求失败: {e}")

def main():
    """主函数"""
    print("🏥 Hospital Spider 本地数据测试")
    print("=" * 50)
    
    # 检查后端是否运行
    try:
        response = requests.get("http://localhost:8080/api/hospitals", timeout=5)
        if response.status_code != 200:
            print("❌ 后端服务未运行或无法访问")
            print("请先启动后端服务: cd backend && go run main.go")
            return
    except Exception as e:
        print("❌ 无法连接到后端服务")
        print("请先启动后端服务: cd backend && go run main.go")
        return
    
    print("✅ 后端服务运行正常")
    
    # 执行测试
    test_local_search()
    test_local_geocode()
    test_api_usage()
    
    print("\n" + "=" * 50)
    print("📊 测试总结:")
    print("✅ 如果看到本地缓存相关的日志，说明后端正确使用了本地数据")
    print("✅ 如果地理编码测试中某些地址显示'使用本地缓存'，说明本地地理编码工作正常")
    print("✅ 如果医院搜索返回数据，说明本地JSON数据被正确使用")
    
    print("\n💡 优化建议:")
    print("1. 检查后端日志，确认是否显示'使用本地JSON数据'和'使用本地缓存'")
    print("2. 如果仍有API调用，可以进一步扩展本地缓存")
    print("3. 考虑将更多医院数据添加到本地JSON文件中")

if __name__ == "__main__":
    main() 