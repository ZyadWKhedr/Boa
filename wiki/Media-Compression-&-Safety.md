# Media Compression & Safety

Boa is engineered with a **Fail-Closed Safety Policy** to ensure critical files are never corrupted or degraded while optimizing multimedia assets.

---

## 1. Safety Architecture

```text
┌────────────────────────────────────────────────────────────────────────┐
│                        CONTENT CLASSIFICATION                          │
│               (Magic Bytes & Container Probing)                        │
└──────────────────────────────────┬─────────────────────────────────────┘
                                   │
         ┌─────────────────────────┴─────────────────────────┐
         ▼                                                   ▼
┌──────────────────────────────────┐        ┌────────────────────────────┐
│   MEDIA FILES (Image/Audio/Vid)  │        │   ALL OTHER / UNKNOWN      │
│   JPEG, PNG, MP3, WAV, MP4...    │        │   Code, Text, Docs, DBs... │
└────────────────┬─────────────────┘        └─────────────┬──────────────┘
                 │                                        │
        [Policy Check #1 & #2]                            │
                 │                                        │
        Lossy Permitted?                                  │
         ├── YES (with explicit consent)                  │
         │   └── JPEG Quality / PNG Palette               │
         │       Audio Bitrate / Video CRF                │
         │                                                │
         └── NO (Default) ────────────────────────────────┼─────────────────┐
                                                          ▼                 ▼
                                                   [100% LOSSLESS]   [100% LOSSLESS]
                                                   DEFLATE (1-9)      Zstandard / Store
```

### Core Safety Guarantees:
1. **100% Lossless by Default**: No file will ever be modified lossy unless the user explicitly opts in.
2. **Magic Byte Content Probing**: File classification relies on binary magic bytes rather than file extensions, preventing accidental re-encoding of renamed binaries or source code.
3. **Immutability of Source Files**: Source files on disk are strictly read-only and never modified in place.
4. **ZIP Comment Metadata Tracking**: Lossy archives are embedded with `Boa-Lossy: ...` comment headers, allowing `bo list` to display prominent warning badges while preserving 100% standard zip compatibility.

---

## 2. Supported Media Pipelines

### Images:
- **JPEG**: Pure Go discrete cosine transform (DCT) frequency reduction and quantization with configurable quality levels (1–100).
- **PNG**: Lossless Deflate line-filter optimization or 256-color palette quantization.
- **Metadata**: Optional stripping of EXIF, GPS, and camera metadata tags.

### Audio:
- Multi-codec support for AAC, Opus, and MP3 with configurable bitrates (64k, 96k, 128k, 192k).

### Video:
- Perceptual Constant Rate Factor (CRF) scaling and resolution downscaling (1080p, 720p, 480p).
