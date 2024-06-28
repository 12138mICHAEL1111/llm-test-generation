import re
import json

from tqdm import tqdm

def find_nearest_function(filepath, line_number):
    # 读取文件内容
    with open(filepath, 'r') as file:
        lines = file.readlines()
    
    # 正则表达式匹配 Go 函数定义，同时捕获函数名称
    function_pattern = re.compile(r'^func(?:\s+\(\s*\w+\s+\*?\w+\s*\))?(\s+\w+)')

    # 从指定行号开始向上搜索
    for i in range(line_number - 1, -1, -1):  # line_number - 1 因为行号通常是从1开始，而列表索引从0开始
        match = function_pattern.search(lines[i])
        if match:
            function_name = match.group(1).strip()  # 获取函数名称
            return function_name  # 返回行号，函数定义和函数名称

    return None

def get_line_from_file(filepath, line_number):
    # 使用 with 语句确保文件操作后自动关闭文件
    with open(filepath, 'r') as file:
        lines = file.readlines()
    
    return lines[line_number - 1].strip()


if __name__ == "__main__":
    test_file_path = "db_test.go"
    error_json = {}
    
    with open('ide_error.json', 'r') as file:
        data = json.load(file)
        
    errors = data["errors"]
    
    for error in tqdm(errors):
        line = get_line_from_file(test_file_path, error["startLineNumber"])
        targetFunction = find_nearest_function(test_file_path, error["startLineNumber"])
        msg = error["message"]
        if targetFunction not in error_json:
            error_json[targetFunction] = f"For code {line}, {msg}. "
        else:
            error_json[targetFunction] = error_json[targetFunction] + f"For code {line}, {msg}. "

    for k in error_json:
        error_json[k] = error_json[k][:len(error_json[k])-1]
        
    with open('compilation_error.json', 'w') as json_file:
        json.dump(error_json, json_file, indent=4)
        

