import sys
import json
import time
from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.chrome.options import Options
import argparse

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--lat', type=float, required=True)
    parser.add_argument('--lng', type=float, required=True)
    args = parser.parse_args()

    chrome_options = Options()
    chrome_options.add_argument("--headless")
    chrome_options.add_argument("--disable-gpu")
    driver = webdriver.Chrome(options=chrome_options)

    url = f"https://www.google.com/maps/search/医院/@{args.lat},{args.lng},14z"
    driver.get(url)
    time.sleep(5)

    hospitals = []
    items = driver.find_elements(By.CSS_SELECTOR, 'div[role="article"]')[:10]
    for item in items:
        try:
            name = item.get_attribute('aria-label') or item.text
            hospitals.append({
                "name": name,
                "lat": args.lat,  # 实际应解析真实坐标
                "lng": args.lng,
            })
        except Exception:
            continue

    # 热力点统计
    heatmap = {}
    for h in hospitals:
        key = (round(h["lat"], 3), round(h["lng"], 3))
        heatmap[key] = heatmap.get(key, 0) + 1
    heatmap_data = [{"lat": k[0], "lng": k[1], "weight": v} for k, v in heatmap.items()]

    driver.quit()
    print(json.dumps({"hospitals": hospitals, "heatmap": heatmap_data}, ensure_ascii=False))

if __name__ == "__main__":
    main() 