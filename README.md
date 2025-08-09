# YankRun

<div align="center">
  <img src="images/yankrun.jpg" alt="YankRun" width="200">
  <p>
    <img src="https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat-square&logo=go" alt="Go Version">
    <img src="https://img.shields.io/badge/OS-Linux%20%7C%20macOS%20%7C%20Windows-darkblue?style=flat-square&logo=windows" alt="OS Support">
    <img src="https://img.shields.io/badge/License-MIT-green?style=flat-square" alt="License">
  </p>
</div>

Template smarter: clone repos and replace tokens safely with size limits, custom delimiters, and JSON/YAML inputs.

## Install

### From Release
<details>
<summary><strong>Linux/macOS (AMD64)</strong></summary>

```sh
curl -L https://github.com/AxeByte/yankrun.axebyte/releases/download/stable/yankrun-linux-amd64.tar.gz -o yankrun-linux-amd64.tar.gz
tar -xvf yankrun-linux-amd64.tar.gz yankrun-linux-amd64
chmod +x yankrun-linux-amd64
sudo mv yankrun-linux-amd64 /usr/local/bin/yankrun
```

</details>

<details>
<summary><strong>Linux/macOS (ARM64)</strong></summary>

```sh
curl -L https://github.com/AxeByte/yankrun.axebyte/releases/download/stable/yankrun-linux-arm64.tar.gz -o yankrun-linux-arm64.tar.gz
tar -xvf yankrun-linux-arm64.tar.gz yankrun-linux-arm64
chmod +x yankrun-linux-arm64
sudo mv yankrun-linux-arm64 /usr/local/bin/yankrun
```

</details>

<details>
<summary><strong>Windows (PowerShell)</strong></summary>

```powershell
Invoke-WebRequest -Uri https://github.com/AxeByte/yankrun.axebyte/releases/download/stable/yankrun-windows-amd64.zip -OutFile yankrun-windows-amd64.zip
Expand-Archive -Path yankrun-windows-amd64.zip -DestinationPath .
Move-Item -Path yankrun-windows-amd64/yankrun-windows-amd64.exe -Destination yankrun.exe
```

</details>

### From Source
<details>
<summary><strong>Build locally</strong></summary>

```sh
git clone https://github.com/AxeByte/yankrun.axebyte.git
cd yankrun.axebyte
go build -o yankrun .
sudo mv yankrun /usr/local/bin/
```

Or install with Go:

```sh
go install github.com/AxeByte/yankrun.axebyte@latest
```

</details>

## Usage

<details>
<summary><strong>Clone & replace</strong></summary>

```sh
yankrun clone \
  -r https://github.com/user/repo.git \
  -i example.json \
  -od ./clonedRepo \
  -v
```

Options:
- `-r, --repo`: Git URL to clone
- `-i, --input`: JSON/YAML with variables and ignore patterns
- `-od, --outputDir`: directory to clone into
- `-fl, --fileSizeLimit`: skip files larger than this (default `3 mb`)
- `-sd, --startDelim`: template start delimiter (default `[[{[`)
- `-ed, --endDelim`: template end delimiter (default `]}]]`)

</details>

<details>
<summary><strong>Template command (interactive)</strong></summary>

```sh
# Analyze placeholders and prompt for values
yankrun template -d ./target-dir -p

# Use defaults or overrides
yankrun template -d ./target-dir -i example.yaml -sd "[[{" -ed "}]]" -fl "5 mb" -p -v
```

What it does:
- Scans `-d` for placeholders between your delimiters (defaults: `[[{[`, `]}]]`).
- Shows a summary of each placeholder with how many matches were found.
- Pre-fills values from `-i` if provided; prompts for missing ones.
- Applies replacements across the directory and prints a completion message.

</details>

## Configuration

<details>
<summary><strong>Interactive setup</strong></summary>

```sh
# Create or update ~/.yankrun/config.yaml
yankrun setup

# Example session
Template start delimiter [[{[]: [[{[
Template end delimiter ]}]][: ]}]]
File size limit (e.g. 3 mb) [3 mb]: 3 mb
```

Flags always override config defaults if provided.

</details>

<details>
<summary><strong>Show current config</strong></summary>

```sh
yankrun setup --show
```

Outputs:

```text
start_delim: [[{[
end_delim: ]}]]
file_size_limit: 3 mb
```

</details>

## Input file format

<details>
<summary><strong>JSON</strong></summary>

```json
{
  "ignore_patterns": ["node_modules", "dist"],
  "variables": [
    { "key": "Company", "value": "Your Company" },
    { "key": "Team", "value": "Your Team" }
  ]
}
```

Notes:
- If your keys do not include delimiters, YankRun wraps them using your configured delimiters (default `[[{[` and `]}]]`). For example, `Company` becomes `[[{[Company]}]]`.
- If your keys already include delimiters, they are used as-is.

</details>

<details>
<summary><strong>YAML</strong></summary>

```yaml
ignore_patterns: [node_modules, dist]
variables:
  - key: Company
    value: Your Company
  - key: Team
    value: Your Team
```

</details>

## Examples

<details>
<summary><strong>Set custom delimiters per run</strong></summary>

```sh
yankrun clone -r <repo> -i example.yaml -od out -sd "[[{" -ed "}]]"
```

</details>

<details>
<summary><strong>Skip large files</strong></summary>

```sh
yankrun clone -r <repo> -i example.json -od out -fl "10 mb"
```

</details>

<details>
<summary><strong>Verbose replacement report</strong></summary>

```sh
yankrun clone -r <repo> -i example.json -od out -v
```

</details>

## Features

- Template values replacement across a directory tree
- Git clone with post-clone templating
- Custom delimiters with smart wrapping
- Size-based skipping (default 3 MB)
- Verbose reporting
- JSON/YAML inputs and ignore patterns
