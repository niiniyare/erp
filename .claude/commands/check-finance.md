# Check Finance Module

Run tests and validate the finance module implementation.

```bash
make test-unit | grep -E "(finance|PASS|FAIL)" && echo "---" && find internal/core/finance -name "*.go" | wc -l && echo "finance Go files found"
```

This command tests the finance module and shows implementation status.