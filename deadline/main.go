// Description: A tool to check for deadlines in source code.
// It searches for the @CHECK annotation and checks if the deadline has passed.
// If it has, it prints the line and returns an error.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

func main() {
	dir := flag.String("dir", ".", "The directory to search for deadlines")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := checkDeadlines(ctx, *dir); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func checkDeadlines(ctx context.Context, dir string) error {
	deadlineExceeded := false

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			if processFile(path) {
				deadlineExceeded = true
			}
		}
		return nil
	})

	if deadlineExceeded {
		return fmt.Errorf("at least one deadline exceeded")
	}
	return err
}

func processFile(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer file.Close()

	regex := regexp.MustCompile(`@CHECK\((\d{4}-\d{2}-\d{2});[^;]*;[^;]*;[^;]*;[^;]*\)`)
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	deadlineExceeded := false

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		matches := regex.FindStringSubmatch(line)

		if len(matches) > 1 {
			deadline, err := time.Parse("2006-01-02", matches[1])
			if err != nil {
				fmt.Printf("Invalid date format in file: %s, line: %d\n", filename, lineNumber)
				continue
			}

			now := time.Now()
			if now.After(deadline) {
				fmt.Printf("Deadline exceeded in file: %s, line: %d\nDEADLINE: %s\n", filename, lineNumber, line)
				deadlineExceeded = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error scanning file: %s, error: %v\n", filename, err)
	}

	return deadlineExceeded
}
