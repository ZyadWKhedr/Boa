# Changelog

All notable changes to **Boa** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [v0.1.0] - 2026-09-27

### Added
- Core streaming compression engine with Deflate levels 0 through 9 (`bo pack`).
- Safe decompression engine with strict Zip-Slip protection (`bo unpack`).
- In-place archive inspection table with CRC32, modification times, permissions, and compression ratios (`bo list`).
- In-terminal real-time File Explorer with arrow-key navigation (`↑↓` / `jk`).
- Multi-level speed and ratio benchmark command (`bo bench`).
- 2-Column high-contrast scannable terminal metric summaries.
- One-line install script (`install.sh`) for macOS and Linux.
- Cross-platform release automation for Darwin, Linux, and Windows.
- Formal security audit documentation (`SECURITY_AUDIT.md`) and threat model.
