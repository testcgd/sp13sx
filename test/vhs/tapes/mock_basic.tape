Require vhs

Output "test/vhs/output/mock_basic.gif"
Set FontSize 14
Set Width 800
Set Height 400
Set Theme "catppuccin-mocha"
Set PlaybackSpeed 1.0

# 使用 Mock 后端测试
Hide
Env SP13SX_TEST_MODE mock
Env SP13SX_SCENARIO basic_chat
Type "./sp13sx -f ./config.local.yml"
Enter

# 等待 TUI 启动
Sleep 3s
Show

# 输入消息
Type "hello"
Enter

# 等待 Mock 响应
Sleep 2s

# 退出
Type "/quit"
Enter
Sleep 1s
