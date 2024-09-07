## Evaluation of Go Unit Test Generation

### File Structure

* GPT_generation and gemini_generation folders store the chat history of RQ1 test generations and fixing

* package_Info stores the code context to the 3 levels

* rq2_completions stores the prompts and corresponding completions

* llm-test-generation.go (GPT) and gemini.go are the main files for running RQ1 

* rq2_claude.py, rq2_evaluate.ipynb and rq2.go are main files for RQ2

* config.go stores the default configurations like the temperature and base prompts 