import re
import subprocess
import json

from tqdm import tqdm

def extract_test_function_names(filepath):
    pattern = re.compile(r"func (Test\w+)\(")

    test_function_names = []

    with open(filepath, "r") as file:
        content = file.read()

    matches = pattern.findall(content)
    for match in matches:
        test_function_names.append(match)

    return test_function_names

# assert fails
def collect_assert(result, error_json, test_function):
    err_msg = ""
    pattern = re.compile(
        r".*?Error:\s*(.*?)\s*"
        r"expected:\s*(.*?)\s*"
        r"actual\s*:\s*(.*?)\s*"
        r"Test:\s*(.*?)\n",
        re.DOTALL,
    )

    matches = pattern.findall(result)
    for match in matches:
        reason, expected, actual, test_reference = match
        test_case = test_reference.split("/")
        if len(test_case) > 1:
            err_msg = (
                err_msg
                + f"Test case {test_case[-1]} fails, {reason} Expected: {expected}, Actual: {actual}. "
            )
        else:
            err_msg = f"{reason} Expected: {expected}, Actual: {actual}."

    error_json[test_function] = err_msg


# terror fails
def collect_terror(result, error_json, test_function):
    err_msg = ""
    pattern = re.compile(
        r"(?:--- FAIL: ([\w/]+) \(.*?\)\s*)?" r"\w+\.go:\d+: (.+?)\n", re.DOTALL
    )
    matches = pattern.findall(result)
    for match in matches:
        test_name, error_message = match
        if not test_name:
            err_msg = err_msg + f"{error_message}. "
        else:
            test_case = test_name.split("/")
            if len(test_case) > 1:
                err_msg = err_msg + f"Test case {test_case[-1]} fails, {error_message} "
            else:
                err_msg = f"{error_message} ."
        
    if err_msg != "":       
        error_json[test_function] = err_msg


def collect_panic(result, error_json, test_function):
    err_msg = ""
    pattern = re.compile(
        r"(?:--- FAIL: ([\w/]+) \(.*?\)\s*)?" r"(panic:.*?)(?=\s*\[recovered\])",
        re.DOTALL,
    )
    matches = pattern.findall(result)
    for match in matches:
        test_name, error_message = match
        if not test_name:
            err_msg = err_msg + f"{error_message}. "
        else:
            test_case = test_name.split("/")
            if len(test_case) > 1:
                err_msg = err_msg + f"Test case {test_case[-1]} fails, {error_message} "
            else:
                err_msg = f"{error_message} ."

    if test_function not in error_json:
        error_json[test_function] = err_msg
    else:
        error_json[test_function] = error_json[test_function] + err_msg

def collect_timeout(result,err_json,test_function):
    if "panic: test timed out" in result:
        err_json[test_function] = result
        return True
    return False

def collect_error(test_function_list,error_json,command):
    for test_function in tqdm(test_function_list):
        command[3] = "^" + test_function+ "$"
        result = subprocess.run(command, capture_output=True, text=True).stdout
        if result[:2] == "ok":
            continue

        index = result.find("\n")
        if index == -1:
            continue
        
        r = collect_timeout(result[:index],error_json,test_function)
        if r:
            continue
        
        result = result[index + 1 :]
        if "\tError Trace" in result:
            collect_assert(result, error_json, test_function)
        else:
            collect_terror(result, error_json, test_function)

        collect_panic(result, error_json, test_function)
        msg = error_json[test_function]
        msg = f"For {test_function}: " + msg
      

def count_fixed(error_json):
    try:
        with open("error.json", 'r') as file:
            old_error_json = json.load(file)
    except:
        return
    
    num_fixed = 0 
    for old_error in old_error_json:
        if old_error not in error_json:
            num_fixed += 1
    
    print("number fixed", num_fixed)
    

if __name__ == "__main__":
    error_json = {}

    test_function_list = extract_test_function_names("db_test.go")
    # print(test_function_list)
    # test_function_list = ["TestGoString"]
    command = [
        "go",
        "test",
        "-run",
        "^TestEqualApprox$",
        "db_test.go",
        "-timeout",
        "10s"
    ]

    collect_error(test_function_list,error_json,command)
    
    count_fixed(error_json)
    
    print("number of failing tests",len(error_json))
    with open('error.json', 'w') as json_file:
        json.dump(error_json, json_file, indent=4)
        


