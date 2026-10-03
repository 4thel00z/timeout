Output .demo-frames/

Set Shell "bash"
Set FontSize 20
Set Width 1100
Set Height 560
Set Theme "Catppuccin Mocha"
Set TypingSpeed 45ms

Hide
Type `export PS1='\[\e[35m\]❯\[\e[0m\] '` Enter
Type "clear" Enter
Show

Type `timeout 5 echo 'done in time'; echo exit=$?` Enter
Sleep 1.2s

Type `timeout 1 sleep 10; echo exit=$?` Enter
Sleep 2s

Type `timeout -v -s INT 1.5 sleep 10; echo exit=$?` Enter
Sleep 2.5s

Type `timeout -v -k 1 1 sh -c "trap '' TERM; sleep 10"; echo exit=$?` Enter
Sleep 3s

Type `timeout --preserve-status 1 sleep 10; echo exit=$?` Enter
Sleep 2.5s
