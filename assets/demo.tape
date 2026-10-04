Output assets/demo.gif

Require timeout

Set Shell "bash"
Set FontSize 20
Set Width 1100
Set Height 520
Set Padding 28
Set Margin 24
Set MarginFill "#6b50ff"
Set WindowBar Colorful
Set BorderRadius 10
Set Theme "Catppuccin Mocha"
Set TypingSpeed 40ms
Set WaitPattern /❯$/
Set WaitTimeout 30s

Hide
Sleep 1s
Type `export PS1='\[\e[35m\]❯\[\e[0m\] '` Enter
Wait
Type "clear" Enter
Wait
Show

Type `timeout 5 echo 'done in time'; echo exit=$?`
Sleep 300ms
Enter
Wait
Sleep 1.5s

Type `timeout 1 sleep 10; echo exit=$?`
Sleep 300ms
Enter
Wait
Sleep 1.5s

Type `timeout -v -s INT 1.5 sleep 10; echo exit=$?`
Sleep 300ms
Enter
Wait
Sleep 1.5s

Type `timeout -v -k 1 1 sh -c "trap '' TERM; sleep 10"; echo exit=$?`
Sleep 300ms
Enter
Wait
Sleep 1.5s

Type `timeout --preserve-status 1 sleep 10; echo exit=$?`
Sleep 300ms
Enter
Wait
Sleep 3s
