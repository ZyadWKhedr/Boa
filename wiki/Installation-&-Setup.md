# Installation & Setup

Boa is distributed as a single standalone, statically compiled binary with zero external dependencies.

---

## 1. Quick Install via Script (macOS & Linux)

The recommended installation method uses the official universal install script:

```bash
curl -fsSL https://raw.githubusercontent.com/ZyadWKhedr/Boa/main/install.sh | bash
```

### What the installer does:
- Detects your OS (`darwin` / `linux`) and architecture (`arm64` / `amd64`).
- Downloads the latest matching release binary from GitHub.
- Places the binary into `~/.local/bin/boa`.
- Signs the binary with `codesign` on macOS (preventing AMFI security kills).
- Creates convenient symlinks: `bo` and `compressor`.

### Custom Release Installation
To install a specific version:
```bash
curl -fsSL https://raw.githubusercontent.com/ZyadWKhedr/Boa/main/install.sh | bash -s -- v0.3.3
```

---

## 2. Install via Go Toolchain

If you have Go installed (1.21+):

```bash
go install github.com/ZyadWKhedr/Boa@latest
```

Ensure `$(go env GOPATH)/bin` is in your `$PATH`.

---

## 3. Build from Source

```bash
git clone https://github.com/ZyadWKhedr/Boa.git
cd Boa
make test
make install
```

---

## 4. Setting up your `$PATH`

If `bo` is not recognized after installation, ensure `~/.local/bin` is in your environment `$PATH`.

### For Zsh (macOS Default & Linux):
```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### For Bash:
```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

---

## 5. Verifying Installation

Check your installed version and runtime configuration:

```bash
bo version
```

Output:
```text
 ____                 
| __ )  ___   __ _    
|  _ \ / _ \ / _` |   https://github.com/ZyadWKhedr/Boa
| |_) | (_) | (_| |   Tight, fast, lossless compression for your files.
|____/ \___/ \__,_|   

 ── Boa Runtime Environment ──
   Version          v0.3.3
   Go Runtime       go1.23.0
   Platform         darwin/arm64
   Concurrency      10 logical threads
   Repository       https://github.com/ZyadWKhedr/Boa
```
