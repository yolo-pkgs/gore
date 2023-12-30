# gore

[![Go](https://github.com/yolo-pkgs/gore/workflows/go/badge.svg?branch=main)](https://github.com/yolo-pkgs/gore/actions/workflows/go.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/yolo-pkgs/gore)](https://goreportcard.com/report/github.com/yolo-pkgs/gore)
[![Release](https://img.shields.io/github/v/release/yolo-pkgs/gore.svg?style=flat-square)](https://github.com/yolo-pkgs/gore)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

"npm list/update -g" for Go

List and update user binaries installed through "go install".

## Installation

Requires Go 1.20+

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
