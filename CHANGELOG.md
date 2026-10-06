# Changelog

All notable changes to **Boa** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [v0.3.2] - 2026-10-06

### Added
- **Dynamic Real-Time Progress Bar**: Responsive in-terminal progress bar with throughput rates (MB/s), file count ratios, and percentage tracking during compression and extraction.

### Changed
- **Minimalist Aesthetic Cleanup**: Removed superfluous decorative emojis from CLI and TUI output while preserving the iconic 🐍 Boa snake and 📁 directory indicators.
- **Badge Standardization**: Updated status badges (`[OK]`, `[INFO]`, `[WARN]`, `[ERROR]`, `[DRY-RUN]`) for clean terminal consistency.

---

## [v0.3.1] - 2026-10-06

### Added
- **ASCII Art & GitHub Link on Startup**: Stylized BOA ASCII header with direct GitHub repository link (`https://github.com/ZyadWKhedr/Boa`) displayed when launching the interactive Bubble Tea dashboard.
- **Primary CLI Command Renaming**: Standardized primary commands on `bo compress` and `bo extract` (retaining `pack`, `unpack`, `zip`, `unzip`, `c`, `x` as backwards-compatible aliases).

### Changed
- **Cleaner CLI Output**: Removed redundant banner printing from direct `bo compress` and `bo extract` execution runs for cleaner piping and automation output.
- **Documentation**: Updated all README and CLI documentation references to `v0.3.1` and `compress`/`extract`.

---

## [v0.3.0] - 2026-10-04

### Added
- **Clean Architecture Refactor**: Separated codebase into `domain`, `usecase`, `infrastructure`, and `presentation` layers.
- **Multi-Media Compression Engine**:
  - Image optimization: Pure Go JPEG quality scaling, PNG 256-color palette quantization, and lossless PNG compression.
  - Audio transcoding: Configurable bitrates (64k–192k) via AAC/Opus/MP3.
  - Video transcoding: Perceptual CRF rate control and downscaling (1080p, 720p, 480p).
  - Metadata stripping: Optional removal of EXIF, GPS, and creation tags during transcoding.
- **Strict Content-First Classifier**: Magic byte inspection to identify true media formats and defeat deceptive file extensions.
- **Fail-Closed Safety Policy**: Dual-policy enforcement guaranteeing documents, code, text, databases, executables, and unknown binaries remain strictly 100% lossless.
- **Educational Knowledge Base (`boa learn`)**: Embedded technique catalog with interactive CLI explanations (`boa learn <id>`) and contextual `?` modals.
- **Bubble Tea Terminal Interface**: Rebuilt interactive dashboard, directory file picker, and multi-step compression wizard with Charm Bubble Tea and Lip Gloss.
- **CLI Flags**: Added `--lossy`, `--quality`, and `--strip-metadata` flags to `bo pack`.
- **ZIP Lossy Comment Badges**: Lossy archives are stamped with `Boa-Lossy: ...` comment tags, rendering clear warnings in `bo list` while maintaining standard unzipper compatibility.

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
