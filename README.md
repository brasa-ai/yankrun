# YankRun

<div align="center">
  <img src="doc/logo.png" alt="YankRun" width="200">
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
curl -L https://github.com/brasa-ai/yankrun/releases/download/stable/yankrun-linux-amd64.tar.gz -o yankrun-linux-amd64.tar.gz
tar -xvf yankrun-linux-amd64.tar.gz yankrun-linux-amd64
chmod +x yankrun-linux-amd64
sudo mv yankrun-linux-amd64 /usr/local/bin/yankrun
```

</details>

<details>
<summary><strong>GitHub discovery config</strong></summary>

You can auto-discover template repos from your GitHub user and/or orgs. All fields are optional; using only orgs is fine.

```yaml
# ~/.yankrun/config.yaml (excerpt)
github:
  orgs: ["brasa-ai", "your-org"]          # one or more orgs (optional)
  user: "your-user"                         # your GitHub user (optional)
  topic: "templates"                        # filter repos by topic (optional)
  prefix: "template-"                       # filter repos by name prefix (optional)
  include_private: true                      # include private repos (requires token)
  token: "GITHUB_TOKEN"                      # optional; for higher rate limits/private
```

Notes:
- If nothing is configured yet, `yankrun generate` will ask for user/orgs inline and save them.
- When both `user` and `orgs` are set, results are merged.

</details>

<details>
<summary><strong>Reset configuration</strong></summary>

```sh
yankrun setup --reset
```

Deletes `~/.yankrun/config.yaml`.

</details>

<details>
<summary><strong>Linux/macOS (ARM64)</strong></summary>

```sh
curl -L https://github.com/brasa-ai/yankrun/releases/download/stable/yankrun-linux-arm64.tar.gz -o yankrun-linux-arm64.tar.gz
tar -xvf yankrun-linux-arm64.tar.gz yankrun-linux-arm64
chmod +x yankrun-linux-arm64
sudo mv yankrun-linux-arm64 /usr/local/bin/yankrun
```

</details>

<details>
<summary><strong>Windows (PowerShell)</strong></summary>

```powershell
Invoke-WebRequest -Uri https://github.com/brasa-ai/yankrun/releases/download/stable/yankrun-windows-amd64.zip -OutFile yankrun-windows-amd64.zip
Expand-Archive -Path yankrun-windows-amd64.zip -DestinationPath .
Move-Item -Path yankrun-windows-amd64/yankrun-windows-amd64.exe -Destination yankrun.exe
```

</details>

### From Source
<details>
<summary><strong>Build locally</strong></summary>

```sh
git clone https://github.com/brasa-ai/yankrun.git
cd yankrun
go build -o yankrun .
sudo mv yankrun /usr/local/bin/
```

Or install with Go:

```sh
go install github.com/brasa-ai/yankrun@latest
```

</details>

## Usage

<details>
<summary><strong>Clone & replace (interactive and non-interactive)</strong></summary>

```sh
# Non-interactive: provide values via --input
yankrun clone \
  --repo https://github.com/brasa-ai/template-tester.git \
  --input examples/values.json \
  --outputDir ./clonedRepo \
  --verbose

# Interactive: prompt for discovered placeholders after clone
yankrun clone \
  --repo git@github.com:brasa-ai/template-tester.git \
  --outputDir ./clonedRepo \
  --prompt --verbose
```

What it does:
- Clones the repository
- Scans for placeholders between your delimiters (defaults: `[[`, `]]`)
- If `-p/--prompt` is set, shows a summary and prompts for values; otherwise uses values from `-i` if provided
- Applies replacements and logs completion

Options:
- `--repo`: Git URL to clone
- `--input`: JSON/YAML with variables (used in non-interactive or as defaults in interactive)
- `--outputDir`: directory to clone into
- `--fileSizeLimit`: skip files larger than this (default `3 mb`)
- `--startDelim`: template start delimiter (default `[[`)
- `--endDelim`: template end delimiter (default `]]`)
- `--prompt` (alias: `--interactive`): ask for values before applying

</details>

<details>
<summary><strong>Generate (choose template repo & branch)</strong></summary>

```sh
# Configure templates in ~/.yankrun/config.yaml
# templates:
#   - name: "Go App"
#     url: "git@github.com:brasa-ai/template-tester.git"
#     description: "Example templates"
#     default_branch: "main"

# Run interactive generator
yankrun generate --prompt --verbose

# Non-interactive values file and custom delimiters
yankrun generate --input examples/values.json --startDelim "[[{" --endDelim "}]]" --fileSizeLimit "5 mb"
```

What it does:
- Loads configured templates from `~/.yankrun/config.yaml`
- Lets you choose a template and branch
- Clones the selected branch
- Removes `.git` so you start a fresh repo
- Scans placeholders, optionally prompts (`-p`), then applies replacements

</details>

<details>
<summary><strong>Template command (interactive)</strong></summary>

```sh
# Analyze placeholders and prompt for values
yankrun template --dir ./examples/project --prompt

