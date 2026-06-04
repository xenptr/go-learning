// Package main demonstrates a simple program that parses BCP 47 language tags.
//
// Source: https://go.dev/doc/tutorial/govulncheck
//
// The program itself is intentionally minimal — the real subject of this
// tutorial is the govulncheck tool, not the Go code. See NOTES.md for the
// full workflow.
//
// The tutorial deliberately starts with a vulnerable version of
// golang.org/x/text (v0.3.5) and upgrades to v0.3.8 to fix two CVEs:
//
//   GO-2021-0113  language.Parse can panic on malformed tags (DoS risk)
//                 Fixed in: v0.3.7
//
//   GO-2022-1059  language.ParseAcceptLanguage is slow on crafted input (DoS risk)
//                 Fixed in: v0.3.8
//
// This go.mod already pins v0.3.8, so govulncheck reports no vulnerabilities.
// If you want to reproduce the tutorial step-by-step, downgrade first:
//
//	go get golang.org/x/text@v0.3.5
//	govulncheck ./...          ← shows 1 actionable + 1 informational vuln
//	go get golang.org/x/text@v0.3.8
//	govulncheck ./...          ← shows "No vulnerabilities found"
//
// Usage:
//
//	go run . en-US zh-Hans fr invalid-tag
package main

import (
	"fmt"
	"os"

	// golang.org/x/text/language implements BCP 47 language tag parsing.
	// BCP 47 tags look like "en", "en-US", "zh-Hans-CN", etc.
	// The vulnerability GO-2021-0113 was in the Parse function of this package.
	"golang.org/x/text/language"
)

func main() {
	// os.Args[0] is the program name; [1:] skips it to get user-supplied args.
	for _, arg := range os.Args[1:] {
		// language.Parse converts a BCP 47 string into a language.Tag value.
		// At v0.3.5 this function could panic on a malformed tag if the input
		// came from an untrusted source — a denial-of-service vector.
		tag, err := language.Parse(arg)
		if err != nil {
			// Malformed tag: report the error and continue.
			fmt.Printf("%s: error: %v\n", arg, err)
		} else if tag == language.Und {
			// language.Und represents an undefined/unknown tag.
			fmt.Printf("%s: undefined\n", arg)
		} else {
			// Successfully parsed: print the canonical form of the tag.
			fmt.Printf("%s: tag %s\n", arg, tag)
		}
	}
}
