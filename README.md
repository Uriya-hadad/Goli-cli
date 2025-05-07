# GOLI-CLI


`goli` is a command-line interface (CLI) tool designed for managing and interacting with applications and instances on the CF landscape.

This README provides an overview of the tool, its commands.


## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
    - [Interactive Mode](#interactive-mode)
    - [Batch mode](#Batch-ModeComing-Soon)
- [Command-Line Completion](#Command-Line-CompletionComing-Soon)

## Installation

To install `goli`, follow these steps:

1. **Download the CLI binary** for your operating system from the [releases page](https://github.wdf.sap.corp/Portal-CF/Goli-Cli/releases).
2. **Make the binary executable**:
   ```sh
   chmod +x /path/to/goli
    ```
3. **Move the binary to a directory in your PATH (or add the directory to your PATH)**:
    ```sh
    sudo mv /path/to/goli /usr/local/bin/goli
    ```
4. **Verify the installation**:
    ```sh
    goli --version
    ```

## Usage
Once installed, you can use goli to manage and interact with your applications and instances. \
Below are examples of the primary commands.

## Interactive Mode
The interactive mode provides a guided experience for managing your applications and instances. \
To start the interactive mode, run the following command:
```sh
goli
```
Try it out and see how it works.

## Batch Mode
The batch mode allows you to run commands without the interactive mode. \
First select the command you want to run(applications / instances), then provide the required flags and arguments (app / instance name). \
For example:
```sh
goli applications --raw
```
```sh
goli instances <instanceName>
```

## Command-Line Completion
***only for macOS** \
***zsh auto-completion is required for this** \
To enable command-line completion for `goli`, follow these steps:
1. download the zsh auto-completion script and place it in the appropriate directory.
2. copy the following command and paste it in your terminal:
```sh
mv /path/to/script <I-user>/.zsh/zsh-completions/src/_goli
```



