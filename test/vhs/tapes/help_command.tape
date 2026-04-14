Require vhs

Output "test/vhs/output/help_command.gif"
Set FontSize 14
Set Width 800
Set Height 400
Set Theme "catppuccin-mocha"
Set PlaybackSpeed 1.0

Hide
Type "./sp13sx -f ./config.local.yml"
Enter
Sleep 3s
Show

Type "/help"
Enter
Sleep 2s

Type "/quit"
Enter
Sleep 1s
