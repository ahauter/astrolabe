llama-server \
  -hf Qwen/Qwen2.5-Coder-0.5B-Instruct-GGUF:Q8_0 \
  -ngl 99 -t 28 -dev CUDA0 -c 2048 -n 64
