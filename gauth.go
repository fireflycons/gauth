package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/pcarrier/gauth/gauth"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	cfgPath := os.Getenv("GAUTH_CONFIG")
	if cfgPath == "" {
		user, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		cfgPath = path.Join(user.HomeDir, ".config/gauth.csv")
	}

	cfgContent, err := gauth.LoadConfigFile(cfgPath, getPassword)
	if err != nil {
		log.Fatalf("Loading config: %v", err)
	}

	urls, err := gauth.ParseConfig(cfgContent)
	if err != nil {
		log.Fatalf("Decoding configuration file: %v", err)
	}

	_, progress := gauth.IndexNow() // TODO: do this per-code

	if len(os.Args) > 1 {
		// Argument is name of an account. Print current code directly to stdout for use in scripts
		// To account for clock skew, delay till 0.25 sec after code change if within 0.25 sec of code changing
		if progress > 29750 {
			time.Sleep(time.Duration(30250-progress) * time.Millisecond)
		} else if progress < 250 {
			time.Sleep(time.Duration(250-progress) * time.Millisecond)
		}

		for _, url := range urls {
			if url.Account == os.Args[1] {
				_, curr, _, err := gauth.Codes(url)
				if err != nil {
					log.Fatalf("Generating codes for %q: %v", url.Account, err)
				}
				fmt.Printf("%s", curr)
				return
			}
		}

		log.Fatalf("Unknown account %q", os.Args[1])

	} else {
		// Print all accounts with timer bar

		tw := tabwriter.NewWriter(os.Stdout, 0, 8, 1, ' ', 0)
		fmt.Fprintln(tw, "\tprev\tcurr\tnext")
		for _, url := range urls {
			prev, curr, next, err := gauth.Codes(url)
			if err != nil {
				log.Fatalf("Generating codes for %q: %v", url.Account, err)
			}
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", url.Account, prev, curr, next)
		}
		tw.Flush()
		fmt.Printf("[%-29s]\n", strings.Repeat("=", int(progress/1000)))
	}
}

func getPassword() ([]byte, error) {
	fmt.Printf("Encryption password: ")
	defer fmt.Println()
	return terminal.ReadPassword(int(syscall.Stdin))
}
