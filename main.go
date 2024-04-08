package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"
	"time"
)

const tmpl = `
I am {{ .Account }}. and this is proof that I have control of  this account as of {{ .Date }}

My PGP key is {{ .Fingerprint }} and my email is {{ .Email }}
`

var (
	FINGERPRINT = regexp.MustCompile(`^fpr:::::::::([A-F0-9]+):`)
	EMAIL       = regexp.MustCompile(`<(.+?)>`)
)

// Define a struct to hold template data
type TemplateData struct {
	Account     string
	Date        string
	Fingerprint string
	Email       string
}

func main() {
	key := flag.String("k", "", "specify signing key (optional)")
	getHeight := flag.Bool("b", false, "get the default signing key")

	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		fmt.Println("Account name is required")
		os.Exit(1)
	}

	accountName := args[0]

	fingerprint, email := pgpInfo(*key)

	tmpl, err := template.New("proof").Parse(tmpl)
	if err != nil {
		panic(err)
	}

	data := TemplateData{
		Account:     accountName,
		Date:        time.Now().UTC().Format(time.RFC3339),
		Fingerprint: fingerprint,
		Email:       email,
	}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to execute template: %v\n", err)
		os.Exit(1)
	}

	if *getHeight {
		height, err := getBTCHeight()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get BTC height: %v\n", err)
			os.Exit(1)
		}

		tpl.WriteString(fmt.Sprintf("\nCurrent BTC blockchain height: %d\n", height))
	}

	sign(tpl.String(), *key)
}

func sign(message, key string) {
	cmdStr := "gpg --clearsign"
	if key != "" {
		cmdStr += " --default-key " + key
	}

	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Stdin = bytes.NewBufferString(message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to sign message: %v\n", err)
		os.Exit(1)
	}
}

// pgpInfo retrieves the PGP fingerprint and the associated email address.
// If the key parameter is empty, it tries to find the default key.
func pgpInfo(key string) (string, string) {
	// Construct the gpg command to list keys. If a key is specified, use it; otherwise, list all.
	cmdStr := "gpg --with-colons --fingerprint"
	if key != "" {
		cmdStr += " " + key
	}

	cmd := exec.Command("sh", "-c", cmdStr)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error running gpg command: %v\n", err)
		return "", ""
	}

	return parse(out.String())
}

// TODO fugly
func parse(output string) (fingerprint, email string) {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if matches := FINGERPRINT.FindStringSubmatch(line); matches != nil && len(matches) > 1 {
			fingerprint = matches[1]
			continue
		}
		if matches := EMAIL.FindStringSubmatch(line); matches != nil && len(matches) > 1 {
			email = matches[1]
			break
		}
	}

	return fingerprint, email
}
