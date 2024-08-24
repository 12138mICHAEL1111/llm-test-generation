import asyncio
import re
import json
from anthropic import AnthropicVertex

def extract_first_code_by_regex(completion):
    pattern = re.compile(r"```.*?\n(.*?)\n```", re.DOTALL)
    match = pattern.search(completion)
    if match:
        return match.group(1)
    return completion

def claude_chat(client, prompt):
    message = client.messages.create(
        max_tokens=4096,
        temperature=0.2,
        messages=[
            {
                "role": "user",
                "content": prompt,
            }
        ],
        model="claude-3-haiku@20240307",
    )
    message = message.content[0].text
    return message

async def process_prompt(client, checksum, prompt, completion_map, lock, semaphore,completed_tasks):
    if not prompt:
        return

    async with semaphore:  # 控制并发数量

        # 因为 claude_chat 是一个同步函数，我们需要使用 run_in_executor 来异步调用它
        loop = asyncio.get_running_loop()
        completion = await loop.run_in_executor(None, claude_chat, client, prompt)
    
    completion = extract_first_code_by_regex(completion)

    async with lock:  # 使用异步锁确保线程安全
        completion_map[checksum] = completion

    async with lock:
        completed_tasks[0] += 1
        print(f"已完成的任务数量: {completed_tasks[0]}")
            
async def generate_completions(client, prompt_map, workers):
    completion_map = {}
    completed_tasks = [0]
    lock = asyncio.Lock()
    semaphore = asyncio.Semaphore(workers)  # 控制并发任务数量

    # 创建异步任务列表，直接使用 asyncio 创建并发任务
    tasks = [
        process_prompt(client, checksum, prompt, completion_map, lock, semaphore,completed_tasks)
        for checksum, prompt in prompt_map.items()
    ]

    # 使用 asyncio.gather 并发运行所有任务
    await asyncio.gather(*tasks)

    return completion_map

def read_json_file(file_path):
    with open(file_path, 'r', encoding='utf-8') as file:
        data = json.load(file)  # 解析 JSON 文件内容为字典
    return data

async def main():
    workers = 80
    LOCATION = "us-central1"
    client = AnthropicVertex(region=LOCATION, project_id="gen-lang-client-0785814681")

    prompt_map = read_json_file("/Users/maike/Desktop/llm-test-generation/rq2_completion/fastjson/prompt.json")
    result = await generate_completions(client, prompt_map, workers)

    with open("/Users/maike/Desktop/llm-test-generation/rq2_completion/fastjson/claude/completion_2.json", 'w', encoding='utf-8') as file:
        json.dump(result, file, ensure_ascii=False, indent=4)

if __name__ == '__main__':
    asyncio.run(main())
