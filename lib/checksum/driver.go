package checksum

import (
	"bufio"
	"encoding/hex"
	"flag"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
)

type Driver struct {
	Name                                                          string
	New                                                           func() hash.Hash
	Version                                                       string
	Flags                                                         *flag.FlagSet
	Binary, Text, Check, Tag, Zero                                *bool
	IgnoreMissing, Quiet, Status, Strict, Warn, Help, VersionFlag *bool
}

func (d *Driver) BindFlags() {
	fs := flag.NewFlagSet(strings.ToLower(d.Name)+"sum", flag.ExitOnError)
	d.Flags = fs

	d.Binary = fs.Bool("b", false, "")
	fs.BoolVar(d.Binary, "binary", false, "")
	d.Text = fs.Bool("t", true, "")
	fs.BoolVar(d.Text, "text", true, "")

	d.Check = fs.Bool("c", false, "")
	fs.BoolVar(d.Check, "check", false, "")

	d.Tag = fs.Bool("tag", false, "")
	d.Zero = fs.Bool("z", false, "")
	fs.BoolVar(d.Zero, "zero", false, "")

	d.IgnoreMissing = fs.Bool("ignore-missing", false, "")
	d.Quiet = fs.Bool("quiet", false, "")
	d.Status = fs.Bool("status", false, "")
	d.Strict = fs.Bool("strict", false, "")
	d.Warn = fs.Bool("w", false, "")
	fs.BoolVar(d.Warn, "warn", false, "")

	d.Help = fs.Bool("h", false, "")
	fs.BoolVar(d.Help, "help", false, "")

	d.VersionFlag = fs.Bool("version", false, "")
}

func (d *Driver) Usage() {
	name := strings.ToLower(d.Name) + "sum"
	fmt.Fprintf(os.Stdout, "Usage: %s [OPTION]... [FILE]...\n", name)
	fmt.Fprintf(os.Stdout, "Print or check %s (%d-bit) checksums.\n\n", d.Name, d.New().Size()*8)
	fmt.Fprint(os.Stdout,
		"  -b, --binary          read in binary mode\n",
		"  -c, --check           read checksums from the FILEs and check them\n",
		"      --tag             create a BSD-style checksum\n",
		"  -t, --text            read in text mode (default)\n",
		"  -z, --zero            end each output line with NUL\n\n",
		"Verification flags:\n",
		"      --ignore-missing  skip missing files\n",
		"      --quiet           don’t print OK for each verified file\n",
		"      --status          suppress all output; exit code shows success\n",
		"      --strict          non-zero on bad checksum lines\n",
		"  -w, --warn            warn about improperly formatted lines\n\n",
		"  -h, --help            display this help and exit\n",
		"      --version         output version information and exit\n")
}

