# Security Model

Security is a foundational design pillar of Boa. Standard unzipping tools have historically suffered from vulnerabilities such as path traversals, archive bombs, and symlink escapes. Boa incorporates active runtime defenses against these threats.

---

## 1. Zip-Slip Defense

Zip-Slip is an arbitrary file overwrite vulnerability occurring when archives contain entry filenames with relative path traversals (e.g. `../../../../etc/passwd` or `subfolder/../../file.txt`).

### Boa's Defense Mechanism:
1. **Canonical Clean**: Every incoming archive entry path is passed through `filepath.Clean`.
2. **Boundary Validation**: The resolved absolute destination path is compared against the extraction root using `filepath.Rel`.
3. **Fail-Closed Abort**: Any entry attempting to traverse outside the designated output folder is rejected with a hard security violation before opening a file handle.

```go
// Path Validation Rule:
rel, err := filepath.Rel(destDir, targetPath)
if err != nil || strings.HasPrefix(rel, "..") || rel == "." {
    return fmt.Errorf("security violation: path traversal detected (%s)", rawPath)
}
```

---

## 2. Malicious Symlink Protection

Symlinks inside untrusted archives can point to sensitive system locations (`/etc/`, `~/.ssh/`). Boa sanitizes symlink targets and prevents creating symlinks that point outside the extraction boundary.

---

## 3. Atomic Staging & File Locking

- During compression, Boa streams data to a temporary file (`.compressor_tmp_*.zip`).
- Upon completion and validation, the file is atomically renamed into place.
- Interrupted or failed compressions never leave corrupt output files in your working tree.
