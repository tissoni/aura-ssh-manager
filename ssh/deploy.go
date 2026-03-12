package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/trntv/sshed/host"
	"golang.org/x/crypto/ssh"
)

func GenerateKeypair(path string) (string, string, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}

	// Generate private key in PEM format
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return "", "", err
	}
	privPem := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	})

	// Generate public key in OpenSSH format
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return "", "", err
	}
	pubBytes := ssh.MarshalAuthorizedKey(sshPub)

	// Save to files
	err = os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		return "", "", err
	}

	err = os.WriteFile(path, privPem, 0600)
	if err != nil {
		return "", "", err
	}

	err = os.WriteFile(path+".pub", pubBytes, 0644)
	if err != nil {
		return "", "", err
	}

	return string(privPem), string(pubBytes), nil
}

func DeployPublicKey(srv *host.Host, pubKey string) error {
	// Construct the command to append the key and fix permissions
	// We use a single SSH command for efficiency
	remoteCmd := fmt.Sprintf(`mkdir -p ~/.ssh && chmod 700 ~/.ssh && echo "%s" >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys`, strings.TrimSpace(pubKey))

	// We'll use the existing SSH logic but for command execution
	var username string
	if srv.User == "" {
		u, _ := os.UserHomeDir()
		username = filepath.Base(u)
	} else {
		username = srv.User
	}

	args := []string{"-p", srv.Port}
	if srv.Port == "" {
		args[1] = "22"
	}
	
	if srv.IdentityFile != "" {
		idFile := strings.TrimSuffix(ConvertTilde(srv.IdentityFile), ".pub")
		args = append(args, "-i", idFile)
	}

	remoteAddr := fmt.Sprintf("%s@%s", username, srv.Hostname)
	args = append(args, remoteAddr, remoteCmd)

	cmd := exec.Command("ssh", args...)
	
	// We might need to handle password prompts here too, but for simplicity we assume key-based or prompt in terminal
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
func RunRemoteCommand(srv *host.Host, remoteCmd string) (string, error) {
	var username string
	if srv.User == "" {
		u, _ := os.UserHomeDir()
		username = filepath.Base(u)
	} else {
		username = srv.User
	}

	args := []string{"-p", srv.Port}
	if srv.Port == "" {
		args[1] = "22"
	}
	args = append(args, "-o", "BatchMode=yes", "-o", "ConnectTimeout=5") // Non-interactive mode for diagnostics

	if srv.IdentityFile != "" {
		idFile := strings.TrimSuffix(ConvertTilde(srv.IdentityFile), ".pub")
		args = append(args, "-i", idFile)
	}

	remoteAddr := fmt.Sprintf("%s@%s", username, srv.Hostname)
	args = append(args, remoteAddr, remoteCmd)

	out, err := exec.Command("ssh", args...).CombinedOutput()
	return string(out), err
}
