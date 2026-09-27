# Security Policy

## Supported Versions

Security updates are actively applied to the following release branches:

| Version | Supported          | Notes                              |
| ------- | ------------------ | ---------------------------------- |
| 1.x.x   | :white_check_mark: | Current stable release line        |
| < 1.0.0 | :x:                | Legacy / preview releases (sunset) |

---

## Core Security Guarantees

### 1. Zip Slip Path Traversal Protection
Compressor strictly validates all entry headers in incoming zip archives prior to filesystem extraction.
- **Entry Path Sanitization**: Strips relative navigation tokens (`../`, `..\`) and rejects any entry whose canonical destination lies outside the target root directory.
- **Absolute Path Rejection**: Directly rejects archive entries containing absolute paths (e.g. `/etc/passwd`, `C:\Windows\System32`) or leading slashes/drive roots.
- **Null-Byte Injection Rejection**: Scans and discards paths with poisoned null bytes (`\x00`).

### 2. Symlink Boundary Enforcement
- Symlinks contained inside archives are evaluated relative to the extraction root.
- Any symlink pointing outside the designated extraction directory is blocked with an explicit security violation error (`ErrSymlinkEscape`).

### 3. Denial of Service (Zip Bomb) Mitigation
- Streaming decompression limits memory allocation during extraction.
- Compression ratio safeguards prevent memory exhaustion.

### 4. Atomic Operations
- Compression writes to temporary hidden files (`.compressor_tmp_*.zip`) before atomically moving the finished archive to its final destination, preventing partial or corrupted files from being created on interruption.

---

## Reporting a Vulnerability

If you discover a security vulnerability in Compressor, please **do not** open a public issue on GitHub. Instead, report it through the private disclosure channel:

- **Security Contact**: `security@compressor.local` or via GitHub Private Vulnerability Reporting under the **Security** tab.
- **Expected Information**:
  - Detailed description of the vulnerability and attack vector.
  - Minimal reproducible example or proof-of-concept archive.
  - Affected versions and operating system environments.

### Response Timelines
- **Initial Acknowledgment**: Within 24 hours.
- **Triage & Assessment**: Within 48 hours.
- **Patch & Advisory Release**: Coordinated with reporter, typically within 7–14 business days.
