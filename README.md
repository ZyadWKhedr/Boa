# ⚡ Compressor

> A dense, high-performance, and secure zip compression CLI engine for terminal power users. Inspired by the refined terminal ergonomics of **[tw93/Mole](https://github.com/tw93/Mole)**.

[![Go Version](https://img.shields.io/badge/go-1.21%2B-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20Windows-blue)](./README.md)
[![Security Audited](https://img.shields.io/badge/Security-ZipSlip%20Protected-brightgreen)](./SECURITY_AUDIT.md)

---

## ✨ Features

- 🏎️ **Streaming I/O Engine**: Low memory footprint during both compression and extraction with direct chunked streams.
- 🛡️ **Zip-Slip Defense**: Built-in canonical path boundary validation that prevents path traversal and malicious symlink attacks.
- 🎚️ **Granular Compression Tuning**: Supports compression levels `0` (Store), `1` (Fastest), `6` (Default), up to `9` (Maximum Deflate).
- 📊 **Dense Terminal UI**: Stable column alignment, single-screen metric summaries, ANSI color palette with `NO_COLOR` support.
- 🔍 **In-Place Inspection**: View archive tables with CRC32, compressed ratios, modification dates, and permissions without extracting.
- ⚡ **Built-In Benchmarking**: Compare compression levels and throughput speeds (`MB/s`) instantly with `compressor bench`.
- 🪟 **True Cross-Platform**: First-class support for macOS (Darwin), Linux, and Windows.

---

## 🚀 Quick Start

### Installation

#### Using Go
```bash
go install compressor@latest
```

#### Build from Source
```bash
git clone https://github.com/your-org/compressor.git
cd compressor
make build
# Binary is ready at ./bin/compressor
```

---

## 📖 Command Reference

```
USAGE
  compressor [command] [flags]

COMMANDS
  pack         Compress folders or files into a secure zip archive (aliases: zip, compress, c, p)
  unpack       Safely decompress a zip archive into a directory (aliases: unzip, extract, x, u)
  list         List contents, file sizes, and metadata of a zip archive (aliases: ls, inspect, l)
  bench        Benchmark compression levels (0, 1, 6, 9) and compare speed vs ratio
  version      Print the version and runtime environment information

GLOBAL FLAGS
  -v, --verbose    Enable verbose per-file terminal logging
  -q, --quiet      Suppress non-essential progress output
      --no-color   Disable ANSI color output (also respects NO_COLOR env)
```

### 1. Compress / Pack Folder
```bash
# Compress folder with default settings
compressor pack ./my-project

# Specify output name and maximum compression (level 9)
compressor pack ./my-project -o release.zip -l 9

# Exclude unwanted artifacts (e.g. node_modules, git, tmp)
compressor pack ./my-project -e "node_modules,*.tmp,.git*"

# Dry-run simulation (preview file count and sizes without writing archive)
compressor pack ./my-project --dry-run
```

### 2. Decompress / Unpack Archive
```bash
# Extract into default folder (named after the zip archive)
compressor unpack release.zip

# Extract to custom directory with overwrite enabled
compressor unpack release.zip -o ./dist --force

# Preview extraction contents safely
compressor unpack release.zip --dry-run
```

### 3. Inspect Archive Contents
```bash
# Tabular terminal inspection
compressor list release.zip

# Output inspection as machine-readable JSON
compressor list release.zip --json
```

### 4. Benchmark Compression Speeds
```bash
compressor bench ./large-dataset
```

---

## 🛡️ Security Architecture

Compressor enforces strict security rules at every stage:
- **Zip Slip Traversal**: Every archive header path is cleaned and validated against destination boundaries using `safety.ValidateDestinationPath`. Entries containing relative escapes (`../`) or absolute paths (`/`, `C:\`) are immediately aborted.
- **Symlink Boundary Checks**: Symlink destinations are resolved and verified not to escape the target directory.
- **Atomic File Generation**: Compression writes to temporary files first (`.compressor_tmp_*.zip`) and performs atomic renames upon successful stream finalization.

See [SECURITY.md](./SECURITY.md) and [SECURITY_AUDIT.md](./SECURITY_AUDIT.md) for detailed audit matrices.

---

## 🏗️ Repository Architecture

```
.
├── cmd/               # Cobra CLI command definitions (pack, unpack, list, bench, root, version)
├── internal/
│   ├── compress/      # Core streaming zip creation engine & compression levels
│   ├── extract/       # Safe zip decompression & inspection engine
│   ├── safety/        # Path traversal, Zip-Slip, and symlink validation
│   ├── stats/         # Byte formatting, speed, and ratio calculation
│   └── ui/            # Dense terminal output, tables, and color themes
├── pkg/
│   └── types/         # Public data structures and execution options
├── main.go            # Binary entry point
├── RULES.md           # Product directions & operational guidelines for AI agents
├── SECURITY.md        # Vulnerability disclosure and security guarantees
├── SECURITY_AUDIT.md  # Threat matrix evaluation report
├── TRADEMARK.md       # Trademark and brand usage guidelines
└── Makefile           # Build, test, and cross-compilation automation
```

---

## 📄 License

Distributed under the MIT License. See [LICENSE](./LICENSE) for details.
