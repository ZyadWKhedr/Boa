# Benchmark Test Dataset

This directory contains a small, reproducible, multi-format dataset designed for benchmarking Boa's compression engine across different Deflate levels.

### Contents:
- `data.json`: Structured JSON document.
- `code.go`: Sample Go source code.
- `records.csv`: Tabular CSV records.
- `document.txt`: Technical text documentation.

### Running Benchmarks on this Dataset:
```bash
# Compare all Deflate levels (0 through 9)
bo bench ./testdata/bench

# Include visual ASCII comparison bars
bo bench ./testdata/bench --compare

# Run 3 iterations per level and compute average throughput
bo bench ./testdata/bench --runs 3

# Export to JSON
bo bench ./testdata/bench --json

# Export to CSV
bo bench ./testdata/bench --csv
```
