import re
import json

function_pattern = re.compile(r'^func(?:\s+\(\s*\w+\s+\*?\w+\s*\))?(\s+\w+)')

m = {}
with open("/Users/maike/Desktop/fastjson/parser.go", 'r') as file:
    # 初始化一个列表，用于存储以"func "开头的行
    func_lines = []

    # 逐行读取文件内容
    for line in file:
        # 检查当前行是否以"func "开头
        if line.startswith('func '):
            line = line.strip()
            match = function_pattern.search(line)
            if match:
                function_name = match.group(1).strip()
                m[function_name] = line[:len(line)-1]
                
with open("package_Info/fastjson/funSig_2.json",'w') as file:
    json.dump(m, file, indent=4)
