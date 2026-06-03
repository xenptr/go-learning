module example.com/hello

go 1.21

// replace redirects the Go toolchain to look for example.com/greetings on the
// local filesystem instead of trying to download it from a remote repository.
// This is the standard approach when both modules live in the same repository
// or on the same machine during development, before the library is published.
replace example.com/greetings => ../greetings

require example.com/greetings v0.0.0-00010101000000-000000000000
