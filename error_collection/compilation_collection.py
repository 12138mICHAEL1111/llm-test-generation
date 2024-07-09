import re
import json

from tqdm import tqdm

def count_func_num(filepath):
    with open(filepath, 'r') as file:
        content = file.read()
    function_pattern = re.compile(r'^func(?:\s+\(\s*\w+\s+\*?\w+\s*\))?(\s+\w+)', re.MULTILINE)
    matches = function_pattern.findall(content)
    test_func_num = 0
    for match in matches:
        if "Test" in match:
            test_func_num +=1
    print("Number of valid functions:", len(matches))
    print("Test functions", test_func_num)
    
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
      
def count_fixed(error_json):
    try:
        with open("error.json", 'r') as file:
            old_error_json = json.load(file)
    except:
        return
    
    num_fixed = 0 
    for old_error in old_error_json:
        if old_error not in error_json:
            print(old_error)
            num_fixed += 1
    
    print("number fixed", num_fixed)

if __name__ == "__main__":
    test_file_path = "db_test.go"
    count_func_num(test_file_path)
    error_json = {}
    
    with open('ide_error.json', 'r') as file:
        errors = json.load(file)
    
    print("number of compilation erros:", len(errors))
    
    for error in tqdm(errors):
        line = get_line_from_file(test_file_path, error["startLineNumber"])
        target_func = find_nearest_function(test_file_path, error["startLineNumber"])
        msg = error["message"]
        if target_func not in error_json:
            error_json[target_func] = f"{target_func}: For code {line}, {msg}. "
        else:
            error_json[target_func] = error_json[target_func] + f"For code {line}, {msg}. "

    print("number of error funcs:", len(error_json))
    
    count_fixed(error_json)
    
    with open('error.json', 'w') as json_file:
        json.dump(error_json, json_file, indent=4)
        

        

