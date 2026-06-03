# 02-modules — Tutorial Notes

Source: https://go.dev/doc/tutorial/create-module

---

## Module structure

```
02-modules/
├── greetings/          ← library module (example.com/greetings)
│   ├── go.mod
│   ├── greetings.go
│   └── greetings_test.go
└── hello/              ← executable module (example.com/hello)
    ├── go.mod
    └── hello.go
```

---

## Running the code

From the `hello/` directory:

```bash
go run .
```

---

## Running the tests

From the `greetings/` directory:

```bash
# minimal output — only shows failures
go test

# verbose output — lists every test and its result
go test -v
```

Expected output (all passing):

```
=== RUN   TestHelloName
--- PASS: TestHelloName (0.00s)
=== RUN   TestHelloEmpty
--- PASS: TestHelloEmpty (0.00s)
PASS
ok      example.com/greetings   0.364s
```

---

## Compile and install the application

This is the last topic of the tutorial. There are two relevant commands:

| Command       | What it does                                                    |
|---------------|-----------------------------------------------------------------|
| `go build`    | Compiles the package + dependencies into a binary in the **current directory**. Does not install. |
| `go install`  | Compiles **and** places the binary in the Go install directory so you can run it from anywhere. |

### 1. Build a local binary

From the `hello/` directory:

```bash
go build
```

This produces a `hello` (or `hello.exe` on Windows) binary in the same directory.
Run it directly:

```bash
# Linux / Mac
./hello

# Windows
hello.exe
```

### 2. Find the install path

Before installing, check where `go install` will place the binary:

```bash
go list -f '{{.Target}}'
# example output: /home/youruser/go/bin/hello
```

### 3. Make sure the install directory is on your PATH

If the directory shown above is not already in your shell's PATH, add it.

**Linux / Mac** (add to `~/.bashrc` or `~/.zshrc` for persistence):

```bash
export PATH=$PATH:/home/youruser/go/bin
```

**Windows**:

```cmd
set PATH=%PATH%;C:\Users\youruser\go\bin
```

Alternatively, change where Go installs binaries to a directory that is already on your PATH:

```bash
go env -w GOBIN=/path/to/your/bin
```

### 4. Install

From the `hello/` directory:

```bash
go install
```

### 5. Run from anywhere

Once installed, you can call the program by name from any directory:

```bash
hello
# map[Darrin:Hail, Darrin! Well met! Gladys:Hi, Gladys. Welcome! Samantha:Great to see you, Samantha!]
```

---

## Key differences: `go run` vs `go build` vs `go install`

| Command      | Compiles | Produces binary | Binary location          | Use when                          |
|--------------|----------|-----------------|--------------------------|-----------------------------------|
| `go run .`   | Yes      | Temporary       | Cleaned up after exit    | Quick iteration during development |
| `go build`   | Yes      | Yes             | Current directory        | Testing the final binary locally  |
| `go install` | Yes      | Yes             | `$GOBIN` / `$GOPATH/bin` | Making the program globally runnable |

---

## The `replace` directive (local development)

The `hello/go.mod` file uses a `replace` directive:

```
replace example.com/greetings => ../greetings
```

This tells the Go toolchain to resolve `example.com/greetings` from the local
filesystem instead of downloading it from a remote registry. This is only needed
while the library hasn't been published yet. Once published with a real version
tag, the `replace` directive is removed and `go.mod` would simply have:

```
require example.com/greetings v1.0.0
```
