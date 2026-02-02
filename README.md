# jvman

Cross-platform Java Version Manager. Install and switch between multiple JDK versions on Linux, macOS, and Windows.

## Installation

Build from source:

```bash
go build -o jvman ./cmd/jvman
```

Initialize jvman and set up shims:

```bash
./jvman init
```

Add the bin directory to your PATH (the init command will show the exact path):

```bash
export PATH="$HOME/.jvman/bin:$PATH"
```

## Usage

### Install a JDK

```bash
jvman install 21                      # Install Temurin 21 (default vendor)
jvman install 17 --vendor=corretto    # Install Amazon Corretto 17
jvman install 11 -v zulu              # Install Azul Zulu 11
```

Supported vendors: `temurin` (default), `corretto`, `zulu`

### List versions

```bash
jvman list                   # Show installed and available versions from all vendors
jvman list --vendor=temurin  # Filter to a specific vendor
```

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

### Other commands

```bash
jvman which      # Show the currently active Java and how it was resolved
jvman remove 21  # Uninstall a version
jvman version    # Show jvman version
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
