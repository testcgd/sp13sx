Require vhs

Output "test/vhs/output/basic_chat.gif"
Set FontSize 14
Set Width 800
Set Height 400
Set Theme "catppuccin-mocha"
Set PlaybackSpeed 1.0
Set LoopOffset 0.5

# 启动程序（隐藏输入过程）
Hide
Type "./sp13sx -f ./config.local.yml"
Enter

# 等待程序完全启动（TUI 初始化需要时间）
Sleep 3s
Show

# 输入问题
Type "hello"
Enter

# 等待 LLM 响应（真实 LLM 需要时间）
Sleep 5s

# 退出
Type "/quit"
Enter
Sleep 1s
