# 07-govulncheck — Tutorial Notes

Source: https://go.dev/doc/tutorial/govulncheck

---

## What this tutorial covers

`govulncheck` is Go's official vulnerability scanner. It scans your module's
dependency graph against the [Go vulnerability database](https://vuln.go.dev)
and reports only the vulnerabilities that your code **actually calls** —
filtering out the noise of vulnerabilities in code paths you never reach.

This is the key distinction from naively checking `go.sum` against a CVE list:
`govulncheck` does static call-graph analysis, so it tells you whether a
vulnerable function is reachable from your code, not just whether you have
the package installed.

---

## Project structure

```
07-govulncheck/
├── go.mod      ← module vuln.tutorial, pinned to safe golang.org/x/text v0.3.8
├── go.sum
├── main.go     ← simple language-tag parser using golang.org/x/text/language
└── NOTES.md    ← this file
```

---

## Install govulncheck

govulncheck is not part of the standard toolchain — install it once:

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
```

This places the binary in your `$GOBIN` (usually `~/go/bin`). Make sure that
directory is on your `PATH`.

---

## Reproducing the tutorial step by step

The `go.mod` in this folder already pins the fixed version (`v0.3.8`), so
`govulncheck` will report clean. To see the full tutorial experience:

### Step 1 — downgrade to the vulnerable version

```bash
go get golang.org/x/text@v0.3.5
```

### Step 2 — scan

```bash
govulncheck ./...
```

You will see two vulnerabilities reported:

```
Your code is affected by 1 vulnerability from 1 module.

Vulnerability #1: GO-2021-0113
  Due to improper index calculation, an incorrectly formatted language tag
  can cause Parse to panic via an out of bounds read. If Parse is used to
  process untrusted user inputs, this may be used as a vector for a denial
  of service attack.
  More info: https://pkg.go.dev/vuln/GO-2021-0113
  Module: golang.org/x/text
    Found in: golang.org/x/text@v0.3.5
    Fixed in: golang.org/x/text@v0.3.7
  Call stacks in your code:
    main.go:12:29: vuln.tutorial.main calls golang.org/x/text/language.Parse

=== Informational ===

Vulnerability #1: GO-2022-1059
  An attacker may cause a denial of service by crafting an Accept-Language
  header which ParseAcceptLanguage will take significant time to parse.
  More info: https://pkg.go.dev/vuln/GO-2022-1059
  Found in: golang.org/x/text@v0.3.5
  Fixed in: golang.org/x/text@v0.3.8
```

### Step 3 — upgrade to fix both

```bash
go get golang.org/x/text@v0.3.8
```

### Step 4 — scan again

```bash
govulncheck ./...
# No vulnerabilities found.
```

---

## Reading govulncheck output

### Actionable vs Informational

| Section | Meaning | Action needed? |
|---|---|---|
| **Actionable** (top section) | Your code has a call stack that reaches the vulnerable function | Yes — fix or mitigate |
| **Informational** | The vulnerable package is in your dependency graph but your code never calls the vulnerable function | Usually no |

The informational section exists to make you aware of what's in your tree,
but govulncheck deliberately avoids alarming you about vulnerabilities you
can't trigger. This "low-noise" design is the main advantage over plain CVE
list matching.

### Call stack output

```
Call stacks in your code:
  main.go:12:29: vuln.tutorial.main calls golang.org/x/text/language.Parse
```

This pinpoints the exact line in **your** code where the vulnerable function
is called. In a large codebase this saves you from grep-hunting.

---

## How to evaluate a vulnerability

Work through this checklist for each **Actionable** finding:

1. **Read the description** — does the vulnerability class apply to your use
   case? (e.g. DoS via untrusted input matters if you expose a public API;
   less so for a CLI tool only you run)

2. **Check the "Fixed in" version** — is a patch available? Most of the time
   the answer is yes.

3. **Check the call stack** — is the vulnerable call path reachable in
   production? Sometimes a test helper calls a vulnerable function but
   production code never does.

4. **Choose an action:**

   | Option | When to use |
   |---|---|
   | Upgrade to fixed version | Almost always the right answer |
   | Remove the dependency entirely | If the library is no longer needed |
   | Stop calling the vulnerable symbol | If upgrade is blocked and you can implement an alternative |
   | Accept the risk | Only if the vulnerability demonstrably cannot be triggered in your deployment |

For **Informational** findings: decide based on likelihood that your code
will gain a call path to the vulnerable function in the future. It is good
practice to upgrade anyway when the cost is low (same module, one `go get`).

---

## Key concepts

### The Go vulnerability database

`govulncheck` queries [https://vuln.go.dev](https://vuln.go.dev) — the
official Go vulnerability database maintained by the Go security team.
Entries are assigned Go-specific IDs (e.g. `GO-2021-0113`) and are also
mapped to CVE and GHSA identifiers.

You can browse entries at `https://pkg.go.dev/vuln/<ID>` or search the
database at `https://pkg.go.dev/vuln/`.

### Why call-graph analysis matters

A naive approach compares your `go.sum` against a list of vulnerable module
versions. This produces many false positives — you may depend on a module
that has a vulnerability in a function you never import or call.

`govulncheck` analyses the actual call graph of your binary. A vulnerability
only appears as **Actionable** if there is a chain of function calls from
your code to the vulnerable function. This means:

- Fewer alerts to investigate
- Higher confidence that each alert represents a real risk
- The call stack output tells you exactly where to look

### `govulncheck` vs `go mod audit` / Dependabot / Snyk

Those tools operate at the module-version level (is this version in a CVE
database?). `govulncheck` operates at the symbol level (does your code
actually call the vulnerable function?). Use `govulncheck` for Go projects;
the others can complement it for non-Go dependencies.

---

## Running govulncheck in CI

Add this to your CI pipeline to catch new vulnerabilities automatically:

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

`govulncheck` exits with a non-zero status when actionable vulnerabilities
are found, so it integrates naturally as a CI gate.

For GitHub Actions specifically, the official
[`golang/govulncheck-action`](https://github.com/golang/govulncheck-action)
handles installation and scanning in one step.
