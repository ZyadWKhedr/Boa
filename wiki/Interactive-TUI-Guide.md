# Interactive TUI Guide

Boa includes a fullscreen Terminal User Interface powered by [Charm Bubble Tea](https://github.com/charmbracelet/bubbletea) and styled with [Lip Gloss](https://github.com/charmbracelet/lipgloss).

Launch the TUI simply by running:
```bash
bo
```

---

## 1. Dashboard Navigation

Upon launching, the interactive menu displays the main options:

```text
 ____                 
| __ )  ___   __ _    
|  _ \ / _ \ / _` |   https://github.com/ZyadWKhedr/Boa
| |_) | (_) | (_| |   Tight, fast, lossless compression for your files.
|____/ \___/ \__,_|   Version v0.3.1  ·  Interactive compression toolkit

➤ 1.  Compress         Package files into a zip archive with smart media options
  2.  Extract          Safely unzip archives with Zip-Slip path defense
  3.  Browse Archive   Inspect files, sizes, ratios & lossy metadata
  4.  Learn            Educational guide: how compression algorithms work
  5.  More...          Benchmarks, level comparison, system status & tools

↑↓/jk Navigate  ·  Enter Select  ·  1-5 Jump  ·  ? Learn  ·  q Quit
```

### Keybindings:
| Key | Action |
|---|---|
| `↑` / `k` | Move selection up |
| `↓` / `j` | Move selection down |
| `1` – `5` | Quick jump to menu item |
| `Enter` | Select active action |
| `?` | Open embedded compression guide |
| `Esc` / `q` | Return to previous menu or exit |

---

## 2. In-Terminal File Explorer

When choosing **Compress** or **Extract**, Boa opens an interactive directory picker with real-time size indicators:

```text
 📁 Select Folder or File to Compress
 Current Directory: ~/Projects/media-gallery

 ➤ 📁  .. (parent directory)
   📁  audio/
   📁  documents/
   📁  photos/
   📁  videos/
   📄  Makefile (1.2 KB)
   📄  README.md (3.4 KB)
   📦  backup.zip (14.2 MB)

 ↑↓/jk Navigate  ·  →/l Open Dir  ·  Enter Select  ·  Space Current Dir  ·  Esc Cancel
```

---

## 3. Multi-Step Media Wizard

If the selected folder contains images, audio, or video files, Boa initiates a step-by-step optimization wizard:

```text
 🐍 Boa Compression Wizard (Step 2 of 4)
 ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 📸 Image Compression (24 files · 48.2 MB)
 Techniques: JPEG quality recompression, PNG palette reduction

 Mode:     [ Lossless ]  ▶ [ Lossy ] ◀
 Quality:  [████████████████░░░░]  80%  (←/→ to adjust)

 📊 Estimated Total Size: 18.5 MB ~ 24.1 MB (Saved: ~50% to ~62%)

 ℹ Tip: 80% retains crisp visual clarity for 99% of viewing contexts.
 Press '?' for in-depth educational guide on JPEG Quality.

 ────────────────────────────────────────────────────────────
 ↑↓ Switch Row  ·  ←→ Adjust  ·  ? Help  ·  Enter Next  ·  Esc Cancel
```

### Wizard Capabilities:
1. **Dynamic Size Estimator**: Automatically scans file types and predicts estimated compressed archive sizes before compression.
2. **Contextual Knowledge Modals**: Press `?` at any stage to read an explanation of the underlying compression algorithm (e.g. DCT frequency reduction, LZ77 sliding window, quantization).
3. **Explicit Lossy Confirmation**: If lossy modes are selected, a safety summary warns of quality trade-offs prior to archive generation.
