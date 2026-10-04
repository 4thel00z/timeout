#!/usr/bin/env bash
set -u

prompt() {
	printf '\033[35m❯\033[0m '
}

typed() {
	local text=$1
	for ((i = 0; i < ${#text}; i++)); do
		printf '%s' "${text:i:1}"
		sleep 0.04
	done
}

run() {
	prompt
	typed "$1"
	sleep 0.4
	printf '\n'
	eval "$1"
	sleep 1.5
}

clear
run "timeout 5 echo 'done in time'; echo exit=\$?"
run "timeout 1 sleep 10; echo exit=\$?"
run "timeout -v -s INT 1.5 sleep 10; echo exit=\$?"
run "timeout -v -k 1 1 sh -c \"trap '' TERM; sleep 10\"; echo exit=\$?"
run "timeout --preserve-status 1 sleep 10; echo exit=\$?"
prompt
sleep 2
