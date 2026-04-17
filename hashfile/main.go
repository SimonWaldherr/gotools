// Description: A tool to compute checksums (MD5, SHA1, SHA256, SHA512) for one or more files.
// It prints the hash and filename in a format similar to md5sum / sha256sum.
package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"flag"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
)

func hashFile(path string, h hash.Hash) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h.Reset()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func main() {
	algo := flag.String("algo", "sha256", "Hash algorithm: md5, sha1, sha256, sha512 (comma-separated for multiple)")
	flag.Parse()

	files := flag.Args()
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: hashfile [flags] <file> [file...]")
		flag.Usage()
		os.Exit(1)
	}

	algos := strings.Split(strings.ToLower(*algo), ",")
	hashers := map[string]hash.Hash{}
	for _, a := range algos {
		a = strings.TrimSpace(a)
		switch a {
		case "md5":
			hashers[a] = md5.New()
		case "sha1":
			hashers[a] = sha1.New()
		case "sha256":
			hashers[a] = sha256.New()
		case "sha512":
			hashers[a] = sha512.New()
		default:
			fmt.Fprintf(os.Stderr, "Unknown algorithm: %s\n", a)
			os.Exit(1)
		}
	}

	exitCode := 0
	for _, path := range files {
		for _, a := range algos {
			a = strings.TrimSpace(a)
			sum, err := hashFile(path, hashers[a])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error hashing %s: %v\n", path, err)
				exitCode = 1
				continue
			}
			if len(algos) > 1 {
				fmt.Printf("%s  %s  (%s)\n", sum, path, a)
			} else {
				fmt.Printf("%s  %s\n", sum, path)
			}
		}
	}
	os.Exit(exitCode)
}
