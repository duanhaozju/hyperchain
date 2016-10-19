#!/usr/bin/env bash
tmux_init()
{
	tmux kill-session hyperchain
	tmux new-session -s "hyperchain" -d -n "local"    # 开启一个会话
	tmux new-window -n "other"
	tmux split-window -v
	tmux select-pane -t 1
	tmux split-window -h
	tmux select-pane -t 3
	tmux split-window -h
	sleep 2
	for _pane in $(tmux list-panes  -t hyperchain -F '#P'); do
    tmux send-keys -t ${_pane} "top" Enter
	done
    tmux -2 attach-session -d           # tmux -2强制启用256color，连接已开启的tmux
}

# 判断是否已有开启的tmux会话，没有则开启
if which tmux 2>&1 >/dev/null; then
    test -z "$TMUX" && (tmux attach || tmux_init)
fi