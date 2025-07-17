import os
import eyed3
import re

# 目标文件夹
folder = r"F:\\TEMP"

def clean_filename(s):
    # 替换Windows非法字符
    return re.sub(r'[\\/:*?"<>|]', '_', s)

for filename in os.listdir(folder):
    if filename.lower().endswith('.mp3'):
        filepath = os.path.join(folder, filename)
        print(f"处理文件: {filename}")
        try:
            audio = eyed3.load(filepath)
            if audio is None or audio.tag is None:
                print("  读取TAG失败")
                continue
            title = audio.tag.title
            artist = audio.tag.artist
            print(f"  读取到title: {title}, artist: {artist}")
            if title:  # 只处理标题不为空的
                artist = artist if artist else ""
                new_name = clean_filename(f"{title}{artist}.mp3")
                new_path = os.path.join(folder, new_name)
                print(f"  新文件名: {new_name}")
                # 避免重名覆盖
                if new_path != filepath and not os.path.exists(new_path):
                    os.rename(filepath, new_path)
                    print(f"  重命名成功: {filename} -> {new_name}")
                else:
                    print(f"  跳过（已存在或同名）: {new_name}")
            else:
                print(f"  跳过（无标题）: {filename}")
        except Exception as e:
            print(f"  处理出错: {filename}，原因: {e}") 