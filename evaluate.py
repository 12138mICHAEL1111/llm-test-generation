from error import error_msg

def evaluate():
    supportive_func = 0
    test_func = 0
    with open('test_function.txt', 'r') as file:
        for line in file:
            if line.startswith("func"):
                if line.startswith("func Test"):
                    test_func+=1
                else:
                    supportive_func +=1
    print("test func num: ", test_func)
    print("supportive func num", supportive_func)

    error_map = {}
    for error in error_msg:
        if error["source"] == "compiler":
            error_type = error["code"]["value"]
            if error_type not in error_map:
                error_map[error_type] = 1
            else:
                error_map[error_type] += 1
    
    print("\nnumber of error: ",sum(error_map.values()))
    for k in error_map:
        print(k,error_map[k])
    
        
if __name__ == "__main__":
    evaluate()