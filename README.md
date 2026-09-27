<div align="center">

<img src="./docs/img/boa-banner.png" width="100%" alt="Boa Banner" />

<br/><br/>

# Boa

**Fast, secure, interactive compression for your terminal.**

[![Go Version](https://img.shields.io/badge/go-1.21%2B-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-00E599.svg)](./LICENSE)
[![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20Windows-00C7BE)](./README.md)
[![Security Audited](https://img.shields.io/badge/Security-ZipSlip%20Protected-brightgreen)](./SECURITY_AUDIT.md)

</div>

<br/>

```text
┌──────────────────────────────────────────────────────────────┐
│                            BOA                               │
│                                                              │
│   1. Pack        Compress folders into dense archives        │
│   2. Unpack      Safely extract zip archives                 │
│   3. Inspect     Explore archive structure & metadata        │
│   4. Benchmark   Compare compression speed & ratios          │
│   5. Status      Runtime health & system info                │
│                                                              │
│   ↑↓ Navigate    Enter Confirm    1-5 Jump    Q Quit         │
└──────────────────────────────────────────────────────────────┘
```

<br/>

## Install

### macOS / Linux (One-Line Script)

```bash
curl -fsSL https://raw.githubusercontent.com/ZyadWKhedr/Boa/main/install.sh | bash
```

### Go Install

```bash
go install github.com/ZyadWKhedr/Boa@latest
```

### Build from Source

```bash
git clone https://github.com/ZyadWKhedr/Boa.git
cd Boa
make install
```

---

## Quick Examples

```bash
bo                      # Launch interactive menu & file explorer
bo pack ./my-folder     # Compress folder into my-folder.zip
bo unpack project.zip   # Extract archive safely
bo list project.zip     # View file sizes, CRC32, and compression ratios
bo bench ./my-folder    # Benchmark throughput speeds (MB/s) across levels (0-9)
```

---

## Why Boa?

Most command-line compression workflows force you to memorize obtuse flags across different utilities (`zip`, `unzip`, `tar`, `7z`) or switch between GUI file managers and terminal windows.

Boa delivers a **modern terminal-native archive experience**:
- **Interactive File Explorer**: Navigate your filesystem, select folders, and compress without typing long paths manually.
- **Streaming Low-RAM Engine**: Compresses large multi-gigabyte directories in direct chunked streams.
- **Secure by Default**: Automatically blocks Zip-Slip directory traversal attacks (`../../`) and malicious symlink escapes before writing to disk.
- **Instant Benchmarking**: Compare compression speeds and ratios across Deflate levels (0 to 9) side-by-side.

<br/>

<div align="center">
  <img src="./docs/img/big-boa.png" width="760" alt="Boa - Compress Your Files" />
</div>

<br/>

---

## Features

- **All-in-One CLI Toolkit**: Combines fast zip compression, safe extraction, in-place header inspection, live speed benchmarking, and an interactive terminal file explorer in a single zero-dependency binary.
- **Streaming I/O Engine**: Low memory footprint during both compression and extraction with direct chunked streams.
- **Zip-Slip & Path Security**: Built-in canonical path boundary validation that blocks directory traversals (`../../`), null-byte injections, and escaping symlinks.
- **Granular Compression Tuning**: Supports compression levels `0` (Store), `1` (Fastest speed), `6` (Default), up to `9` (Maximum Deflate).
- **Interactive Terminal File Explorer**: Browse folders and zip files directly from your terminal using arrow keys (`↑↓` / `jk`).
- **Dense Terminal UI**: Stable column alignment, single-screen summaries, and ANSI color palette with automatic `NO_COLOR` support.
- **Cross-Platform Parity**: Native builds for macOS (Apple Silicon & Intel), Linux (x86_64 & ARM), and Windows.

---

## Command Reference

```
COMMANDS
  bo                           Main menu & file explorer
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

### Options and Flags

```bash
bo pack ./my-folder --dry-run          # Preview files to pack and estimated size
bo unpack archive.zip --dry-run        # Preview extraction paths safely
bo pack ./my-folder -e "node_modules"  # Exclude patterns (*.tmp, .git*, node_modules)
bo pack ./my-folder -l 9               # Use maximum compression level
bo unpack archive.zip -o ./dist -f     # Overwrite destination files with --force
bo list archive.zip --json             # Export structured JSON metadata
```

---

## Safety Architecture

Boa validates paths, enforces extraction boundaries, and asks for confirmation when overwriting existing files.

- **Zip-Slip Defense**: All entry paths in incoming archives are sanitized with canonical prefix checks. Any entry with `../`, leading slashes (`/`), or drive roots (`C:\`) is blocked with an explicit security error before disk access.
- **Symlink Boundary Checks**: Symlinks pointing outside the extraction boundary are prevented from executing.
- **Atomic Operations**: Compression writes to temporary files first (`.compressor_tmp_*.zip`) before atomically moving the finished archive into place.
- **Same-Directory Defaults**: Compressed files are automatically saved directly in the same parent directory as the target source.
- Review [SECURITY.md](./SECURITY.md) and [SECURITY_AUDIT.md](./SECURITY_AUDIT.md) for full threat matrix evaluations and verification benchmarks.

---

## Features in Detail

### 1. Interactive Main Menu & File Picker
Typing `bo` in your terminal launches the real-time interactive dashboard with arrow-key (`↑↓` / `jk`) navigation:

```text
 ____                 
| __ )  ___   __ _    
|  _ \ / _ \ / _` |   
| |_) | (_) | (_| |   https://github.com/ZyadWKhedr/Boa
|____/ \___/ \__,_|   Tight, fast, lossless compression for your files.

Update 1.0.0 available, run bo update

   1. Pack        Compress folders into dense archives
   2. Unpack      Safely extract zip archives
   3. Inspect     Explore archive structure & metadata
   4. Benchmark   Compare compression speed & ratios
 ➤ 5. Status      Runtime health & system info

 ↑↓ / jk Navigate  |  Enter Confirm  |  1-5 Jump  |  V Version  |  Q Quit
```

Selecting **Pack** or **Unpack** opens the in-terminal **File Explorer**:

```text
▶ Select Folder or File to Compress
Location: /Users/zyadwael/Go/compressor

➤ ✔  [SELECT THIS CURRENT FOLDER: compressor]
  📁  .. (Parent Directory)
  📁  cmd/
  📁  docs/
  📁  internal/
  📁  pkg/
  📄  Makefile
  📄  README.md
  📄  main.go

========================================================================
 ↑↓ Navigate  |  Enter Open/Select  |  S Select Current Folder  |  Q Cancel
```

---

### 2. Pack (Compression)
`bo pack` streams and compresses files and directories with customizable levels:

```text
$ bo pack ./cmd -l 6

 ✔ Compression Complete

   Archive          ~/Desktop/cmd.zip
   Original Size    13.7 KB
   Compressed       7.4 KB
   Space Saved      6.3 KB (46.1% reduction)
   Ratio            1.85x
   Speed            14.2 MB/s
   Packed Items     7 files (avg 1.9 KB), 1 folders
   Duration         2ms
```

---

### 3. Unpack (Decompression)
`bo unpack` extracts files safely with path verification:

```text
$ bo unpack ./cmd.zip -o ./extracted --force

 ✔ Extraction Complete

   Source Archive   ~/Desktop/cmd.zip
   Extracted To     ~/Desktop/extracted
   Total Size       13.7 KB
   Items            7 files (avg 1.9 KB), 1 folders
   Speed            12.3 MB/s
   Duration         1ms
```

---

### 4. Inspect (List Archive)
`bo list` displays a dense tabular breakdown of archive entries without extracting:

```text
$ bo list ./cmd.zip

Type  Permissions  Original  Packed  Ratio  Modified Date     Path           
────  ───────────  ────────  ──────  ─────  ────────────────  ───────────────
DIR   drwxr-xr-x          -       -      -  2026-09-27 13:11  cmd/           
FILE  -rw-r--r--     2.4 KB  1.1 KB    55%  2026-09-27 13:08  cmd/bench.go   
FILE  -rw-r--r--     1.6 KB   618 B    63%  2026-09-27 13:11  cmd/cmd_test.go
FILE  -rw-r--r--     1.1 KB   647 B    43%  2026-09-27 13:08  cmd/list.go    
FILE  -rw-r--r--     2.8 KB  1.3 KB    55%  2026-09-27 13:08  cmd/pack.go    
FILE  -rw-r--r--     2.2 KB  1.1 KB    50%  2026-09-27 13:07  cmd/root.go    
FILE  -rw-r--r--     2.6 KB  1.2 KB    54%  2026-09-27 13:08  cmd/unpack.go  
FILE  -rw-r--r--      946 B   479 B    49%  2026-09-27 13:08  cmd/version.go 

   Archive          ~/Desktop/cmd.zip
   Total Entries    7 files (avg 1.9 KB), 1 folders
   Original Size    13.7 KB
   Archive Size     7.4 KB
   Total Savings    1.85x (46.1% saved)
```

---

### 5. Benchmark (Speed vs Ratio)
`bo bench` evaluates compression levels 0, 1, 6, and 9 side-by-side with throughput metrics:

```text
$ bo bench ./cmd

▶ Benchmarking Compression Levels on: ./cmd

Level        Original Size  Packed Size  Saved   Time      Speed
───────────  ─────────────  ───────────  ─────  ─────  ─────────
0 (Store)          13.7 KB      14.7 KB  -7.6%  661µs  20.2 MB/s
1 (Fastest)        13.7 KB       7.9 KB  42.2%    1ms  10.3 MB/s
6 (Default)        13.7 KB       7.4 KB  46.1%    1ms  10.6 MB/s
9 (Best)           13.7 KB       7.3 KB  46.5%    1ms  10.2 MB/s
```

---

### 6. JSON Export
`bo list --json archive.zip` returns structured metadata for scripting:

```json
{
  "total_files": 7,
  "total_dirs": 1,
  "uncompressed_bytes": 14032,
  "compressed_bytes": 7578,
  "average_file_size": 2004,
  "compression_ratio": 0.54,
  "space_saved_bytes": 6454,
  "space_saved_percent": 46.1,
  "archive_path": "cmd.zip"
}
```

---

## Support & Community

If Boa helped you, give it a star on GitHub, [share it on X](https://twitter.com/intent/tweet?url=https://github.com/ZyadWKhedr/Boa&text=Boa%20-%20Tight%2C%20fast%2C%20lossless%20compression%20for%20your%20files.), or open an issue or pull request.

- **Author**: Zyad Wael
- **X (Twitter)**: [@Zyadwaelll05](https://x.com/Zyadwaelll05)
- **LinkedIn**: [Zyad Wael](https://www.linkedin.com/in/zyad-wael-a9035a275/)
- **Email**: [ziad.w.khedr@gmail.com](mailto:ziad.w.khedr@gmail.com)

---

## License

Boa is free open source software released under the **MIT License**. See [LICENSE](./LICENSE) for details.
