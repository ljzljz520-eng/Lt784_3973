# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
?   	example.com/permission-selector/cmd/selector	[no test files]
ok  	example.com/permission-selector/internal/api	0.015s
ok  	example.com/permission-selector/internal/audit	0.016s
ok  	example.com/permission-selector/internal/config	0.001s
ok  	example.com/permission-selector/internal/domain	0.001s
ok  	example.com/permission-selector/internal/metrics	0.014s
ok  	example.com/permission-selector/internal/org	0.014s
ok  	example.com/permission-selector/internal/policy	0.001s
ok  	example.com/permission-selector/internal/report	0.012s
ok  	example.com/permission-selector/internal/selector	0.013s
ok  	example.com/permission-selector/internal/store	0.010s
--- FAIL: TestWorkflow17 (0.00s)
    workflow17_test.go:24: result page should be readable after first dispatch: invalid page
FAIL
FAIL	example.com/permission-selector/internal/workflow	0.027s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Node.js version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/selector): exit `0`
- Frontend build (web): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Node.js version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/selector): exit `0`
- Frontend build (web): exit `0`
