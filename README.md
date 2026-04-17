# gotools
A repository for tools that aren't interesting enough to have their own repo.

## Tools

### CommitPlanner
Commit planner is a tool to plan commits for a git repository.
It reads a file with the commits and applies them to the git repository.

### Deadline
A tool to check for deadlines in source code.
It searches for the @CHECK annotation and checks if the deadline has passed.
If it has, it prints the line and returns an error.

### Easter
Calculates the date of Easter for a given year using the Gauss Easter formula

### invrevproxy
// wip

### WeSoc
A simple WebSocket client written in Go

### jsonformat
A tool to format and validate JSON from a file or stdin.
It reads JSON input, validates it, and writes it to stdout in pretty-printed or compact form.
Flags: `-file`, `-compact`, `-validate`, `-indent`

### portcheck
A tool to check if TCP ports are open on a remote host.
Accepts a host and one or more ports (including ranges like `80-90`) and reports which are open or closed.
Flags: `-host`, `-ports`, `-timeout`, `-parallel`

### envcheck
A tool to verify that required environment variables are set.
Reads variable names from flags or a file and exits with a non-zero code if any are missing.
Useful in CI/CD pipelines and startup scripts.
Flags: `-vars`, `-file`, `-quiet`

### hashfile
A tool to compute checksums (MD5, SHA1, SHA256, SHA512) for one or more files.
Output format is compatible with md5sum / sha256sum.
Supports multiple algorithms at once via comma-separated list.
Flags: `-algo`

