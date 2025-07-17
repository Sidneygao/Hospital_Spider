import os
import collections

folder = r"F:/TEMP"

names = [f for f in os.listdir(folder) if f.lower().endswith('.mp3')]
counter = collections.Counter(names)
dupes = [name for name, count in counter.items() if count > 1]

if dupes:
    print("重复文件:")
    for name in dupes:
        print(name)
else:
    print("无重复文件") 