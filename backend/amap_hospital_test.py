import os
import requests
import json
import time

# 高德API KEY从环境变量读取
AMAP_KEY = os.getenv('AMAP_KEY')
if not AMAP_KEY:
    raise Exception('请先在环境变量中设置AMAP_KEY')

# 测试医院及坐标
TEST_CASES = [
    {
        'name': '北京协和医院',
        'location': '116.417345,39.913249',  # 协和医院东单院区
        'typecodes': ['090000', '090101', '091000', '091001', '091002'],
        'keywords': ['协和医院', '北京协和医院']
    },
    {
        'name': '北京同仁医院',
        'location': '116.418261,39.899411',  # 同仁医院东区
        'typecodes': ['090000', '090101', '091000', '091001', '091002'],
        'keywords': ['同仁医院', '北京同仁医院']
    }
]

# 查询半径
RADIUS = 3000

# 高德POI周边搜索
def search_around(location, typecode=None, keyword=None):
    url = 'https://restapi.amap.com/v3/place/around'
    params = {
        'key': AMAP_KEY,
        'location': location,
        'radius': RADIUS,
        'offset': 25,
        'page': 1,
        'extensions': 'all',
        'output': 'JSON',
    }
    if typecode:
        params['types'] = typecode
    if keyword:
        params['keywords'] = keyword
    try:
        resp = requests.get(url, params=params, timeout=10)
        data = resp.json()
        return data
    except Exception as e:
        return {'error': str(e)}

# 主测试流程
if __name__ == '__main__':
    output_lines = []
    for case in TEST_CASES:
        output_lines.append(f'\n==== 测试: {case["name"]} ====' )
        # 1. 090000主类
        output_lines.append('\n[090000主类]')
        data = search_around(case['location'], typecode='090000')
        if 'error' in data:
            output_lines.append(f'请求异常: {data["error"]}')
        else:
            pois = data.get('pois', [])
            output_lines.append(f'返回POI数量: {len(pois)}')
            if pois:
                for poi in pois:
                    output_lines.append(f'- {poi.get("name")} ({poi.get("typecode")})')
            else:
                output_lines.append('无POI返回')
        time.sleep(1)
        # 2. 0910xx等特殊typecode
        for tc in case['typecodes']:
            if tc == '090000':
                continue
            output_lines.append(f'\n[{tc} 精确typecode]')
            data = search_around(case['location'], typecode=tc)
            if 'error' in data:
                output_lines.append(f'请求异常: {data["error"]}')
            else:
                pois = data.get('pois', [])
                output_lines.append(f'返回POI数量: {len(pois)}')
                if pois:
                    for poi in pois:
                        output_lines.append(f'- {poi.get("name")} ({poi.get("typecode")})')
                else:
                    output_lines.append('无POI返回')
            time.sleep(1)
        # 3. 关键字兜底
        for kw in case['keywords']:
            output_lines.append(f'\n[关键字: {kw}]')
            data = search_around(case['location'], keyword=kw)
            if 'error' in data:
                output_lines.append(f'请求异常: {data["error"]}')
            else:
                pois = data.get('pois', [])
                output_lines.append(f'返回POI数量: {len(pois)}')
                if pois:
                    for poi in pois:
                        output_lines.append(f'- {poi.get("name")} ({poi.get("typecode")})')
                else:
                    output_lines.append('无POI返回')
            time.sleep(1)
    output_lines.append('\n测试完成。')
    with open('result_amap_test.txt', 'w', encoding='utf-8') as f:
        for line in output_lines:
            print(line)
            f.write(line + '\n') 