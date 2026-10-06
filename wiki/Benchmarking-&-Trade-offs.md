# Benchmarking & Trade-offs

Boa features a high-precision benchmarking suite designed to analyze the trade-off between space savings and CPU throughput.

---

## 1. Running Benchmarks

```bash
# Benchmark directory across levels 0 to 9
bo bench ./my-project

# Visual side-by-side comparison charts
bo bench ./my-project --compare

# Average across 3 runs with JSON export
bo bench ./my-project --runs 3 --json > results.json
```

---

## 2. Interpreting Compression Levels

| Level | Mode | Characteristics | Best Use Case |
|---|---|---|---|
| **0** | Store | No compression, raw container wrapping, wire-speed I/O | Pre-compressed media, temporary bundles |
| **1** | Fastest | Greedy LZ77 match search, maximum throughput | Real-time streams, high-speed networks |
| **6** | Default ★ | Balanced match length and evaluation depth (~95% max ratio) | General purpose everyday compression |
| **9** | Maximum | Deep match chains, lazy evaluation, maximum space savings | Archival storage, cold backups |

---

## 3. The Efficiency Knee-Point

In compression theory, increasing compression levels from Level 6 to Level 9 exhibits diminishing returns:
- Space saved typically increases by only **0.5% to 1.5%**.
- CPU time and memory consumption increase by **200% to 400%**.

Boa automatically computes and highlights the **Best Balance Knee-Point** during benchmark comparisons.
