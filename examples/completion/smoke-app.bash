_smoke_app() {
	args=("${COMP_WORDS[@]:1:$COMP_CWORD}")
	cur="${COMP_WORDS[COMP_CWORD]}"

	mapfile -t COMPREPLY < <(GO_FLAGS_COMPLETION=1 ${COMP_WORDS[0]} "${args[@]}")

	if [[ "$cur" == --*=* || "$cur" == -*=* ]]; then
		compopt -o nospace 2>/dev/null
	fi

	return 0
}

complete -F _smoke_app smoke-app
