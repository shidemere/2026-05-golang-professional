#!/bin/sh

# This helper has no side effects: it only prints its arguments and selected
# environment variables. The ${VAR+x} check distinguishes an unset variable
# from a variable with an empty value.
show_var() {
	name=$1
	eval 'is_set=${'"$name"'+x}'
	if [ "$is_set" = x ]; then
		eval 'value=${'"$name"'}'
		printf '%s=<%s>\n' "$name" "$value"
	else
		printf '%s=<UNSET>\n' "$name"
	fi
}

printf 'args=<%s>\n' "$*"
show_var HELLO
show_var BAR
show_var FOO
show_var EMPTY
show_var UNSET
