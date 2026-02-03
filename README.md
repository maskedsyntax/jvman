# jvman

Cross-platform Java Version Manager. Install and switch between multiple JDK versions on Linux, macOS, and Windows.

## Installation

### Download Binary (Recommended)

Download the latest release for your platform from [GitHub Releases](https://github.com/maskedsyntax/jvman/releases).

Linux/macOS:
```bash
# Download and extract (replace VERSION and PLATFORM)
curl -LO https://github.com/maskedsyntax/jvman/releases/download/VERSION/jvman_VERSION_PLATFORM.tar.gz
tar -xzf jvman_VERSION_PLATFORM.tar.gz
sudo mv jvman /usr/local/bin/
```

Windows: Download the `.zip` file and add `jvman.exe` to your PATH.

### Using Go

If you have Go installed:
```bash
go install github.com/maskedsyntax/jvman/cmd/jvman@latest
```

**Note:** This installs the binary to `$HOME/go/bin`. Make sure this directory is in your PATH:
```bash
export PATH="$HOME/go/bin:$PATH"
```

### Build from Source

```bash
git clone https://github.com/maskedsyntax/jvman.git
cd jvman
go build -o jvman ./cmd/jvman
```

### Setup

After installation, initialize jvman:

```bash
jvman init
```

Add the bin directory to your PATH (the init command shows the exact path):

```bash
export PATH="$HOME/.jvman/bin:$PATH"
```

Add this line to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.) to make it permanent.

## Usage

### Install a JDK

```bash
jvman install 21                      # Install Temurin 21 (default vendor)
jvman install 17 --vendor=corretto    # Install Amazon Corretto 17
jvman install 11 -v zulu              # Install Azul Zulu 11
jvman install 21 --arch=aarch64       # Install for specific architecture
```

Supported vendors: `temurin` (default), `corretto`, `zulu`

Supported architectures: `x64`, `aarch64`

### List versions

```bash
jvman list                   # Show installed and available versions from all vendors
jvman list --vendor=temurin  # Filter to a specific vendor
jvman list --refresh         # Bypass cache and fetch fresh data
```

Version lists are cached for 1 hour. Use `--refresh` or `jvman cache clear` to get fresh data.

### Switch versions

Set the global default:

```bash
jvman global 21           # Use version 21 (auto-detects vendor)
jvman global temurin-21   # Use specific vendor-version
```

Set a project-local version (creates a `.jvman` file):

```bash
jvman use 17
```

### Run with a specific version

Use `exec` to run a command with a specific Java version without changing your global or local settings:

```bash
jvman exec 21 java -version
jvman exec corretto-17 javac Main.java
jvman exec zulu-11 mvn clean install
```

This sets `JAVA_HOME` and prepends the JDK's bin directory to `PATH` for the executed command.

### Interactive mode

Launch the terminal UI for browsing and managing installed versions:

```bash
jvman tui
```

In the TUI you can:
- Browse installed versions
- Press Enter to set a version as global
- Press d to remove a version
- Use / to filter the list
- Press q to quit

### Other commands

```bash
jvman which        # Show the currently active Java and how it was resolved
jvman remove 21    # Uninstall a version
jvman cache clear  # Clear the version cache
jvman upgrade      # Check for jvman updates
jvman version      # Show jvman version
```

## Version Resolution

jvman resolves the active Java version in this order:

1. `.jvman` file in the current directory or any parent directory
2. Local override set in config for the current directory
3. Global default

## Shims

After installation, `~/.jvman/bin` contains shims for common JDK tools:

- java, javac, jar, jshell, javadoc, jarsigner, keytool, jlink, jpackage

These shims automatically use the resolved Java version based on your current directory.

## Shell Completion

Generate shell completion scripts:

```bash
jvman completion bash > /etc/bash_completion.d/jvman   # Bash
jvman completion zsh > "${fpath[1]}/_jvman"            # Zsh
jvman completion fish > ~/.config/fish/completions/jvman.fish  # Fish
```

Run `jvman completion <shell> --help` for detailed instructions.
