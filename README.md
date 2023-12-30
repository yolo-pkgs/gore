# gore

![Go](https://github.com/yolo-pkgs/gore/actions/workflows/go.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/yolo-pkgs/gore)](https://goreportcard.com/report/github.com/yolo-pkgs/gore)
[![Release](https://img.shields.io/github/v/release/yolo-pkgs/gore.svg?style=flat-square)](https://github.com/yolo-pkgs/gore)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

"npm list/update -g" for Go

List and update user binaries installed through "go install". Works with public, private and development (v0.0.0-...) packages.

![list](https://github.com/yolo-pkgs/gore/assets/71007493/e121225a-af70-4d7a-9d87-e4470608a36a)

![update](https://github.com/yolo-pkgs/gore/assets/71007493/67125d67-2491-4cc0-b650-6fb211cadafb)

## Installation

Requires Go 1.20+

```bash
go install github.com/yolo-pkgs/gore@latest
```

## How to use

```bash
# List all user binaries (and available updates) installed with 'go install'
gore list

# Pass --dev (or -d) to also check dev packages like v0.0.0-...
gore list -d

# Pass --group (or -g) to group packages by domain
gore list -g

# Pass --extra (or -e) to print extra info
gore list -e

# Pass --simple (or -s) to print without table
gore list -s

# Update all user binaries
gore update

# Dump installation commands
gore dump
gore dump --latest
```

![group](https://github.com/yolo-pkgs/gore/assets/71007493/a006adea-a23b-40f8-9009-76cc7e03205f)
