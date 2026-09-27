<div align="center">

<img src="./docs/img/boa-banner.png" width="100%" alt="Boa Banner" />

<br/><br/>

[![Go Version](https://img.shields.io/badge/go-1.21%2B-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-00E599.svg)](./LICENSE)
[![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20Windows-00C7BE)](./README.md)
[![Security Audited](https://img.shields.io/badge/Security-ZipSlip%20Protected-brightgreen)](./SECURITY_AUDIT.md)

</div>

---

<div align="center">
  <img src="./docs/img/big-boa.png" width="760" alt="Boa - Compress Your Files" />
</div>

---

## ⚡ Interactive Main Menu

Launch the real-time interactive dashboard by typing `bo` or `boa`:

```text
 ____                 
| __ )  ___   __ _    
|  _ \ / _ \ / _` |   
| |_) | (_) | (_| |   https://github.com/zyadwael/boa
|____/ \___/ \__,_|   Tight, fast, lossless compression for your files.

Update 1.0.0 available, run bo update

   1. Pack        Compress folders into dense archives
   2. Unpack      Safely extract zip archives
   3. Inspect     Explore archive structure & metadata
   4. Benchmark   Compare compression speed & ratios
 ➤ 5. Status      Runtime health & system info

 ↑↓ / jk Navigate  |  Enter Confirm  |  1-5 Jump  |  V Version  |  Q Quit
```

---

## ✨ Features

- 🏎️ **Streaming I/O Engine**: Minimal RAM footprint with low-latency chunked streaming for large files.
- 🛡️ **Zip-Slip Defense**: Strict path boundary validation blocking directory traversals and escaping symlinks.
- 🎚️ **Granular Compression Tuning**: Levels `0` (Store), `1` (Fastest), `6` (Default), up to `9` (Maximum Deflate).
- 📊 **Dense Terminal UI**: Stable column alignment, single-screen summaries with human-friendly space savings.
- 🔍 **In-Place Inspection**: View archive tables with CRC32, compressed ratios, modification dates, and permissions.
- ⚡ **Built-In Benchmarking**: Compare compression throughput (`MB/s`) and ratios instantly with `bo bench`.
- 🪟 **True Cross-Platform**: Native builds and paths for macOS, Linux, and Windows.

---

## 🚀 Installation & Setup

### Quick Install (macOS / Linux)
```bash
make install
# Installs 'boa' and alias 'bo' to ~/.local/bin
```

### Build from Source
```bash
git clone https://github.com/zyadwael/boa.git
cd boa
make build
# Binary is ready at ./bin/bo
```

---

## 📖 Command Reference

```
COMMANDS
  bo                           Main menu
  bo pack                      Compress folders into zip archives
  bo unpack                    Safely extract zip archives
  bo list                      Inspect contents & compression ratios
  bo bench                     Benchmark compression levels (0-9)
  bo version                   Show version & platform info
  bo --help                    Show help

  bo pack ./folder -o dist.zip -l 9
  bo pack ./folder --dry-run
  bo unpack dist.zip -o ./out --force
  bo list dist.zip --json
  bo bench ./large-data

OPTIONS
  -v, --verbose                Show detailed operation logs
  -q, --quiet                  Suppress non-essential output
      --no-color               Disable ANSI color formatting
```

### 1. Compress / Pack Folder (`bo pack` / `bo zip`)
```bash
# Compress folder with default settings
bo pack ./my-project

# Specify output name and maximum compression (level 9)
bo pack ./my-project -o release.zip -l 9

# Exclude unwanted artifacts (e.g. node_modules, git, tmp)
bo pack ./my-project -e "node_modules,*.tmp,.git*"

# Dry-run simulation (preview file count and sizes without writing archive)
bo pack ./my-project --dry-run
```

### 2. Decompress / Unpack Archive (`bo unpack` / `bo unzip`)
```bash
# Extract into default folder (named after the zip archive)
bo unpack release.zip

# Extract to custom directory with overwrite enabled
bo unpack release.zip -o ./dist --force

# Preview extraction contents safely
bo unpack release.zip --dry-run
```

### 3. Inspect Archive Contents (`bo list` / `bo ls`)
```bash
# Tabular terminal inspection
bo list release.zip

# Output inspection as machine-readable JSON
bo list release.zip --json
```

### 4. Benchmark Compression Speeds (`bo bench`)
```bash
bo bench ./large-dataset
```

---

## 🛡️ Security Architecture

Boa enforces strict security rules at every stage:
- **Zip Slip Traversal**: Every archive header path is cleaned and validated against destination boundaries using `safety.ValidateDestinationPath`. Entries containing relative escapes (`../`) or absolute paths (`/`, `C:\`) are immediately aborted.
- **Symlink Boundary Checks**: Symlink destinations are resolved and verified not to escape the target directory.
- **Atomic File Generation**: Compression writes to temporary files first (`.compressor_tmp_*.zip`) and performs atomic renames upon successful stream finalization.

See [SECURITY.md](./SECURITY.md) and [SECURITY_AUDIT.md](./SECURITY_AUDIT.md) for detailed audit matrices.

---

## 📄 License

Distributed under the MIT License. See [LICENSE](./LICENSE) for details.
