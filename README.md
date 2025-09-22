# 🥅 Goalie

[![Go Reference](https://pkg.go.dev/badge/github.com/ras0q/goalie.svg)](https://pkg.go.dev/github.com/ras0q/goalie) [![autofix enabled](https://shields.io/badge/autofix.ci-yes-success?logo=data:image/svg+xml;base64,PHN2ZyBmaWxsPSIjZmZmIiB2aWV3Qm94PSIwIDAgMTI4IDEyOCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cGF0aCB0cmFuc2Zvcm09InNjYWxlKDAuMDYxLC0wLjA2MSkgdHJhbnNsYXRlKC0yNTAsLTE3NTApIiBkPSJNMTMyNSAtMzQwcS0xMTUgMCAtMTY0LjUgMzIuNXQtNDkuNSAxMTQuNXEwIDMyIDUgNzAuNXQxMC41IDcyLjV0NS41IDU0djIyMHEtMzQgLTkgLTY5LjUgLTE0dC03MS41IC01cS0xMzYgMCAtMjUxLjUgNjJ0LTE5MSAxNjl0LTkyLjUgMjQxcS05MCAxMjAgLTkwIDI2NnEwIDEwOCA0OC41IDIwMC41dDEzMiAxNTUuNXQxODguNSA4MXExNSA5OSAxMDAuNSAxODAuNXQyMTcgMTMwLjV0MjgyLjUgNDlxMTM2IDAgMjU2LjUgLTQ2IHQyMDkgLTEyNy41dDEyOC41IC0xODkuNXExNDkgLTgyIDIyNyAtMjEzLjV0NzggLTI5OS41cTAgLTEzNiAtNTggLTI0NnQtMTY1LjUgLTE4NC41dC0yNTYuNSAtMTAzLjVsLTI0MyAtMzAwdi01MnEwIC0yNyAzLjUgLTU2LjV0Ni41IC01Ny41dDMgLTUycTAgLTg1IC00MS41IC0xMTguNXQtMTU3LjUgLTMzLjV6TTEzMjUgLTI2MHE3NyAwIDk4IDE0LjV0MjEgNTcuNXEwIDI5IC0zIDY4dC02LjUgNzN0LTMuNSA0OHY2NGwyMDcgMjQ5IHEtMzEgMCAtNjAgNS41dC01NCAxMi41bC0xMDQgLTEyM3EtMSAzNCAtMiA2My41dC0xIDU0LjVxMCA2OSA5IDEyM2wzMSAyMDBsLTExNSAtMjhsLTQ2IC0yNzFsLTIwNSAyMjZxLTE5IC0xNSAtNDMgLTI4LjV0LTU1IC0yNi41bDIxOSAtMjQydi0yNzZxMCAtMjAgLTUuNSAtNjB0LTEwLjUgLTc5dC01IC01OHEwIC00MCAzMCAtNTMuNXQxMDQgLTEzLjV6TTEyNjIgNjE2cS0xMTkgMCAtMjI5LjUgMzQuNXQtMTkzLjUgOTYuNWw0OCA2NCBxNzMgLTU1IDE3MC41IC04NXQyMDQuNSAtMzBxMTM3IDAgMjQ5IDQ1LjV0MTc5IDEyMXQ2NyAxNjUuNWg4MHEwIC0xMTQgLTc3LjUgLTIwNy41dC0yMDggLTE0OXQtMjg5LjUgLTU1LjV6TTgwMyA1OTVxODAgMCAxNDkgMjkuNXQxMDggNzIuNWwyMjEgLTY3bDMwOSA4NnE0NyAtMzIgMTA0LjUgLTUwdDExNy41IC0xOHE5MSAwIDE2NSAzOHQxMTguNSAxMDMuNXQ0NC41IDE0Ni41cTAgNzYgLTM0LjUgMTQ5dC05NS41IDEzNHQtMTQzIDk5IHEtMzcgMTA3IC0xMTUuNSAxODMuNXQtMTg2IDExNy41dC0yMzAuNSA0MXEtMTAzIDAgLTE5Ny41IC0yNnQtMTY5IC03Mi41dC0xMTcuNSAtMTA4dC00MyAtMTMxLjVxMCAtMzQgMTQuNSAtNjIuNXQ0MC41IC01MC41bC01NSAtNTlxLTM0IDI5IC01NCA2NS41dC0yNSA4MS41cS04MSAtMTggLTE0NSAtNzB0LTEwMSAtMTI1LjV0LTM3IC0xNTguNXEwIC0xMDIgNDguNSAtMTgwLjV0MTI5LjUgLTEyM3QxNzkgLTQ0LjV6Ii8+PC9zdmc+)](https://autofix.ci)


Goalie (/góʊli/) is a Go library designed to **reliably capture and collect errors from `defer`'d cleanup functions**, such as `file.Close()`, `conn.Close()`, or `tx.Rollback()`.

Named for its role, much like a **goalie (goalkeeper)**, Goalie ensures that no errors from deferred cleanup operations are missed at the end of Go function execution.

## The Story of Common `defer` Mistakes

It's late at night. You've just written some code to read a file and process its contents:

```go
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    // 🚨 Careful!
    defer f.Close()
    // ... do something ...
    return nil
}
```

At first glance, this looks fine. But running a static analysis tool like `errcheck` will already warn you:

```
errcheck: Error return value of `f.Close` is not checked
```

Unfortunately, **any error returned by `f.Close()` is silently ignored**. If `f.Close()` fails, you'll never know!

You might try to improve things by logging the error in a deferred function:

```diff
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
-   // 🚨 Careful!
-   defer f.Close()
+   defer func() {
+       if cerr := f.Close(); cerr != nil {
+           // 😶 Only logs the error, does not return it!
+           log.Printf("failed to close: %v", cerr)
+       }
+   }()
    // ... do something ...
    return nil
}
```

But here, while the error is just logged, **the error from `f.Close()` is not returned to the caller** and is effectively lost for upstream error handling.

Worse yet, after being nagged one too many times by errcheck’s warnings about unchecked errors in defer statements, a tired developer might be tempted to silence the error completely. The result? Adding a comment like `//nolint:errcheck` to just make the warning go away—this is the worst possible move:

```diff
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
-   // 🚨 Careful!
-   defer f.Close()
+   // 😱 Oh no! You can't be serious! Errors are just thrown away!
+   defer f.Close() //nolint:errcheck
    // ... do something ...
    return nil
}
```

This practice hides problems instead of solving them, making error detection even harder.

These patterns are surprisingly common—and subtle! Goalie exists to ensure that all errors from deferred cleanup are reliably collected and returned, never lost or forgotten.

## A Better Way: Goalie

With Goalie, you can ensure that all errors from deferred cleanup are properly collected and reported. No more silent failures or hidden cleanup mistakes!

> [!CAUTION]
> Goalie is only for handling errors from cleanup operations, not for general error handling.

```diff
-func processFile(path string) error {
+func processFile(path string) (err error) {
+   g := goalie.New()
+   // 👍 Use g.Collect to collect all captured errors at final!
+   defer g.Collect(&err)

    f, err := os.Open(path)
    if err != nil {
        return err
    }
-   // 🚨 Careful!
-   defer f.Close()
+   // 👍 Use g.Guard to capture errors from the deferred cleanup!
+   defer g.Guard(f.Close)
    // ... do something ...
    return nil
}
```

See [Godoc](https://pkg.go.dev/github.com/ras0q/goalie), [./goalie_test.go](./goalie_test.go) and [./_examples](./_examples) for details.

## Integrating Goalie into Existing Projects

For existing projects, we provide a migration tool to automatically insert Goalie's error handling patterns!

The tool will analyze your code and suggest fixes for any `defer` statements that might be missing error handling.

Run the migrator `usegoalie` on your project:

> [!CAUTION]
> Always review the changes made by the migrator, especially in complex functions, to ensure correctness.

```bash
# Check changes
go run github.com/ras0q/goalie/usegoalie/cmd/usegoalie@latest -diff -fix ./...

# Apply changes
go run github.com/ras0q/goalie/usegoalie/cmd/usegoalie@latest -fix ./...
```

After migration, you should organize imports.

```bash
go mod tidy
go run golang.org/x/tools/cmd/goimports@latest -w .
```

## Acknowledgement

This project was inspired by the error handling patterns found in
[`go.dev/x/pkgsite`](https://cs.opensource.google/go/x/pkgsite/+/master:internal/derrors/derrors.go;l=231-244;drc=c20a88edadfbe20d624856081ccf9de2a2e6b945).
