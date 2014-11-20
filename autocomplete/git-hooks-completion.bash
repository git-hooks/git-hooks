# bash completion support for git-hooks.

_git_hooks() {
    local subcommands="$(git hooks --generate-bash-completion)"
	local subcommand="$(__git_find_on_cmdline "$subcommands")"
	if [ -z "$subcommand" ]; then
		__gitcomp "$subcommands"
		return
	fi
}