# Use defaults or overrides (YAML values)
yankrun template --dir ./examples/project --input examples/values.yaml --startDelim "[[{" --endDelim "}]]" --fileSizeLimit "5 mb" --prompt --verbose
```

What it does:
- Scans `--dir` for placeholders between your delimiters (defaults: `[[`, `]]`).
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
Template start delimiter [[]: [[
Template end delimiter ]]: ]]
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
start_delim: [[
end_delim: ]]
file_size_limit: 3 mb
```

</details>

## Input file format

<details>
<summary><strong>JSON</strong> (see `examples/values.json` in the tester repo)</summary>

```json
{
  "ignore_patterns": ["node_modules", "dist"],
  "variables": [
    { "key": "APP_NAME", "value": "TemplateTester" },
    { "key": "PROJECT_NAME", "value": "DemoProject" },
    { "key": "USER_NAME", "value": "axebyte" },
    { "key": "USER_EMAIL", "value": "user@example.com" },
    { "key": "VERSION", "value": "1.0.0" },
    { "key": "APP_NAME_APPLY_UPPERCASE", "value": "my-app" },
    { "key": "PROJECT_NAME_APPLY_DOWNCASE", "value": "MyProject" },
    { "key": "SLUG_APPLY_REPLACE", "value": "my-project-name" }
  ],
  "functions": {
    "APPLY_REPLACE": {
      "-": "_",
      " ": "-"
    }
  }
}
```

Notes:
- If your keys do not include delimiters, YankRun wraps them using your configured delimiters. For example, with start `[[` and end `]]`, `APP_NAME` becomes `my-app`.
- If your keys already include delimiters, they are used as-is.
- **Smart Transformations**: Template names containing `APPLY_UPPERCASE` or `APPLY_DOWNCASE` automatically transform values to upper or lowercase.
- **Custom Replace Functions**: Use `APPLY_REPLACE` in the functions section to define custom string replacements (e.g., `-` to `_`, spaces to `-`).

</details>

<details>
<summary><strong>YAML</strong> (see `examples/values.yaml` in the tester repo)</summary>

```yaml
ignore_patterns: [node_modules, dist]
variables:
  - key: APP_NAME
    value: TemplateTester
  - key: PROJECT_NAME
    value: DemoProject
  - key: USER_NAME
    value: axebyte
  - key: USER_EMAIL
    value: user@example.com
  - key: VERSION
    value: "1.0.0"
  - key: APP_NAME_APPLY_UPPERCASE
    value: my-app
  - key: PROJECT_NAME_APPLY_DOWNCASE
    value: MyProject
  - key: SLUG_APPLY_REPLACE
    value: my-project-name
functions:
  APPLY_REPLACE:
    "-": "_"
    " ": "-"
```

</details>

## Smart Template Features

**How Transformations Work**: Transformations are only applied to template names that contain the corresponding keywords. This means:
- `APP_NAME_APPLY_UPPERCASE` gets uppercase transformation
- `APP_NAME_APPLY_DOWNCASE` gets lowercase transformation  
- `SLUG_APPLY_REPLACE` gets custom replacement functions
- `APP_NAME_APPLY_UPPERCASE_APPLY_REPLACE` gets BOTH uppercase AND replacement functions
- `SLUG` (without keywords) gets NO transformations

**Example**:
```yaml
variables:
  - key: "APP_NAME"                           # No transformations
    value: "my-app"
  - key: "APP_NAME_APPLY_UPPERCASE"           # Uppercase only
    value: "my-app"
  - key: "APP_NAME_APPLY_REPLACE"             # Replace only
    value: "my-app"
  - key: "APP_NAME_APPLY_UPPERCASE_APPLY_REPLACE"  # Both transformations
    value: "my-project-name"
functions:
  APPLY_REPLACE:
    "-": "_"
```

Results:
- `APP_NAME` → `my-app` (no change)
- `APP_NAME_APPLY_UPPERCASE` → `MY-APP` (uppercase)
- `APP_NAME_APPLY_REPLACE` → `my-app` (no hyphens to replace)
- `APP_NAME_APPLY_UPPERCASE_APPLY_REPLACE` → `MY_PROJECT_NAME` (uppercase + replace)

<details>
<summary><strong>Automatic Case Transformations</strong></summary>

YankRun automatically applies transformations based on template names:

```yaml
variables:
  - key: "APP_NAME_APPLY_UPPERCASE"
    value: "my-app"
  - key: "PROJECT_NAME_APPLY_DOWNCASE" 
    value: "MyProject"
```

