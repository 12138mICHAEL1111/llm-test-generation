import re
import json

function_pattern = re.compile(r'^func(?:\s+\(\s*\w+\s+\*?\w+\s*\))?(\s+\w+)')

m = {}
with open("/Users/maike/Desktop/fastjson/parser.go", 'r') as file:
    func_lines = []

    for line in file:
        if line.startswith('func '):
            line = line.strip()
            match = function_pattern.search(line)
            if match:
                function_name = match.group(1).strip()
                m[function_name] = line[:len(line)-1]
                
with open("package_Info/fastjson/funSig_2.json",'w') as file:
    json.dump(m, file, indent=4)
