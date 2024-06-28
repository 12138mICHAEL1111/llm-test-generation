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
        # print("-----------")
        # print(test_name,error_message)
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


def collect_error(test_function_list,error_json,command):
    for test_function in tqdm(test_function_list):
        command[3] = "^" + test_function+ "$"
        result = subprocess.run(command, capture_output=True, text=True).stdout
        print(result)
        if result[:2] == "ok":
            continue

        index = result.find("\n")
        if index != -1:
            result = result[index + 1 :]
        else:
            print("No newline character found in the text.")
            
        if "\tError Trace" in result:
            collect_assert(result, error_json, test_function)
        else:
            collect_terror(result, error_json, test_function)

        collect_panic(result, error_json, test_function)
        msg = error_json[test_function]
        msg = f"For {test_function}: " + msg
        # remove last white space
        error_json[test_function] = msg[: len(msg) - 1]


if __name__ == "__main__":
    error_json = {}

    # test_function_list = extract_test_function_names("floats/floats_test.go")
    test_function_list = ["TestAdd"]
    command = [
        "go",
        "test",
        "-run",
        "^TestEqualApprox$",
        "/Users/maike/Desktop/gonum/floats/floats_test.go",
        "/Users/maike/Desktop/gonum/floats/floats.go",
    ]

    collect_error(test_function_list,error_json,command)
    
    with open('failed_error.json', 'w') as json_file:
        json.dump(error_json, json_file, indent=4)
        


