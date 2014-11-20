# zsh completion support for git-hooks.

_git-hooks ()
{
	local curcontext="$curcontext" state line
	typeset -A opt_args

	_arguments -C \
		':command:->command' \
		'*::options:->options'

	case $state in
		(command)
			local -a subcommands
            subcommands=("${(@f)$(git hooks --generate-bash-completion)}")
			_describe -t commands 'git hooks ' subcommands
		;;

		(options)
		;;
	esac
}

zstyle ':completion:*:*:git:*' user-commands hooks:'description for foo'
