Require vhs

Output "test/vhs/output/real_llm.gif"
Set FontSize 14
Set Width 800
Set Height 400
Set Theme "catppuccin-mocha"
Set PlaybackSpeed 1.0

# 使用真实 LLM 测试（需要设置 PRIVATE_OPENAI_API_KEY 环境变量）
Hide
Type "./sp13sx -f ./config.local.yml"
Enter

# 等待 TUI 完全启动
Sleep 4s
Show

# 输入问题
Type "你好，请用一句话介绍你自己"
Enter

# 等待 LLM 响应（真实 LLM 需要 5-15 秒）
Sleep 10s

# 退出
Type "/quit"
Enter
Sleep 1s
