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

    async with semaphore:  

       
        loop = asyncio.get_running_loop()
        completion = await loop.run_in_executor(None, claude_chat, client, prompt)
    
    completion = extract_first_code_by_regex(completion)

    async with lock:  
        completion_map[checksum] = completion

    async with lock:
        completed_tasks[0] += 1
        print(f"task: {completed_tasks[0]}")
            
async def generate_completions(client, prompt_map, workers):
    completion_map = {}
    completed_tasks = [0]
    lock = asyncio.Lock()
    semaphore = asyncio.Semaphore(workers)  

    tasks = [
        process_prompt(client, checksum, prompt, completion_map, lock, semaphore,completed_tasks)
        for checksum, prompt in prompt_map.items()
    ]

    await asyncio.gather(*tasks)

    return completion_map

def read_json_file(file_path):
    with open(file_path, 'r', encoding='utf-8') as file:
        data = json.load(file) 
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
