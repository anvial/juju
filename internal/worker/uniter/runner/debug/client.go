// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package debug

import (
	"encoding/base64"
	"strings"

	goyaml "gopkg.in/yaml.v2"
)

type hookArgs struct {
	Hooks   []string `yaml:"hooks,omitempty"`
	DebugAt string   `yaml:"debug-at,omitempty"`
}

// ClientScript returns a bash script suitable for executing
// on the unit system to intercept matching hooks or actions via tmux shell.
func ClientScript(c *HooksContext, match []string, debugAt string) string {
	// If any argument is "*", then the client is interested in all.
	for _, m := range match {
		if m == "*" {
			match = nil
			break
		}
	}

	s := strings.Replace(debugHooksClientScript, "{unit_name}", c.Unit, -1)
	s = strings.Replace(s, "{tmux_conf}", tmuxConf, 1)
	s = strings.Replace(s, "{entry_flock}", c.ClientFileLock(), -1)
	s = strings.Replace(s, "{exit_flock}", c.ClientExitFileLock(), -1)

	base64Args := Base64HookArgs(match, debugAt)
	s = strings.Replace(s, "{hook_args}", base64Args, 1)
	return s
}

// Base64HookArgs returns the encoded arguments for defining debug-hook behavior.
// This is a base64 encoded yaml blob containing serialized arguments.
func Base64HookArgs(match []string, debugAt string) string {
	yamlArgs := encodeArgs(match, debugAt)
	return base64.StdEncoding.EncodeToString(yamlArgs)
}

func encodeArgs(args []string, debugAt string) []byte {
	// Marshal to YAML, then encode in base64 to avoid shell escapes.
	yamlArgs, err := goyaml.Marshal(hookArgs{Hooks: args, DebugAt: debugAt})
	if err != nil {
		// This should not happen: we're in full control.
		panic(err)
	}
	return yamlArgs
}

const debugHooksClientScript = `#!/bin/bash
(
cleanup_on_exit() 
{ 
	echo "Cleaning up the debug session"
	tmux kill-session -t {unit_name}; 
}
trap cleanup_on_exit EXIT

# Lock the juju-<unit>-debug lockfile.
flock -n 8 || (
	echo "Found existing debug sessions, attempting to reconnect" 2>&1
	exec tmux attach-session -t {unit_name}
	exit $?
	)
(
# Close the inherited lock FD, or tmux will keep it open.
exec 8>&-

# Write out the debug-hooks args.
echo "{hook_args}" | base64 -d > {entry_flock}

# Lock the juju-<unit>-debug-exit lockfile.
flock -n 9 || exit 1

# Wait for tmux to be installed.
while [ ! -f /usr/bin/tmux ]; do
    sleep 1
done

if [ ! -f ~/.tmux.conf ]; then
	if [ -f /usr/share/byobu/profiles/tmux ]; then
		# Use byobu/tmux profile for familiar keybindings and branding
		echo "source-file /usr/share/byobu/profiles/tmux" > ~/.tmux.conf
	else
		# Otherwise, use the legacy juju/tmux configuration
		cat > ~/.tmux.conf <<END
{tmux_conf}
END
	fi
fi

(
    # Close the inherited lock FD, or tmux will keep it open.
    exec 9>&-
	# Since we just use byobu tmux configs without byobu-tmux, we need
	# to export this to prevent the TERM being set to empty string.
	export BYOBU_TERM=$TERM
    if ! tmux has-session -t {unit_name}; then
		tmux new-session -d -s {unit_name}
	fi
	client_count=$(tmux list-clients | wc -l)
	if [ $client_count -ge 1 ]; then
		session_name={unit_name}"-"$client_count
		tmux new-session -d -t {unit_name} -s $session_name
		exec tmux attach-session -t $session_name \; set-option destroy-unattached
	else
	    exec tmux attach-session -t {unit_name}
	fi
)
) 9>{exit_flock}
) 8>{entry_flock}
exit $?
`

const tmuxConf = `
# Status bar
set-option -g status-bg black
set-option -g status-fg white

set-window-option -g window-status-current-bg red
set-window-option -g window-status-current-attr bright

set-option -g status-right ''

# Panes
set-option -g pane-border-fg white
set-option -g pane-active-border-fg white

# Monitor activity on windows
set-window-option -g monitor-activity on

# Screen bindings, since people are more familiar with that.
set-option -g prefix C-a
bind C-a last-window
bind a send-key C-a

bind | split-window -h
bind - split-window -v

# Fix CTRL-PGUP/PGDOWN for vim
set-window-option -g xterm-keys on

# Prevent ESC key from adding delay and breaking Vim's ESC > arrow key
set-option -s escape-time 0
`
