# gore

[![Go](https://github.com/kevwan/depu/workflows/Go/badge.svg?branch=main)](https://github.com/yolo-pkgs/gore/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/kevwan/depu)](https://goreportcard.com/report/github.com/yolo-pkgs/gore)
[![Release](https://img.shields.io/github/v/release/kevwan/depu.svg?style=flat-square)](https://github.com/yolo-pkgs/gore)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

"npm list/update -g" for Go

## Installation

```bash
go install github.com/yolo-pkgs/gore@latest
```

## How to use

```bash
# List all user binaries (and available updates) installed with 'go install'
gore list

# Update all user binaries
gore update

# Dump installation commands
gore dump
gore dump --latest
```
