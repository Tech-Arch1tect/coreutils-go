package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	flagBinary        bool
	flagText          bool
	flagCheck         bool
	flagTag           bool
	flagZero          bool
	flagIgnoreMissing bool
	flagQuiet         bool
	flagStatus        bool
	flagStrict        bool
	flagWarn          bool
	flagHelp          bool
	flagVersion       bool
)

// set up flags; text mode default is true, binary default false
func init() {
	flag.BoolVar(&flagBinary, "b", false, "")
	flag.BoolVar(&flagBinary, "binary", false, "")
	flag.BoolVar(&flagText, "t", true, "")
	flag.BoolVar(&flagText, "text", true, "")

	flag.BoolVar(&flagCheck, "c", false, "")
	flag.BoolVar(&flagCheck, "check", false, "")

	flag.BoolVar(&flagTag, "tag", false, "")
	flag.BoolVar(&flagZero, "z", false, "")
	flag.BoolVar(&flagZero, "zero", false, "")

	flag.BoolVar(&flagIgnoreMissing, "ignore-missing", false, "")
	flag.BoolVar(&flagQuiet, "quiet", false, "")
	flag.BoolVar(&flagStatus, "status", false, "")
	flag.BoolVar(&flagStrict, "strict", false, "")
	flag.BoolVar(&flagWarn, "w", false, "")
	flag.BoolVar(&flagWarn, "warn", false, "")

	flag.BoolVar(&flagHelp, "h", false, "")
	flag.BoolVar(&flagHelp, "help", false, "")
	flag.BoolVar(&flagVersion, "version", false, "")
}

