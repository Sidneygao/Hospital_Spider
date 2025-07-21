import requests
import json
import time
import os

AMAP_KEY = os.environ.get('AMAP_KEY')
if not AMAP_KEY:
    raise Exception('请先设置环境变量AMAP_KEY')
CITY = '北京'
KEYWORD = '三甲医院'  # 可改为 '三级甲等医院' 试试
OUTPUT = 'beijing_tier3_hospitals_by_keyword.json'

all_pois = []
page = 1
while True:
    url = (
        f'https://restapi.amap.com/v3/place/text?key={AMAP_KEY}'
        f'&keywords={KEYWORD}&city={CITY}&citylimit=true&offset=25&page={page}'
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
    time.sleep(0.5)

# 去重
unique = {}
for poi in all_pois:
    unique[poi['id']] = {
        'id': poi['id'],
        'name': poi['name'],
        'location': poi['location'],
        'type': poi['type'],
        'typecode': poi['typecode'],
        'address': poi['address'],
        'tel': poi['tel'],
    }

with open(OUTPUT, 'w', encoding='utf-8') as f:
    json.dump(list(unique.values()), f, ensure_ascii=False, indent=2)

print(f'最终去重后共{len(unique)}条，已保存到 {OUTPUT}')
print('前10条POI示例:')
for i, poi in enumerate(list(unique.values())[:10]):
    print(f"{i+1}. {poi['name']} | {poi['address']} | {poi['id']}") 