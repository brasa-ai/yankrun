# YankRun

<div align="center">
  <img src="images/yankrun.jpg" alt="YankRun">
</div>

YankRun is a powerful CLI tool designed to facilitate template value replacement and repository cloning with template file replacements. Whether you're setting up new environments or managing configuration files, YankRun simplifies the process.

## Features

- **Template Values**: Easily replace template values in files.
- **Clone Repositories**: Clone Git repositories and apply template replacements.
- **Configurable File Size Limit**: Skip files that exceed a specified size during the replacement process.
- **Verbose Mode**: Get detailed logs of the operations being performed.
- **Support for YAML and JSON**: Read and parse replacements from both YAML and JSON files.
- **Flexible Ignoring Patterns**: Specify patterns for files and directories to ignore during the replacement process.

## Installation

To install YankRun, ensure you have Go installed and run the following command:

```sh
go get -u example.com/myapp
```

## Usage

YankRun provides two main commands: `template` and `clone`. Below are the details and examples for each command.

<details>
<summary><strong>Template Command</strong></summary>

The `template` command is used to replace template values in a specified file.

#### Usage

```sh
yankrun template -i <input-file>
```

#### Example

```sh
yankrun template -i example.json
```

</details>

<details>
<summary><strong>Clone Command</strong></summary>

The `clone` command clones a Git repository and replaces template values in the files.

#### Usage

```sh
yankrun clone -r <repo-url> -i <input-file> -od <output-dir> -fl <file-size-limit> -v
```

#### Options

- `-r, --repo`: URL of the repository to clone.
- `-i, --input`: Input file containing template values for replacement.
- `-od, --outputDir`: Output directory to clone the repository.
- `-fl, --fileSizeLimit`: File size limit to ignore replacements in files exceeding the limit (e.g., "5 mb").
- `-v, --verbose`: Enable verbose mode for detailed output.

#### Example

```sh
yankrun clone -r https://github.com/user/repo.git -i example.json -od ./clonedRepo -fl "5 mb" -v
```

</details>

<details>
<summary><strong>Input File Format</strong></summary>

The input file should be a JSON or YAML file with the following structure:

#### JSON

```json
{
    "ignore_patterns": [],
    "variables": [
        {
            "key": "<!Company!>",
            "value": "Your Company"
        },
        {
            "key": "<!Team!>",
            "value": "Your Team"
        }
    ]
}
```

#### YAML

```yaml
ignore_patterns: []
variables:
  - key: "<!Company!>"
    value: "Your Company"
  - key: "<!Team!>"
    value: "Your Team"
```

</details>

<details>
<summary><strong>Verbose Mode</strong></summary>

Enabling verbose mode provides detailed logs of the operations, including the number of replacements made in each file.

```sh
yankrun clone -r https://github.com/user/repo.git -i example.json -od ./clonedRepo -fl "5 mb" -v
```

</details>

<details>
<summary><strong>Setting File Size Limit</strong></summary>

You can specify a file size limit to skip files that exceed the limit during the replacement process.

```sh
yankrun clone -r https://github.com/user/repo.git -i example.json -od ./clonedRepo -fl "10 mb"
```

</details>

## Error Handling

YankRun will panic and stop execution if an error occurs during file reading, parsing, or writing operations. Ensure your input files and directories are correctly specified to avoid interruptions.

## Contributing

We welcome contributions! Please fork the repository and submit pull requests.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.