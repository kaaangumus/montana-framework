# Montana

Montana is a local exploit search tool inspired by `searchsploit`.

It searches `index.json` and prints matching exploit records. The repository also keeps a local `exploits/` archive for offline reference, but the CLI does not execute, copy, export, or open exploit code.

## Why This Changed

The interactive framework behavior was removed. Montana no longer opens shells, runs Nmap, suggests exploit execution paths, or copies exploit files into the working directory.

The current CLI searches metadata:

- ID
- Title
- CVE
- Platform
- Category
- Author
- Date
- Original source link

The `exploits/` directory is kept as an offline archive. Because that directory contains proof-of-concept exploit text, antivirus products such as Windows Defender may still flag or quarantine files from the archive.

## Features

- Fast local search over `index.json`
- Keyword search
- CVE search
- Platform and category filters
- Detail view by exploit ID
- No shell mode
- No exploit export or copy feature

## Installation

Requirements:

- Go 1.20+

Build locally:

```bash
go build -o montana .
```

Run from the repository:

```bash
./montana -q "wordpress rce"
```

Install on Linux:

```bash
chmod +x install.sh
sudo ./install.sh
```

The installer copies the binary to `/usr/local/bin/montana` and `index.json` to `/usr/local/share/montana/index.json`.

## Usage

Search by keyword:

```bash
montana -q "apache rce"
```

You can also pass the query directly:

```bash
montana wordpress 5.2
```

Search by CVE:

```bash
montana -cve CVE-2021-41773
```

Filter by platform and category:

```bash
montana -platform linux -category remote -q openssh
```

Show a single record:

```bash
montana -id 33814
```

Use a custom index file:

```bash
montana -index /path/to/index.json -q nginx
```

Or set:

```bash
export MONTANA_INDEX=/path/to/index.json
```

## Data Model

`index.json` is expected to contain an array of records:

```json
[
  {
    "exploit_id": 33814,
    "date": "01/01/2020",
    "category": "remote",
    "platform": "linux",
    "author": "researcher",
    "cve": ["CVE-0000-0000"],
    "title": "Example product remote issue",
    "original_link": "https://example.com/source"
  }
]
```

## Safety Notes

Montana is a search and indexing utility. It does not execute exploit source code and does not execute external security tools.

Use the metadata only for authorized security research, lab work, patch verification, and defensive triage.

## License

See [LICENSE](LICENSE).
