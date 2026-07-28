# Plugins

## Overview

This guide demonstrates how to write plugins for the
`cd` CLI tool. Plugins are a way to extend `cd` CLI with new sub-commands,
allowing for custom features which are not part of the default distribution
of the `cd` CLI.

If you would like to take a look at the original proposal, head over to this [enhancement proposal](../proposals/cd-cli-pluin.md).
It covers how the plugin mechanism works, its benefits, motivations, and the goals it aims to achieve.

## Prerequisites

You need to have a working `cd` binary installed locally. You can follow
the [cli installation documentation](https://argo-cd.readthedocs.io/en/stable/cli_installation/) to install the binary.

## Create `cd` plugins

A plugin is a standalone executable file whose name begins with cd-.
To install a plugin, move its executable file to any directory included in your PATH.
Ensure that the PATH configuration specifies the full absolute path to the executable,
not a relative path. `cd` allows plugins to add custom commands such as
`cd my-plugin arg1 arg2 --flag1` by executing a `cd-my-plugin` binary in the PATH.

## Limitations

1. It is currently not possible to create plugins that overwrite existing
`cd` commands. For example, creating a plugin such as `cd-version`
will cause the plugin to never get executed, as the existing `cd version`
command will always take precedence over it. Due to this limitation, it is
also not possible to use plugins to add new subcommands to existing `cd` commands.
For example, adding a subcommand `cd cluster upgrade` by naming your plugin
`cd-cluster` will cause the plugin to be ignored.

2. It is currently not possible to parse the global flags set by `cd` CLI. For example, 
if you have set any global flag value such as `--logformat` value to `text`, the plugin will
not parse the global flags and pass the default value to the `--logformat` flag which is `json`.
The flag parsing will work exactly the same way for existing `cd` commands which means executing a
existing cd command such as `cd cluster list` will correctly parse the flag value as `text`.

## Conditions for an `cd` plugin

Any binary that you would want to execute as an `cd` plugin need to satisfy the following three conditions:

1. The binary should start with `cd-` as the prefix name. For example,
   `cd-demo-plugin` or `cd-demo_plugin` is a valid binary name but not
   `cd_demo-plugin` or `cd_demo_plugin`.
2. The binary should have executable permissions otherwise it will be ignored.
3. The binary should reside anywhere in the system's absolute PATH.

## Writing `cd` plugins

### Naming a plugin

An Hanzo CD plugin’s filename must start with `cd-`. The subcommands implemented
by the plugin are determined by the portion of the filename after the `cd-` prefix.
Anything after `cd-` will become a subcommand for `cd`.

For example, A plugin named `cd-demo-plugin` is invoked when the user types:
```bash
cd demo-plugin [args] [flags]
```

The `cd` CLI determines which plugin to invoke based on the subcommands provided.

For example, executing the following command:
```bash
cd my-custom-command [args] [flags]
```
will lead to the execution of plugin named `cd-my-custom-command` if it is present in the PATH.

### Writing a plugin

A plugin can be written in any programming language or script that allows you to write command-line commands.

A plugin determines which command path it wishes to implement based on its name.

For example, If a binary named `cd-demo-plugin` is available in your system's absolute PATH, and the user runs the following command:

```bash
cd demo-plugin subcommand1 --flag=true
```

Hanzo CD will translate and execute the corresponding plugin with the following command:

```bash
cd-demo-plugin subcommand1 --flag=true
```

Similarly, if a plugin named `cd-demo-demo-plugin` is found in the absolute PATH, and the user invokes:

```bash
cd demo-demo-plugin subcommand2 subcommand3 --flag=true
```

Hanzo CD will execute the plugin as:

```bash
cd-demo-demo-plugin subcommand2 subcommand3 --flag=true
```

### Example plugin
```bash
#!/bin/bash

# Check if the cd CLI is installed
if ! command -v cd &> /dev/null; then
    echo "Error: Hanzo CD CLI (cd) is not installed. Please install it first."
    exit 1
fi

if [[ "$1" == "version" ]]
then
    echo "displaying cd version..."
    cd version
    exit 0
fi


echo "I am a plugin named cd-foo"
```

### Using a plugin

To use a plugin, make the plugin executable:
```bash
sudo chmod +x ./cd-foo
```

and place it anywhere in your `PATH`:
```bash
sudo mv ./cd-foo /usr/local/bin
```

You may now invoke your plugin as a cd command:
```bash
cd foo
```

This would give the following output
```bash
I am a plugin named cd-foo
```

All args and flags are passed as-is to the executable:
```bash
cd foo version
```

This would give the following output
```bash
DEBU[0000] command does not exist, looking for a plugin... 
displaying cd version...
2025/01/16 13:24:36 maxprocs: Leaving GOMAXPROCS=16: CPU quota undefined
cd: v2.13.0-rc2+0f083c9
  BuildDate: 2024-09-20T11:59:25Z
  GitCommit: 0f083c9e58638fc292cf064e294a1aa53caa5630
  GitTreeState: clean
  GoVersion: go1.22.7
  Compiler: gc
  Platform: linux/amd64
cd-server: v2.13.0-rc2+0f083c9
  BuildDate: 2024-09-20T11:59:25Z
  GitCommit: 0f083c9e58638fc292cf064e294a1aa53caa5630
  GitTreeState: clean
  GoVersion: go1.22.7
  Compiler: gc
  Platform: linux/amd64
  Kustomize Version: v5.4.3 2024-07-19T16:40:33Z
  Helm Version: v3.15.2+g1a500d5
  Kubectl Version: v0.31.0
  Jsonnet Version: v0.20.0
```

## Distributing `cd` plugins

If you’ve developed an Hanzo CD plugin for others to use,
you should carefully consider how to package, distribute, and
deliver updates to ensure a smooth installation and upgrade process
for your users.

### Native / platform specific package management

You can distribute your plugin using traditional package managers,
such as `apt` or `yum` for Linux, `Chocolatey` for Windows, and `Homebrew` for macOS.
These package managers are well-suited for distributing plugins as they can
place executables directly into the user's PATH, making them easily accessible.

However, as a plugin author, choosing this approach comes with the responsibility of
maintaining and updating the plugin's distribution packages across multiple platforms
for every release. This includes testing for compatibility, ensuring timely updates,
and managing versioning to provide a seamless experience for your users.

### Source code

You can publish the source code of your plugin, for example,
in a Git repository. This allows users to access and inspect
the code directly. Users who want to install the plugin will need
to fetch the code, set up a suitable build environment (if the plugin requires compiling),
and manually deploy it.
