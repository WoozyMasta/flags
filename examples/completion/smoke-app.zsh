#compdef smoke-app

_smoke_app() {
	local -a completions
	local IFS=$'\n'
	local cur="${words[CURRENT]}"

	completions=($(GO_FLAGS_COMPLETION=1 "${words[@]}"))
	(( ${#completions} )) || return 1

	if [[ "$cur" == --*=* || "$cur" == -*=* ]]; then
		compadd -S '' -- "${completions[@]}"
	else
		compadd -- "${completions[@]}"
	fi

	return 0
}

compdef _smoke_app 'smoke-app'