func usage() {
	fmt.Fprintf(os.Stdout, "Usage: %s [OPTION]... [FILE]...\n", os.Args[0])
	fmt.Fprintln(os.Stdout, "Print or check MD5 (128-bit) checksums.")
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "  -b, --binary          read in binary mode")
	fmt.Fprintln(os.Stdout, "  -c, --check           read checksums from the FILEs and check them")
	fmt.Fprintln(os.Stdout, "      --tag             create a BSD-style checksum")
	fmt.Fprintln(os.Stdout, "  -t, --text            read in text mode (default)")
	fmt.Fprintln(os.Stdout, "  -z, --zero            end each output line with NUL")
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "Verification flags:")
	fmt.Fprintln(os.Stdout, "      --ignore-missing  skip missing files")
	fmt.Fprintln(os.Stdout, "      --quiet           don’t print OK for each verified file")
	fmt.Fprintln(os.Stdout, "      --status          suppress all output; exit code shows success")
	fmt.Fprintln(os.Stdout, "      --strict          non-zero on bad checksum lines")
	fmt.Fprintln(os.Stdout, "  -w, --warn            warn about improperly formatted lines")
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "  -h, --help            display this help and exit")
	fmt.Fprintln(os.Stdout, "      --version         output version information and exit")
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
			// non-printable ASCII or DEL
			if r < 32 || r == 127 {
				// use \xHH hex escape for control characters
				b.WriteString(fmt.Sprintf("\\x%02x", r))
			} else {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

func computeSum(r io.Reader) (string, error) {
	h := md5.New()
	// copy all data into MD5 hash
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	// Sum(nil) returns raw bytes; hex.EncodeToString makes the hex representation
	return hex.EncodeToString(h.Sum(nil)), nil
}

func printSum(sum, name string) {
	outName := name
	// if not in zero-terminated or BSD-tag mode, escape special chars
	if !flagZero && !flagTag {
		outName = escapeName(name)
	}
	if !flagZero && !flagTag && outName != name {
		// leading backslash indicates escaping occurred
		fmt.Print("\\")
	}

	if flagTag {
		// BSD-style: "MD5 (filename) = checksum"
		fmt.Printf("MD5 (%s) = %s", outName, sum)
	} else {
		// GNU style: checksum, space(s), mode indicator, filename
		if flagBinary {
			fmt.Printf("%s *%s", sum, outName) // '*' denotes binary mode
		} else {
			fmt.Printf("%s  %s", sum, outName) // two spaces for text mode
		}
		if flagZero {
			fmt.Print("\x00") // NUL terminator
		} else {
			fmt.Println()
		}
	}
}

func verifyLine(line, filename string, lineno int) (sum, name string, err error) {
	// strip trailing CR, LF, or NUL
	line = strings.TrimRight(line, "\r\n\x00")

	// handle BSD-style "MD5 (name) = sum"
	if strings.HasPrefix(line, "MD5") {
		rest := strings.TrimSpace(line[3:])
		if idx := strings.Index(rest, ")"); idx != -1 {
			namePart := rest[1:idx]
			after := strings.TrimSpace(rest[idx+1:])
			if strings.HasPrefix(after, "=") {
				sumPart := strings.TrimSpace(after[1:])
				if namePart != "" && len(sumPart) >= 32 {
					return sumPart, namePart, nil
				}
			}
		}
		return "", "", fmt.Errorf("%s: %d: improperly formatted MD5 checksum line", filename, lineno)
	}

	// expect at least 32-char hex sum + space + mode + optional space + filename
	if len(line) < 34 {
		return "", "", fmt.Errorf("%s: %d: improperly formatted MD5 checksum line", filename, lineno)
	}

	sum = line[:32]
	pos := 32
	if line[pos] != ' ' {
		return "", "", fmt.Errorf("%s: %d: improperly formatted MD5 checksum line", filename, lineno)
	}
	pos++

	mode := line[pos]
	// mode must be ' ' (text) or '*' (binary)
	if mode != ' ' && mode != '*' {
		return "", "", fmt.Errorf("%s: %d: improperly formatted MD5 checksum line", filename, lineno)
	}
	pos++

	// skip extra space before filename, if present
	if pos < len(line) && line[pos] == ' ' {
		pos++
	}

	name = line[pos:]
	if name == "" {
		return "", "", fmt.Errorf("%s: %d: improperly formatted MD5 checksum line", filename, lineno)
	}

	// verify sum is valid hex
	for _, r := range sum {
		if !('0' <= r && r <= '9' || 'a' <= r && r <= 'f' || 'A' <= r && r <= 'F') {
			return "", "", fmt.Errorf("%s: %d: improperly formatted MD5 checksum line", filename, lineno)
		}
	}
	return sum, name, nil
}

func checkFiles(files []string) int {
	exitCode := 0
	totalValid := 0
	malformed := 0
	mismatches := 0

	for _, f := range files {
		file, err := os.Open(f)
		if err != nil {
			// missing file may be ignored if flagIgnoreMissing set
			if os.IsNotExist(err) && flagIgnoreMissing {
				continue
			}
			exitCode = 1
			if flagWarn {
				fmt.Fprintf(os.Stderr, "md5sum: %s: %v\n", f, err)
			}
			continue
		}

		validLines := 0
		verifiedCount := 0
		r := bufio.NewReader(file)
		lineno := 0

		for {
			rawLine, err := r.ReadString('\n')
			// EOF with no data => break
			if rawLine == "" && err == io.EOF {
				break
			}
			// skip blank lines
			if strings.TrimRight(rawLine, "\r\n\x00") == "" {
				if err == io.EOF {
					break
				}
				continue
			}
			lineno++

			sum, name, ferr := verifyLine(rawLine, f, lineno)
			if ferr != nil {
				malformed++
				if flagStrict {
					exitCode = 1
				}
				if flagWarn {
					fmt.Fprintf(os.Stderr, "md5sum: %s\n", ferr)
				}
				continue
			}

			validLines++
			target, err2 := os.Open(name)
			if err2 != nil {
				// skip missing targets if ignoring
				if os.IsNotExist(err2) && flagIgnoreMissing {
					continue
				}
				exitCode = 1
				continue
			}

			verifiedCount++
			digest, _ := computeSum(target)
			target.Close()

			// compare computed digest to expected sum
			if digest != sum {
				mismatches++
				if !flagStatus {
					fmt.Printf("%s: FAILED\n", name)
				}
				exitCode = 1
			} else if !flagQuiet && !flagStatus {
				fmt.Printf("%s: OK\n", name)
			}

			if err == io.EOF {
				break
			}
		}

		file.Close()

		if flagCheck && flagIgnoreMissing && verifiedCount == 0 {
			fmt.Fprintf(os.Stderr, "md5sum: %s: no file was verified\n", f)
			exitCode = 1
			continue
		}

		totalValid += validLines
		if validLines == 0 {
			fmt.Fprintf(os.Stderr, "md5sum: %s: no properly formatted checksum lines found\n", f)
			if flagCheck {
				exitCode = 1
			}
			continue
		}
	}
	if !flagStatus {
		// report summary warnings
		if malformed > 0 && (flagWarn || flagStrict || totalValid > 0) {
			if malformed == 1 {
				fmt.Fprintln(os.Stderr, "md5sum: WARNING: 1 line is improperly formatted")
			} else {
				fmt.Fprintf(os.Stderr, "md5sum: WARNING: %d lines are improperly formatted\n", malformed)
			}
		}
		if mismatches > 0 {
			if mismatches == 1 {
				fmt.Fprintln(os.Stderr, "md5sum: WARNING: 1 computed checksum did NOT match")
			} else {
				fmt.Fprintf(os.Stderr, "md5sum: WARNING: %d computed checksums did NOT match\n", mismatches)
			}
		}
	}

	return exitCode
}

func main() {
	flag.Usage = usage
	flag.Parse()
	files := flag.Args()

	// --ignore-missing only makes sense when verifying
	if flagIgnoreMissing && !flagCheck {
		fmt.Fprintln(os.Stderr,
			"md5sum: the --ignore-missing option is meaningful only when verifying checksums")
		fmt.Fprintln(os.Stderr, "Try 'md5sum --help' for more information.")
		os.Exit(1)
	}

	if flagHelp {
		usage()
		os.Exit(0)
	}
	if flagVersion {
		fmt.Println("md5sum (go) 1.0")
		os.Exit(0)
	}

	if flagCheck {
		if len(files) == 0 {
			fmt.Fprintln(os.Stderr, "md5sum: no checksum files specified")
			os.Exit(1)
		}
		os.Exit(checkFiles(files))
	}

	// when no files given, read from stdin
	if len(files) == 0 {
		sum, err := computeSum(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error computing sum: %v\n", err)
			os.Exit(1)
		}
		printSum(sum, "-")
		return
	}

	// compute and print sums for each named file
	for _, name := range files {
		f, err := os.Open(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
			continue
		}
		sum, err := computeSum(f)
		f.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
			continue
		}
		printSum(sum, name)
	}
}
