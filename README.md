# Scripts

Scripts is a powerful Go-based CLI tool that provides a collection of useful scripts for enhancing your command-line workflow. It offers a range of functionalities including Git operations, Tmux session management, SSH utilities, and more.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Commands](#commands)
- [Contributing](#contributing)
- [License](#license)

## Features

- Git operations powered by LLM (semantic commits, staging files)
- Tmux session management (create, switch, list, sync)
- SSH utilities
- Text case conversion
- Command watching with interval-based execution

## Installation

### Pre-built Binary

Download the latest binaries from the [`Release`](https://github.com/cloudbridgeuy/scripts/releases) page.

### Building from Source

To build and install `scripts` from source, follow these steps:

1. Ensure you have Go installed on your system (version 1.16 or later recommended).
2. Clone the repository:
   ```
   git clone https://github.com/cloudbridgeuy/scripts.git
   cd scripts
   ```
3. Build the binary:
   ```
   go build -ldflags="-w -s" -o scripts
   ```
4. Move the binary to your local bin directory:
   ```
   mv scripts ~/.local/bin/scripts
   ```
5. Make the binary executable:
   ```
   chmod +x ~/.local/bin/scripts
   ```

Ensure that `~/.local/bin` is in your PATH.

## Usage

After installation, you can run `scripts` from anywhere in your terminal:

```
scripts [command] [subcommand] [flags]
```

For a list of available commands, run:

```
scripts --help
```

## Commands

Here's a brief overview of the main commands:

- `scripts git`: Git-related operations
  - `semantic`: Create semantic commits
- `scripts tmux`: Tmux session management
  - `new`: Create a new session
  - `display`: Show running sessions
  - `go`: Switch to a session
  - `sync`: Synchronize sessions
- `scripts ssh`: SSH utilities
- `scripts case`: Text case conversion
- `scripts watch`: Watch and execute commands at intervals

For detailed information on each command, use the `--help` flag:

```
scripts [command] --help
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
