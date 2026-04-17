// Description: A tool to check if TCP ports are open on a remote host.
// It accepts a host and one or more ports and reports which ports are open or closed.
package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func checkPort(host string, port int, timeout time.Duration) bool {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func main() {
	host := flag.String("host", "localhost", "Hostname or IP address to check")
	portsFlag := flag.String("ports", "80", "Comma-separated list of ports to check (e.g. 22,80,443)")
	timeoutFlag := flag.Duration("timeout", 3*time.Second, "Connection timeout per port")
	parallel := flag.Bool("parallel", false, "Check all ports in parallel")
	flag.Parse()

	if *host == "" {
		fmt.Fprintln(os.Stderr, "Error: host is required")
		flag.Usage()
		os.Exit(1)
	}

	portStrings := strings.Split(*portsFlag, ",")
	var ports []int
	for _, ps := range portStrings {
		ps = strings.TrimSpace(ps)
		// support ranges like 80-85
		if strings.Contains(ps, "-") {
			parts := strings.SplitN(ps, "-", 2)
			start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err1 != nil || err2 != nil || start > end {
				fmt.Fprintf(os.Stderr, "Invalid port range: %s\n", ps)
				os.Exit(1)
			}
			for p := start; p <= end; p++ {
				ports = append(ports, p)
			}
		} else {
			p, err := strconv.Atoi(ps)
			if err != nil || p < 1 || p > 65535 {
				fmt.Fprintf(os.Stderr, "Invalid port: %s\n", ps)
				os.Exit(1)
			}
			ports = append(ports, p)
		}
	}

	anyOpen := false

	if *parallel {
		type result struct {
			port int
			open bool
		}
		results := make([]result, len(ports))
		var wg sync.WaitGroup
		for i, port := range ports {
			wg.Add(1)
			go func(idx, p int) {
				defer wg.Done()
				results[idx] = result{port: p, open: checkPort(*host, p, *timeoutFlag)}
			}(i, port)
		}
		wg.Wait()
		for _, r := range results {
			if r.open {
				fmt.Printf("%s:%d  open\n", *host, r.port)
				anyOpen = true
			} else {
				fmt.Printf("%s:%d  closed\n", *host, r.port)
			}
		}
	} else {
		for _, port := range ports {
			if checkPort(*host, port, *timeoutFlag) {
				fmt.Printf("%s:%d  open\n", *host, port)
				anyOpen = true
			} else {
				fmt.Printf("%s:%d  closed\n", *host, port)
			}
		}
	}

	if !anyOpen {
		os.Exit(1)
	}
}
