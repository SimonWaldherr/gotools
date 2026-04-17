// Description: A tool to verify that required environment variables are set.
// It reads variable names from flags or a file and exits with an error if any are missing.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	varsFlag := flag.String("vars", "", "Comma-separated list of required environment variable names")
	fileFlag := flag.String("file", "", "Path to a file containing one variable name per line")
	quietFlag := flag.Bool("quiet", false, "Suppress output; only use exit code")
	flag.Parse()

	var required []string

	if *varsFlag != "" {
		for _, v := range strings.Split(*varsFlag, ",") {
			v = strings.TrimSpace(v)
			if v != "" {
				required = append(required, v)
			}
		}
	}

	if *fileFlag != "" {
		f, err := os.Open(*fileFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			required = append(required, line)
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
	}

	// Also accept bare arguments on the command line
	required = append(required, flag.Args()...)

	if len(required) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: envcheck -vars VAR1,VAR2 or envcheck VAR1 VAR2")
		flag.Usage()
		os.Exit(1)
	}

	missing := []string{}
	for _, name := range required {
		if _, ok := os.LookupEnv(name); !ok {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		if !*quietFlag {
			for _, name := range missing {
				fmt.Fprintf(os.Stderr, "Missing environment variable: %s\n", name)
			}
		}
		os.Exit(1)
	}

	if !*quietFlag {
		fmt.Printf("All %d required environment variable(s) are set.\n", len(required))
	}
}
