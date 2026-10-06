# Welcome to the Boa Wiki 🐍

Welcome to the official documentation for **Boa**, the fast, secure, and interactive compression engine for the terminal.

---

## Navigation

- **[Installation & Setup](Installation-&-Setup)** — Install via curl script, Go toolchain, or precompiled release binaries.
- **[Interactive TUI Guide](Interactive-TUI-Guide)** — Master the Bubble Tea terminal dashboard, file explorer, and media wizard.
- **[CLI Reference](CLI-Reference)** — Complete guide to all flags, subcommands, and streaming options.
- **[Media Compression & Safety](Media-Compression-&-Safety)** — Understand the fail-closed lossless policy and media optimization pipelines.
- **[Benchmarking & Trade-offs](Benchmarking-&-Trade-offs)** — Analyze throughput speeds, space savings, and algorithm efficiency knee-points.
- **[Security Model](Security-Model)** — Details on Zip-Slip path defense, canonical boundary sanitization, and atomic operations.

---

## Quick Start

```bash
# Install latest release
curl -fsSL https://raw.githubusercontent.com/ZyadWKhedr/Boa/main/install.sh | bash

# Launch interactive terminal UI
bo

# Quick CLI compression
bo compress ./my-folder

# Safe extraction
bo extract ./my-folder.zip
```

---

## Why Boa?

1. **Safety First**: Code, text, documents, databases, and unknown binaries are **guaranteed 100% lossless**.
2. **Interactive Terminal Experience**: Built with Charm Bubble Tea & Lip Gloss for a fluid, mouse-free terminal workflow.
3. **Media Awareness**: Perceptual optimization for JPEG, PNG, audio (AAC/Opus/MP3), and video (CRF downscaling) with explicit consent modals.
4. **Zip-Slip Immune**: Defends against path traversal attacks (`../../`) and escaping symlinks before any write to disk.
5. **Multi-Engine Power**: Choose standard DEFLATE, raw Store, or modern Zstandard (`zstd`).
