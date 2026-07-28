package commands

import (
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	bashCompletionFunc = `
__cd_list_apps() {
	local -a cd_out
	if cd_out=($(cd app list --output name 2>/dev/null)); then
		COMPREPLY+=( $( compgen -W "${cd_out[*]}" -- "$cur" ) )
	fi
}

__cd_list_app_history() {
	local app=$1
	local -a cd_out
	if cd_out=($(cd app history $app --output id 2>/dev/null)); then
		COMPREPLY+=( $( compgen -W "${cd_out[*]}" -- "$cur" ) )
	fi
}

__cd_app_rollback() {
	local -a command
	for comp_word in "${COMP_WORDS[@]}"; do
		if [[ $comp_word =~ ^-.*$ ]]; then
			continue
		fi
		command+=($comp_word)
	done

	# fourth arg is app (if present): e.g.- cd app rollback guestbook
	local app=${command[3]}
	local id=${command[4]}
	if [[ -z $app || $app == $cur ]]; then
		__cd_list_apps
	elif [[ -z $id || $id == $cur ]]; then
		__cd_list_app_history $app
	fi
}

__cd_list_servers() {
	local -a cd_out
	if cd_out=($(cd cluster list --output server 2>/dev/null)); then
		COMPREPLY+=( $( compgen -W "${cd_out[*]}" -- "$cur" ) )
	fi
}

__cd_list_repos() {
	local -a cd_out
	if cd_out=($(cd repo list --output url 2>/dev/null)); then
		COMPREPLY+=( $( compgen -W "${cd_out[*]}" -- "$cur" ) )
	fi
}

__cd_list_projects() {
	local -a cd_out
	if cd_out=($(cd proj list --output name 2>/dev/null)); then
		COMPREPLY+=( $( compgen -W "${cd_out[*]}" -- "$cur" ) )
	fi
}

__cd_list_namespaces() {
	local -a cd_out
	if cd_out=($(kubectl get namespaces --no-headers 2>/dev/null | cut -f1 -d' ' 2>/dev/null)); then
		COMPREPLY+=( $( compgen -W "${cd_out[*]}" -- "$cur" ) )
	fi
}

__cd_proj_server_namespace() {
	local -a command
	for comp_word in "${COMP_WORDS[@]}"; do
		if [[ $comp_word =~ ^-.*$ ]]; then
			continue
		fi
		command+=($comp_word)
	done

	# expect something like this: cd proj add-destination PROJECT SERVER NAMESPACE
	local project=${command[3]}
	local server=${command[4]}
	local namespace=${command[5]}
	if [[ -z $project || $project == $cur ]]; then
		__cd_list_projects
	elif [[ -z $server || $server == $cur ]]; then
		__cd_list_servers
	elif [[ -z $namespace || $namespace == $cur ]]; then
		__cd_list_namespaces
	fi
}

__cd_list_project_role() {
	local project="$1"
	local -a cd_out
	if cd_out=($(cd proj role list "$project" --output=name 2>/dev/null)); then
		COMPREPLY+=( $( compgen -W "${cd_out[*]}" -- "$cur" ) )
	fi
}

__cd_proj_role(){
	local -a command
	for comp_word in "${COMP_WORDS[@]}"; do
		if [[ $comp_word =~ ^-.*$ ]]; then
			continue
		fi
		command+=($comp_word)
	done

	# expect something like this: cd proj role add-policy PROJECT ROLE-NAME
	local project=${command[4]}
	local role=${command[5]}
	if [[ -z $project || $project == $cur ]]; then
		__cd_list_projects
	elif [[ -z $role || $role == $cur ]]; then
		__cd_list_project_role $project
	fi
}

__cd_custom_func() {
	case ${last_command} in
		cd_app_delete | \
		cd_app_diff | \
		cd_app_edit | \
		cd_app_get | \
		cd_app_history | \
		cd_app_manifests | \
		cd_app_patch-resource | \
		cd_app_set | \
		cd_app_sync | \
		cd_app_terminate-op | \
		cd_app_unset | \
		cd_app_wait | \
		cd_app_create)
			__cd_list_apps
			return
			;;
		cd_app_rollback)
			__cd_app_rollback
			return
			;;
		cd_cluster_get | \
		cd_cluster_rm | \
		cd_cluster_set | \
		cd_login | \
		cd_cluster_add)
			__cd_list_servers
			return
			;;
		cd_repo_rm | \
		cd_repo_add)
			__cd_list_repos
			return
			;;
		cd_proj_add-destination | \
		cd_proj_remove-destination)
			__cd_proj_server_namespace
			return
			;;
		cd_proj_add-source | \
		cd_proj_remove-source | \
		cd_proj_allow-cluster-resource | \
		cd_proj_allow-namespace-resource | \
		cd_proj_deny-cluster-resource | \
		cd_proj_deny-namespace-resource | \
		cd_proj_delete | \
		cd_proj_edit | \
		cd_proj_get | \
		cd_proj_set | \
		cd_proj_role_list)
			__cd_list_projects
			return
			;;
		cd_proj_role_remove-policy | \
		cd_proj_role_add-policy | \
		cd_proj_role_create | \
		cd_proj_role_delete | \
		cd_proj_role_get | \
		cd_proj_role_create-token | \
		cd_proj_role_delete-token)
			__cd_proj_role
			return
			;;
		*)
			;;
	esac
}
	`
)

func NewCompletionCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "completion SHELL",
		Short: "Output shell completion code for the specified shell (bash, zsh or fish)",
		Long: `Write bash, zsh or fish shell completion code to standard output.

For bash, ensure you have bash completions installed and enabled.
To access completions in your current shell, run
$ source <(cd completion bash)
Alternatively, write it to a file and source in .bash_profile

For zsh, add the following to your ~/.zshrc file:
source <(cd completion zsh)
compdef _cd cd

Optionally, also add the following, in case you are getting errors involving compdef & compinit such as command not found: compdef:
autoload -Uz compinit
compinit 
`,
		Example: `# For bash
$ source <(cd completion bash)

# For zsh
$ cd completion zsh > _cd
$ source _cd

# For fish
$ cd completion fish > ~/.config/fish/completions/cd.fish
$ source ~/.config/fish/completions/cd.fish

# For powershell
$ mkdir -Force "$HOME\Documents\PowerShell" | Out-Null
$ cd completion powershell > $HOME\Documents\PowerShell\cd_completion.ps1

Add the following lines to your powershell profile

$ # Hanzo CD tab completion
if (Test-Path "$HOME\Documents\PowerShell\cd_completion.ps1") {
    . "$HOME\Documents\PowerShell\cd_completion.ps1"
}

Then reload your profile
$ . $PROFILE
`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			shell := args[0]
			rootCommand := NewCommand()
			rootCommand.BashCompletionFunction = bashCompletionFunc
			availableCompletions := map[string]func(out io.Writer, cmd *cobra.Command) error{
				"bash":       runCompletionBash,
				"zsh":        runCompletionZsh,
				"fish":       runCompletionFish,
				"powershell": runCompletionPowershell,
			}
			completion, ok := availableCompletions[shell]
			if !ok {
				fmt.Printf("Invalid shell '%s'. The supported shells are bash, zsh and fish.\n", shell)
				os.Exit(1)
			}
			if err := completion(os.Stdout, rootCommand); err != nil {
				log.Fatal(err)
			}
		},
	}

	return command
}

func runCompletionBash(out io.Writer, cmd *cobra.Command) error {
	return cmd.GenBashCompletion(out)
}

func runCompletionZsh(out io.Writer, cmd *cobra.Command) error {
	return cmd.GenZshCompletion(out)
}

func runCompletionFish(out io.Writer, cmd *cobra.Command) error {
	return cmd.GenFishCompletion(out, true)
}

func runCompletionPowershell(out io.Writer, cmd *cobra.Command) error {
	return cmd.GenPowerShellCompletionWithDesc(out)
}
