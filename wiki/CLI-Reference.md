# CLI Reference

Boa offers a versatile command-line interface with streaming I/O, scriptable JSON/CSV exports, and security validation.

---

## 1. `bo compress` (Packaging)

Compress files or directories into a `.zip` archive.

```bash
bo compress <source-folder-or-file> [flags]
```

### Aliases:
`pack`, `zip`, `c`, `p`

### Flags:
| Flag | Short | Default | Description |
|---|---|---|---|
| `--output` | `-o` | `<source>.zip` | Explicit output zip path |
| `--method` | `-m` | `deflate` | Compression algorithm (`deflate`, `store`, `zstd`) |
| `--level` | `-l` | `-1` | Compression level (`0`=Store, `1`–`9`=DEFLATE, `1`–`11`=ZSTD) |
| `--exclude` | `-e` | `.git*,.DS_Store,node_modules` | Comma-separated exclusion glob patterns |
| `--force` | `-f` | `false` | Overwrite destination archive if it exists |
| `--dry-run` | | `false` | Simulate compression without writing to disk |
| `--lossy` | | `""` | Comma-separated lossy categories (`images,audio,video`) |
| `--quality` | | `80` | Quality target for lossy compression (1–100) |
| `--strip-metadata`| | `false` | Strip EXIF, GPS, and creation tags during media encoding |

### Examples:
```bash
# Basic compression
bo compress ./project

# Maximum DEFLATE compression with custom output path
bo compress ./data -o backup.zip -l 9 -f

# High-speed modern Zstandard compression
bo compress ./logs -m zstd -l 3

# Lossy image compression at 75% quality
bo compress ./photos --lossy images --quality 75 --strip-metadata
```

---

## 2. `bo extract` (Decompression)

Extract archives safely with Zip-Slip path sanitization.

```bash
bo extract <archive.zip> [flags]
```

### Aliases:
`unpack`, `unzip`, `x`, `u`, `d`

### Flags:
| Flag | Short | Default | Description |
|---|---|---|---|
| `--output` | `-o` | `<parent>/<archive_name>` | Extraction destination directory |
| `--force` | `-f` | `false` | Overwrite destination files if present |

### Examples:
```bash
bo extract archive.zip
bo extract archive.zip -o ./destination -f
```

---

## 3. `bo list` (Archive Inspection)

Inspect archive entries, CRC32 checksums, timestamps, and lossy metadata without extracting.

```bash
bo list <archive.zip> [flags]
```

### Aliases:
`ls`, `l`, `inspect`, `info`

### Flags:
| Flag | Description |
|---|---|
| `--all` | Display entire file list without truncation |
| `--json` | Output structured JSON metadata |

---

## 4. `bo bench` (Benchmark Engine)

Evaluate compression speed, duration, space savings, and throughput across levels (0 to 9).

```bash
bo bench <target-path> [flags]
```

### Aliases:
`benchmark`, `b`, `test`

### Flags:
| Flag | Default | Description |
|---|---|---|
| `--level` | `-1` | Benchmark a specific level only (0–9) |
| `--runs` | `1` | Number of iterations for statistical averaging |
| `--compare` | `false` | Visual side-by-side bar chart comparison |
| `--json` | `false` | Export results as structured JSON |
| `--csv` | `false` | Export results as CSV |
| `--keep-archives` | `false` | Preserve generated test zip files on disk |

---

## 5. `bo learn` (Knowledge Base)

Explore interactive explanations of data compression mechanics:

```bash
bo learn [technique-id]
```

Examples:
```bash
bo learn                # View all available techniques
bo learn deflate        # Deep-dive on LZ77 and Huffman Coding
bo learn jpeg-quality   # Learn DCT frequency transform & quantization
```

---

## 6. `bo update` (Auto-Updater)

Check for and automatically install the newest pre-built Boa release binary from GitHub.

```bash
bo update [flags]
```

### Aliases:
`upgrade`, `self-update`

### Flags:
| Flag | Short | Default | Description |
|---|---|---|---|
| `--check` | `-c` | `false` | Check for updates without downloading or installing |
| `--yes` | `-y` | `false` | Automatically accept prompts and install update |

---

## 7. `bo uninstall` (Safe System Removal)

Cleanly remove Boa binaries, symlinks, and aliases (`boa`, `bo`, `compressor`) from `~/.local/bin` and `/usr/local/bin`.

```bash
bo uninstall [flags]
```

### Flags:
| Flag | Short | Default | Description |
|---|---|---|---|
| `--force` | `-f` | `false` | Bypass confirmation prompt and uninstall immediately |

---

## 8. Global Flags

| Flag | Short | Description |
|---|---|---|
| `--verbose` | `-v` | Enable detailed step-by-step logging |
| `--quiet` | `-q` | Suppress non-essential output |
| `--no-color` | | Disable ANSI terminal colors |
| `--help` | `-h` | Show help reference |
