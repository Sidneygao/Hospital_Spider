import json
import pandas as pd
import math
import os

print('模拟协和合并脚本启动')

def haversine(lat1, lng1, lat2, lng2):
    R = 6371000
    phi1, phi2 = math.radians(lat1), math.radians(lat2)
    dphi = math.radians(lat2 - lat1)
    dlambda = math.radians(lng2 - lng1)
    a = math.sin(dphi/2)**2 + math.cos(phi1)*math.cos(phi2)*math.sin(dlambda/2)**2
    return 2*R*math.asin(math.sqrt(a))

def lcs(s1, s2):
    m, n = len(s1), len(s2)
    dp = [[0]*(n+1) for _ in range(m+1)]
    maxlen, end = 0, 0
    for i in range(1, m+1):
        for j in range(1, n+1):
            if s1[i-1] == s2[j-1]:
                dp[i][j] = dp[i-1][j-1]+1
                if dp[i][j] > maxlen:
                    maxlen = dp[i][j]
                    end = i
    return s1[end-maxlen:end] if maxlen >= 4 else ''

def is_xiehe(name):
    return '协和医院' in name

def parse_latlng(poi):
    loc = poi.get('location')
    if not loc:
        return None, None
    try:
        lng, lat = map(float, loc.split(','))
        return lat, lng
    except:
        return None, None

# 读取指定缓存
cache_path = 'backend/cache/cache_116.417671,39.920235_5000_hospitals.json'
if not os.path.exists(cache_path):
    print(f'文件不存在: {cache_path}')
    exit(1)
with open(cache_path, 'r', encoding='utf-8') as f:
    data = json.load(f)
pois = data.get('pois', data)
print(f'总POI数: {len(pois)}')
print('前10个POI名称:')
for p in pois[:10]:
    print('  ', p.get('name', '无名'))

# 只取090100/090101
pois_0901 = [p for p in pois if str(p.get('typecode', '')).startswith('090100') or str(p.get('typecode', '')).startswith('090101')]
print(f'090100/090101相关POI数: {len(pois_0901)}')

# 找出所有包含“协和医院”的POI
xiehe_pois = [p for p in pois_0901 if '协和医院' in p.get('name', '')]
print(f'协和相关POI数: {len(xiehe_pois)}')
if xiehe_pois:
    print('协和相关POI名称:')
    for p in xiehe_pois:
        print('  ', p.get('name'))
else:
    print('无协和相关POI')

# 模拟合并：最长公共子串为“北京协和医院”且距离<300米
main_name = '北京协和医院'
merged_group = []
for i, p1 in enumerate(xiehe_pois):
    name1 = p1.get('name', '')
    lat1, lng1 = parse_latlng(p1)
    if not name1 or lat1 is None:
        continue
    if main_name in name1:
        for j, p2 in enumerate(xiehe_pois):
            if i == j:
                continue
            name2 = p2.get('name', '')
            lat2, lng2 = parse_latlng(p2)
            if not name2 or lat2 is None:
                continue
            if main_name in name2:
                dist = haversine(lat1, lng1, lat2, lng2)
                if dist < 300:
                    if p1 not in merged_group:
                        merged_group.append(p1)
                    if p2 not in merged_group:
                        merged_group.append(p2)
print(f'合并分组数: {len(merged_group)}')
if merged_group:
    print('合并分组POI名称:')
    for p in merged_group:
        print('  ', p.get('name'))
    # 生成新主POI
    new_poi = merged_group[0].copy()
    new_poi['name'] = main_name
    new_poi['tel'] = ''
    new_poi['childtype'] = []
    new_poi['hospital_category'] = '三甲/综合医院合并主名'
    print('新主POI:', new_poi)
    # 被合并POI类别置为None
    for p in merged_group:
        p['hospital_category'] = None
    print('被合并POI类别字段:')
    for p in merged_group:
        print('  ', p.get('name'), p.get('hospital_category'))
else:
    print('无可合并协和POI') 