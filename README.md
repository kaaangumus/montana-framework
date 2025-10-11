# Montana Framework

---

*"A lightweight and modular exploit framework for researchers and developers."*

---

Montana is an interactive framework designed to manage and utilize a local collection of exploits, inspired by the structure of the `0day.today` archive. It allows for rapid searching, analysis, and deployment of security proofs-of-concept.

## Features

- **Interactive Shell:** An easy-to-use command-line interface with contextual prompts.
- **Exploit Database:** Automatically loads and indexes exploits from a user-provided `index.json` file.
- **Advanced Search:** Quickly find exploits by keyword, searching across titles, platforms, authors, and CVEs.
- **Nmap Integration:** Launch Nmap scans directly from the framework.
- **Smart Suggestions:** Get automatic exploit recommendations based on the results of your Nmap scans.
- **Easy Export:** Save any exploit's content to a file for external use with the `save` command.

## Project Structure

The project is organized as follows:

```
.
├── exploits/               # Directory for exploit files
├── go.mod
├── go.sum
├── index.json              # Exploit index file
├── install.sh              # Installation script
├── main.go                 # Main application source code
└── README.md
```

## Installation

To install the Montana Framework on your system and make it accessible from anywhere in your terminal, follow these steps:

1.  **Prerequisites:**
    - **Go:** Version 1.18 or higher.
    - **Nmap:** Must be installed and in your system's PATH.

2.  **Clone the repository (if you haven't already):**
    ```bash
    git clone https://github.com/your-username/montana-framework.git
    cd montana-framework
    ```

3.  **Run the installer:**
    The `install.sh` script will build the binary, and copy it to `/usr/local/bin`. It will also copy the `exploits` directory and `index.json` to `/usr/local/share/montana-framework`.

    **Important:** You need to run the script with `sudo` because it writes to system directories.

    ```bash
    chmod +x install.sh
    sudo ./install.sh
    ```

4.  **Run the application:**
    Once the installation is complete, you can run the framework from anywhere in your terminal:
    ```bash
    montana-framework
    ```

## Usage

### Main Commands

- `search <keyword(s)>`: Search for exploits. 
  - *Example:* `search wordpress 5.2`
- `list`: Interactively lists exploits. It first prompts for a category and then displays the filtered results in a scrollable view.
  - *Example:* Type `list`, then enter a category like `web-applications` or `all`.
- `use <ID>`: Select an exploit to interact with. The prompt will change to show the active exploit.
  - *Example:* `use 33814`
- `nmap <nmap_args>`: Run an Nmap scan. For best results with `suggest`, use service version detection (`-sV`).
  - *Example:* `nmap -sV 127.0.0.1`
- `suggest`: Get exploit suggestions based on the last Nmap scan.
- `help`: Display the help message.
- `exit`: Close the framework.

### Exploit Context Commands

After selecting an exploit with `use <ID>`, the following commands become available:

- `info`: Show detailed information about the active exploit.
- `content`: Display the full source code/content of the exploit.
- `save <filename>`: Save the exploit's content to a new file.
  - *Example:* `save my_exploit.py`
- `back`: Exit from the current exploit context and return to the main menu.

## Development

If you want to run the application from the source code without installing it, you can use the following command:

```bash
go run main.go
```