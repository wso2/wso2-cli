package repo

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"golang.org/x/crypto/ssh"
)

// Please note that this method has some limitations:

//   - It assumes that the git command-line tool is installed
//     and properly configured on the system where the Go program is running.
//   - It assumes that the git command-line tool is configured to use a
//     credential helper that supports the fill operation.
//   - It assumes that the credential helper is configured to store credentials for the given URL.
func getGitCredentials(url string) (username string, password string, err error) {
	cmd := exec.Command("git", "credential", "fill")
	cmd.Stdin = strings.NewReader("url=" + url + "\n")
	output, err := cmd.Output()
	if err != nil {
		return "", "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "username=") {
			username = strings.TrimPrefix(line, "username=")
		} else if strings.HasPrefix(line, "password=") {
			password = strings.TrimPrefix(line, "password=")
		}
	}

	return username, password, nil
}

func getGitPublicKeys() (gitssh.AuthMethod, error) {
	privateKey, err := os.ReadFile(os.ExpandEnv("$HOME/.ssh/id_rsa"))
	if err != nil {
		return nil, err
	}
	publicKeys, err := gitssh.NewPublicKeys("git", privateKey, "")
	if err != nil {
		return nil, err
	}

	hostKey, err := getHostKey("github.com")
	if err != nil {
		return nil, err
	}

	publicKeys.HostKeyCallback = ssh.FixedHostKey(hostKey)
	return publicKeys, nil
}

func getHostKey(host string) (ssh.PublicKey, error) {
	file, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var hostKey ssh.PublicKey
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		if len(fields) != 3 {
			continue
		}

		if strings.Contains(fields[0], host) {
			var err error
			hostKey, _, _, _, err = ssh.ParseAuthorizedKey(scanner.Bytes())
			if err != nil {
				return nil, err
			}
			break
		}
	}

	if hostKey == nil {
		return nil, errors.New("no hostkey found for " + host)
	}

	return hostKey, nil
}
