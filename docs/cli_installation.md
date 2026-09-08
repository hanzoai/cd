# Installation

You can download the latest Hanzo CD version from [the latest release page of this repository](https://github.com/hanzoai/cd/releases/latest), which will include the `cd` CLI.

## Linux and WSL

### ArchLinux

```bash
pacman -S cd
```

### Homebrew

```bash
brew install cd
```

### Download With Curl

#### Download latest version

```bash
curl -sSL -o cd-linux-amd64 https://github.com/hanzoai/cd/releases/latest/download/cd-linux-amd64
sudo install -m 555 cd-linux-amd64 /usr/local/bin/cd
rm cd-linux-amd64
```

#### Download concrete version

Set `VERSION` replacing `<TAG>` in the command below with the version of Hanzo CD you would like to download:

```bash
VERSION=<TAG> # Select desired TAG from https://github.com/hanzoai/cd/releases
curl -sSL -o cd-linux-amd64 https://github.com/hanzoai/cd/releases/download/$VERSION/cd-linux-amd64
sudo install -m 555 cd-linux-amd64 /usr/local/bin/cd
rm cd-linux-amd64
```

#### Download latest stable version

You can download the latest stable release by executing below steps:

```bash
VERSION=$(curl -L -s https://raw.githubusercontent.com/hanzoai/cd/main/VERSION)
curl -sSL -o cd-linux-amd64 https://github.com/hanzoai/cd/releases/download/v$VERSION/cd-linux-amd64
sudo install -m 555 cd-linux-amd64 /usr/local/bin/cd
rm cd-linux-amd64
```

You should now be able to run `cd` commands.


## Mac

### Install via Homebrew or Curl

You can install the CLI using `Homebrew` or a `Curl` command:

#### Homebrew

Both Intel and Apple Silicon Macs can use Homebrew:

```bash
brew install cd
```

#### Download With Curl

Choose the appropriate binary for your Mac's architecture:

##### For Intel Macs (x86_64)

You can view the latest version of Hanzo CD at the link above or run the following command to grab the version:

```bash
VERSION=$(curl --silent "https://api.github.com/repos/hanzoai/cd/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
```

Replace `VERSION` in the command below with the version of Hanzo CD you would like to download:

```bash
curl -sSL -o cd https://github.com/hanzoai/cd/releases/download/$VERSION/cd-darwin-amd64
```

Install the Hanzo CD CLI binary:

```bash
sudo install -m 555 cd /usr/local/bin/cd
rm cd
```

##### For Apple Silicon Macs (M1/M2/M3)

You can view the latest version of Hanzo CD at the link above or run the following command to grab the version:

```bash
VERSION=$(curl --silent "https://api.github.com/repos/hanzoai/cd/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
```

Replace `VERSION` in the command below with the version of Hanzo CD you would like to download:

```bash
curl -sSL -o cd https://github.com/hanzoai/cd/releases/download/$VERSION/cd-darwin-arm64
```

Install the Hanzo CD CLI binary:

```bash
sudo install -m 555 cd /usr/local/bin/cd
rm cd
```

After finishing either of the instructions above, you should now be able to run `cd` commands.


## Windows

### Download With PowerShell: Invoke-WebRequest

You can view the latest version of Hanzo CD at the link above or run the following command to grab the version:

```powershell
$version = (Invoke-RestMethod https://api.github.com/repos/hanzoai/cd/releases/latest).tag_name
```

Replace `$version` in the command below with the version of Hanzo CD you would like to download:

```powershell
$url = "https://github.com/hanzoai/cd/releases/download/" + $version + "/cd-windows-amd64.exe"
$output = "cd.exe"

Invoke-WebRequest -Uri $url -OutFile $output
```
Also please note you will probably need to move the file into your PATH.
Use following command to add Hanzo CD into environment variables PATH

```powershell
[Environment]::SetEnvironmentVariable("Path", "$env:Path;C:\Path\To\Hanzo CD-CLI", "User")
```


After finishing the instructions above, you should now be able to run `cd` commands.