Results:
- `APP_NAME_APPLY_UPPERCASE` → `MY-APP`
- `PROJECT_NAME_APPLY_DOWNCASE` → `myproject`

</details>

<details>
<summary><strong>Custom String Replacements</strong></summary>

Define custom replacement rules in the `functions` section. These only apply to templates with "APPLY_REPLACE" in their name:

```yaml
variables:
  - key: "SLUG_APPLY_REPLACE"
    value: "my-project-name"
functions:
  APPLY_REPLACE:
    "-": "_"      # Replace hyphens with underscores
    " ": "-"      # Replace spaces with hyphens
```

Result: `my-project-name` → `my_project_name`

**Important**: The `SLUG` template (without APPLY_REPLACE) would remain unchanged even if functions are defined.

</details>

<details>
<summary><strong>Combined Transformations</strong></summary>

You can combine multiple transformations by including multiple keywords in the template name:

```yaml
variables:
  - key: "APP_NAME_APPLY_UPPERCASE_APPLY_REPLACE"
    value: "my-project-name"
functions:
  APPLY_REPLACE:
    "-": "_"
```

Result: `my-project-name` → `MY_PROJECT_NAME` (uppercase + replace)

**Important**: APPLY_REPLACE functions only apply to templates with "APPLY_REPLACE" in their name, not globally to all values.

</details>

## Examples

<details>
<summary><strong>Set custom delimiters per run</strong></summary>

```sh
yankrun clone --repo git@github.com:brasa-ai/template-tester.git --input examples/values.yaml --outputDir out --startDelim "[[{" --endDelim "}]]"
```

</details>

<details>
<summary><strong>Skip large files</strong></summary>

```sh
yankrun clone --repo git@github.com:brasa-ai/template-tester.git --input examples/values.json --outputDir out --fileSizeLimit "10 mb"
```

</details>

<details>
<summary><strong>Verbose replacement report</strong></summary>

```sh
yankrun clone --repo <repo> --input example.json --outputDir out --verbose
```

## Why YankRun? Practical problems it solves

<details>
<summary><strong>1) Bootstrap a new project from a template</strong></summary>

Problem: You maintain a template repo (CI, lint, base code). You want to create a new project with your org/app names filled in, without carrying over the template’s git history.

Solution:

```sh
# Choose template + branch, clone, remove .git, scan tokens, fill values
yankrun generate --prompt --verbose
```

Outcome: Fresh repo with placeholders (e.g., [[NAME]], [[PROJECT_NAME]]) replaced and no template history.

</details>

<details>
<summary><strong>2) Rollout org-wide config changes across many files</strong></summary>

Problem: You have dozens of files with tokens for company, team, emails, or versions. Manual search/replace is error-prone.

Solution:

```sh
# Define values once
cat > values.json << 'EOF'
{
  "variables": [
    { "key": "COMPANY", "value": "Acme Corp" },
    { "key": "TEAM", "value": "Platform" },
    { "key": "VERSION", "value": "2.1.0" }
  ]
}
EOF

# Apply everywhere safely with size limits
yankrun template --dir . --input values.json --fileSizeLimit "5 mb" --verbose
```

Outcome: Consistent updates with a per-file replacement report, skipping large/binary files.

</details>

<details>
<summary><strong>3) Customize a sample app quickly (no prompts)</strong></summary>

Problem: You want a non-interactive pipeline (CI/CD) to stamp out a project with predetermined values.

Solution:

```sh
yankrun clone \
  --repo git@github.com:brasa-ai/template-tester.git \
  --input examples/values.json \
  --outputDir ./my-app \
  --startDelim "[[{" --endDelim "}]]" \
  --verbose
```

Outcome: Fully templated project ready for commit in automated flows.

</details>

<details>
<summary><strong>4) Smart transformations for consistent naming</strong></summary>

Problem: You need consistent naming conventions across files (uppercase for constants, lowercase for variables, slugified for URLs).

Solution:

```yaml
variables:
  - key: "APP_NAME_APPLY_UPPERCASE"
    value: "my-app"
  - key: "PROJECT_NAME_APPLY_DOWNCASE"
    value: "MyProject"
  - key: "SLUG_APPLY_REPLACE"
    value: "my project name"
functions:
  APPLY_REPLACE:
    " ": "-"
```

Outcome: Automatic generation of `MY-APP`, `myproject`, and `my-project-name` from single input values.

**Note**: Each transformation only applies to templates containing the corresponding keyword in their name.

</details>


</details>

## Features

- Template values replacement across a directory tree
- Git clone with post-clone templating
- Custom delimiters with smart wrapping
- Size-based skipping (default 3 MB)
- Verbose reporting
- JSON/YAML inputs and ignore patterns
- **Smart Template Transformations**: Automatic case conversion and custom string replacements (only applied to templates with matching keywords)
