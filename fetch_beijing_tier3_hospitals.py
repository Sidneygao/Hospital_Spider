import requests
import json
import time

AMAP_KEY = '请填写你的高德KEY'  # TODO: 替换为你的高德API KEY
CITY = '北京'
KEYWORD = '三级甲等医院'
TYPECODE = '090101'
OUTPUT = 'beijing_tier3_hospitals.json'

all_pois = []
page = 1
while True:
    url = (
        f'https://restapi.amap.com/v3/place/text?key={AMAP_KEY}'
        f'&keywords={KEYWORD}&city={CITY}&types={TYPECODE}&citylimit=true&offset=25&page={page}'
    )
    resp = requests.get(url)
    data = resp.json()
    pois = data.get('pois', [])
    if not pois:
        break
    all_pois.extend(pois)
    print(f'已抓取第{page}页，累计{len(all_pois)}条')
    if len(pois) < 25:
        break
    page += 1
    time.sleep(0.5)  # 防止QPS过高

# 去重
unique = {}
for poi in all_pois:
    unique[poi['id']] = {
        'id': poi['id'],
        'name': poi['name'],
        'location': poi['location'],
        'typecode': poi['typecode'],
        'address': poi.get('address', ''),
        'tel': poi.get('tel', ''),
    }

with open(OUTPUT, 'w', encoding='utf-8') as f:
    json.dump(list(unique.values()), f, ensure_ascii=False, indent=2)

print(f'北京三甲医院POI名单已保存到 {OUTPUT}，共{len(unique)}条') 