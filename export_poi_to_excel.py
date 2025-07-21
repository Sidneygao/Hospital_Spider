import json
import pandas as pd
import os
import sys
import subprocess

# 读取JSON
with open('backend/cache/merged_poi_result.json', 'r', encoding='utf-8') as f:
    data = json.load(f)

pois = data.get('pois', [])

# 扁平化处理嵌套字段
for poi in pois:
    # 将biz_ext、photos等嵌套字段转为字符串，便于Excel查看
    for k in ['biz_ext', 'photos', 'poiweight', 'importance', 'biz_type', 'shopid', 'shopinfo', 'tel']:
        if k in poi:
            poi[k] = str(poi[k])

# 导出为Excel
excel_path = 'backend/cache/poi_export.xlsx'
df = pd.DataFrame(pois)
df.to_excel(excel_path, index=False)

# 自动打开Excel
if sys.platform.startswith('win'):
    os.startfile(excel_path)
else:
    subprocess.call(['open', excel_path]) 