func escapeName(name string) string {
	var b strings.Builder
	for _, r := range name {
		switch r {
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		case '\\':
			b.WriteString("\\\\")
		default:
			if r < 32 || r == 127 {
				b.WriteString(fmt.Sprintf("\\x%02x", r))
			} else {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

func computeSum(h hash.Hash, r io.Reader) (string, error) {
	h.Reset()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (d *Driver) printSum(sum, name string) {
	out := name
	if !*d.Zero && !*d.Tag {
		out = escapeName(name)
	}
	if !*d.Zero && !*d.Tag && out != name {
		fmt.Print("\\")
	}

	if *d.Tag {
		fmt.Printf("%s (%s) = %s", d.Name, out, sum)
		if *d.Zero {
			fmt.Print("\x00")
		} else {
			fmt.Println()
		}
	} else {
		if *d.Binary {
			fmt.Printf("%s *%s", sum, out)
		} else {
			fmt.Printf("%s  %s", sum, out)
		}
		if *d.Zero {
			fmt.Print("\x00")
		} else {
			fmt.Println()
		}
	}
}

func (d *Driver) verifyLine(line, fname string, lineno int) (string, string, error) {
	line = strings.TrimRight(line, "\r\n\x00")

	if strings.HasPrefix(line, d.Name) {
		rest := strings.TrimSpace(line[len(d.Name):])
		if idx := strings.Index(rest, ")"); idx != -1 {
			var namePart string
			switch {
			case strings.HasPrefix(rest, "("):
				namePart = rest[1:idx]
			case strings.HasPrefix(rest, " ("):
				namePart = rest[2:idx]
			default:
				break
			}
			post := strings.TrimSpace(rest[idx+1:])
			if strings.HasPrefix(post, "=") {
				hashStr := strings.TrimSpace(post[1:])
				if namePart != "" && len(hashStr) >= d.New().Size()*2 {
					return hashStr, namePart, nil
				}
			}
		}
		return "", "", fmt.Errorf("%s: %d: improperly formatted %s checksum line",
			fname, lineno, d.Name)
	}

	min := d.New().Size()*2 + 3
	if len(line) < min {
		return "", "", fmt.Errorf("%s: %d: improperly formatted %s checksum line", fname, lineno, d.Name)
	}
	hashStr := line[:d.New().Size()*2]
	pos := d.New().Size() * 2
	if line[pos] != ' ' {
		return "", "", fmt.Errorf("%s: %d: improperly formatted %s checksum line", fname, lineno, d.Name)
	}
	pos++
	mode := line[pos]
	if mode != ' ' && mode != '*' {
		return "", "", fmt.Errorf("%s: %d: improperly formatted %s checksum line", fname, lineno, d.Name)
	}
	pos++
	if pos < len(line) && line[pos] == ' ' {
		pos++
	}
	name := line[pos:]
	if name == "" {
		return "", "", fmt.Errorf("%s: %d: improperly formatted %s checksum line", fname, lineno, d.Name)
	}
	for _, r := range hashStr {
		if !strings.ContainsRune("0123456789abcdefABCDEF", r) {
			return "", "", fmt.Errorf("%s: %d: improperly formatted %s checksum line", fname, lineno, d.Name)
		}
	}
	return hashStr, name, nil
}

func (d *Driver) checkFiles(files []string) int {
	code, total, malformed, mismatches := 0, 0, 0, 0
	prog := strings.ToLower(d.Name) + "sum"

	for _, fname := range files {
		f, err := os.Open(fname)
		if err != nil {
			if os.IsNotExist(err) && *d.IgnoreMissing {
				continue
			}
			code = 1
			if *d.Warn {
				fmt.Fprintf(os.Stderr, "%s: %s: %v\n", prog, fname, err)
			}
			continue
		}
		reader := bufio.NewReader(f)
		lineno, valid, verified := 0, 0, 0

		for {
			line, errr := reader.ReadString('\n')
			if line == "" && errr == io.EOF {
				break
			}
			if strings.TrimRight(line, "\r\n\x00") == "" {
				if errr == io.EOF {
					break
				}
				continue
			}
			lineno++
			expect, path, verr := d.verifyLine(line, fname, lineno)
			if verr != nil {
				malformed++
				if *d.Strict {
					code = 1
				}
				if *d.Warn {
					fmt.Fprintf(os.Stderr, "%s: %v\n", prog, verr)
				}
				continue
			}
			valid++
			tgt, err2 := os.Open(path)
			if err2 != nil {
				if os.IsNotExist(err2) && *d.IgnoreMissing {
					continue
				}
				code = 1
				continue
			}
			verified++
			got, _ := computeSum(d.New(), tgt)
			tgt.Close()

			if got != expect {
				mismatches++
				if !*d.Status {
					fmt.Printf("%s: FAILED\n", path)
				}
				code = 1
			} else if !*d.Quiet && !*d.Status {
				fmt.Printf("%s: OK\n", path)
			}
			if errr == io.EOF {
				break
			}
		}
		f.Close()

		if *d.Check && *d.IgnoreMissing && verified == 0 {
			fmt.Fprintf(os.Stderr, "%s: %s: no file was verified\n", prog, fname)
			code = 1
		}
		total += valid
		if valid == 0 {
			fmt.Fprintf(os.Stderr, "%s: %s: no properly formatted checksum lines found\n", prog, fname)
			if *d.Check {
				code = 1
			}
		}
	}

	if !*d.Status {
		if malformed > 0 && (*d.Warn || *d.Strict || total > 0) {
			if malformed == 1 {
				fmt.Fprintf(os.Stderr, "%s: WARNING: 1 line is improperly formatted\n", strings.ToLower(d.Name)+"sum")
			} else {
				fmt.Fprintf(os.Stderr, "%s: WARNING: %d lines are improperly formatted\n", strings.ToLower(d.Name)+"sum", malformed)
			}
		}
		if mismatches > 0 {
			if mismatches == 1 {
				fmt.Fprintf(os.Stderr, "%s: WARNING: 1 computed checksum did NOT match\n", strings.ToLower(d.Name)+"sum")
			} else {
				fmt.Fprintf(os.Stderr, "%s: WARNING: %d computed checksums did NOT match\n", strings.ToLower(d.Name)+"sum", mismatches)
			}
		}
	}

	return code
}

func (d *Driver) Run() {
	d.BindFlags()
	d.Flags.Usage = d.Usage
	d.Flags.Parse(os.Args[1:])
	args := d.Flags.Args()
	prog := strings.ToLower(d.Name) + "sum"

	if *d.Help {
		d.Usage()
		os.Exit(0)
	}
	if *d.VersionFlag {
		fmt.Printf("%s (%s) %s\n", prog, d.Name, d.Version)
		os.Exit(0)
	}
	if *d.IgnoreMissing && !*d.Check {
		fmt.Fprintln(os.Stderr, prog+": the --ignore-missing option is meaningful only when verifying checksums")
		fmt.Fprintf(os.Stderr, "Try '%s --help' for more information.\n", prog)
		os.Exit(1)
	}

	if *d.Check {
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, prog+": no checksum files specified")
			os.Exit(1)
		}
		os.Exit(d.checkFiles(args))
	}

	if len(args) == 0 {
		sum, err := computeSum(d.New(), os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error computing sum: %v\n", err)
			os.Exit(1)
		}
		d.printSum(sum, "-")
		return
	}

	for _, path := range args {
		f, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			continue
		}
		sum, err := computeSum(d.New(), f)
		f.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			continue
		}
		d.printSum(sum, path)
	}
}
