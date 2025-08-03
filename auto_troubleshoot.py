#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Hospital Spider 自动排查脚本
用于诊断 GitHub 连接和 Render 部署问题
"""

import requests
import json
import subprocess
import sys
import time
from datetime import datetime

def check_git_status():
    """检查 Git 状态"""
    print("🔍 检查 Git 状态...")
    try:
        result = subprocess.run(['git', 'status'], capture_output=True, text=True)
        if result.returncode == 0:
            print("✅ Git 状态正常")
            return True
        else:
            print(f"❌ Git 状态异常: {result.stderr}")
            return False
    except Exception as e:
        print(f"❌ Git 检查失败: {e}")
        return False

def check_remote_repository():
    """检查远程仓库配置"""
    print("🔍 检查远程仓库配置...")
    try:
        result = subprocess.run(['git', 'remote', '-v'], capture_output=True, text=True)
        if result.returncode == 0:
            print("✅ 远程仓库配置:")
            print(result.stdout)
            return True
        else:
            print(f"❌ 远程仓库配置异常: {result.stderr}")
            return False
    except Exception as e:
        print(f"❌ 远程仓库检查失败: {e}")
        return False

def check_github_repository_access():
    """检查 GitHub 仓库访问权限"""
    print("🔍 检查 GitHub 仓库访问权限...")
    repo_url = "https://github.com/Sidneygao/Hospital_Spider"
    try:
        response = requests.get(repo_url, timeout=10)
        if response.status_code == 200:
            print(f"✅ GitHub 仓库可访问: {repo_url}")
            return True
        else:
            print(f"❌ GitHub 仓库访问异常 (状态码: {response.status_code})")
            return False
    except Exception as e:
        print(f"❌ GitHub 仓库访问失败: {e}")
        return False

def check_repository_visibility():
    """检查仓库可见性"""
    print("🔍 检查仓库可见性...")
    repo_url = "https://github.com/Sidneygao/Hospital_Spider"
    try:
        response = requests.get(repo_url, timeout=10)
        if response.status_code == 200:
            # 检查页面内容是否包含私有仓库的标识
            if "This is a private repository" in response.text:
                print("❌ 仓库是私有的，需要设为公开")
                return False
            elif "Public" in response.text:
                print("✅ 仓库是公开的")
                return True
            else:
                print("⚠️ 无法确定仓库可见性，请手动检查")
                return None
        else:
            print(f"❌ 无法访问仓库 (状态码: {response.status_code})")
            return False
    except Exception as e:
        print(f"❌ 仓库可见性检查失败: {e}")
        return False

def check_render_api():
    """检查 Render API 状态"""
    print("🔍 检查 Render API 状态...")
    try:
        response = requests.get("https://api.render.com/v1/services", timeout=10)
        if response.status_code == 200:
            print("✅ Render API 正常")
            return True
        else:
            print(f"❌ Render API 异常 (状态码: {response.status_code})")
            return False
    except Exception as e:
        print(f"❌ Render API 检查失败: {e}")
        return False

def check_github_api():
    """检查 GitHub API 状态"""
    print("🔍 检查 GitHub API 状态...")
    try:
        response = requests.get("https://api.github.com", timeout=10)
        if response.status_code == 200:
            print("✅ GitHub API 正常")
            return True
        else:
            print(f"❌ GitHub API 异常 (状态码: {response.status_code})")
            return False
    except Exception as e:
        print(f"❌ GitHub API 检查失败: {e}")
        return False

def check_network_connectivity():
    """检查网络连接"""
    print("🔍 检查网络连接...")
    test_urls = [
        "https://github.com",
        "https://render.com",
        "https://api.github.com"
    ]
    
    all_ok = True
    for url in test_urls:
        try:
            response = requests.get(url, timeout=5)
            if response.status_code == 200:
                print(f"✅ {url} - 正常")
            else:
                print(f"❌ {url} - 异常 (状态码: {response.status_code})")
                all_ok = False
        except Exception as e:
            print(f"❌ {url} - 连接失败: {e}")
            all_ok = False
    
    return all_ok

def generate_diagnostic_report():
    """生成诊断报告"""
    print("\n" + "=" * 60)
    print("🏥 Hospital Spider 自动诊断报告")
    print("=" * 60)
    print(f"诊断时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    
    # 执行各项检查
    checks = {
        "Git 状态": check_git_status(),
        "远程仓库配置": check_remote_repository(),
        "GitHub 仓库访问": check_github_repository_access(),
        "仓库可见性": check_repository_visibility(),
        "网络连接": check_network_connectivity(),
        "GitHub API": check_github_api(),
        "Render API": check_render_api()
    }
    
    # 生成报告
    print("\n📊 诊断结果:")
    print("-" * 40)
    
    passed = 0
    failed = 0
    warnings = 0
    
    for check_name, result in checks.items():
        if result is True:
            print(f"✅ {check_name}: 正常")
            passed += 1
        elif result is False:
            print(f"❌ {check_name}: 异常")
            failed += 1
        else:
            print(f"⚠️ {check_name}: 需要手动检查")
            warnings += 1
    
    print("\n" + "=" * 60)
    print("📈 统计结果:")
    print(f"✅ 正常: {passed}")
    print(f"❌ 异常: {failed}")
    print(f"⚠️ 需要检查: {warnings}")
    
    # 提供建议
    print("\n💡 建议:")
    if failed > 0:
        print("🔧 发现异常，请按照以下步骤解决:")
        if checks["仓库可见性"] is False:
            print("1. 将 GitHub 仓库设为公开")
            print("   - 访问: https://github.com/Sidneygao/Hospital_Spider")
            print("   - 点击 Settings → General → Danger Zone")
            print("   - 点击 'Change repository visibility' → 'Make public'")
        
        if checks["网络连接"] is False:
            print("2. 检查网络连接")
            print("   - 确保可以访问 GitHub 和 Render")
            print("   - 尝试清除浏览器缓存")
        
        print("3. 重新连接 GitHub 账户到 Render")
        print("   - 访问: https://dashboard.render.com")
        print("   - 在 Account Settings 中重新连接 GitHub")
    
    elif warnings > 0:
        print("⚠️ 部分检查需要手动确认")
        print("请按照 GitHub连接问题解决指南.md 中的步骤操作")
    
    else:
        print("🎉 所有检查通过！")
        print("现在可以在 Render 中创建 Blueprint 了")
    
    print("\n📚 参考文档:")
    print("- GitHub连接问题解决指南.md")
    print("- 部署检查清单.md")
    
    return failed == 0

def main():
    """主函数"""
    print("🔧 Hospital Spider 自动诊断工具")
    print("正在检查系统状态...")
    
    try:
        success = generate_diagnostic_report()
        
        if success:
            print("\n🎉 诊断完成！系统状态正常，可以继续部署。")
        else:
            print("\n⚠️ 诊断完成！发现一些问题需要解决。")
        
        print("\n按 Enter 键退出...")
        input()
        
    except KeyboardInterrupt:
        print("\n\n⏹️ 诊断被用户中断")
    except Exception as e:
        print(f"\n❌ 诊断过程中发生错误: {e}")

if __name__ == "__main__":
    main() 