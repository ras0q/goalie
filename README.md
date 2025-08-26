# 🥅 Goalie

[![Go Reference](https://pkg.go.dev/badge/github.com/ras0q/goalie.svg)](https://pkg.go.dev/github.com/ras0q/goalie)

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
