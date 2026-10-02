// Package sshrun dials managed hosts and runs commands / uploads files over SSH.
package sshrun

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type Target struct {
	Host       string
	Port       int
	User       string
	Method     string // key | password
	KeyPEM     string
	Password   string
	Passphrase string
}

type Client struct {
	client *ssh.Client
}

func Dial(target Target, timeout time.Duration) (*Client, error) {
	if target.Port <= 0 {
		target.Port = 22
	}
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	auth, err := authMethods(target)
	if err != nil {
		return nil, err
	}

	cfg := &ssh.ClientConfig{
		User:            target.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}

	addr := net.JoinHostPort(target.Host, strconv.Itoa(target.Port))
	conn, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		return nil, fmt.Errorf("ssh dial %s: %w", addr, err)
	}
	return &Client{client: conn}, nil
}

func (c *Client) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}

func (c *Client) Run(cmd string) (stdout, stderr string, err error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", "", err
	}
	defer session.Close()

	var outBuf, errBuf bytes.Buffer
	session.Stdout = &outBuf
	session.Stderr = &errBuf
	err = session.Run(cmd)
	return outBuf.String(), errBuf.String(), err
}

func (c *Client) Upload(remotePath, mode string, r io.Reader) error {
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		return err
	}

	cmd := fmt.Sprintf("cat > %s && chmod %s %s", shellQuote(remotePath), shellQuote(mode), shellQuote(remotePath))
	if err := session.Start(cmd); err != nil {
		_ = stdin.Close()
		return err
	}

	if _, err := io.Copy(stdin, r); err != nil {
		_ = stdin.Close()
		_ = session.Wait()
		return err
	}
	if err := stdin.Close(); err != nil {
		_ = session.Wait()
		return err
	}
	return session.Wait()
}

func (c *Client) UploadBytes(remotePath, mode string, data []byte) error {
	return c.Upload(remotePath, mode, bytes.NewReader(data))
}

func authMethods(target Target) ([]ssh.AuthMethod, error) {
	switch target.Method {
	case "password":
		if target.Password == "" {
			return nil, fmt.Errorf("password is required")
		}
		return []ssh.AuthMethod{ssh.Password(target.Password)}, nil
	case "key":
		if strings.TrimSpace(target.KeyPEM) == "" {
			return nil, fmt.Errorf("private key is required")
		}
		var signer ssh.Signer
		var err error
		if target.Passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(target.KeyPEM), []byte(target.Passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(target.KeyPEM))
		}
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		return []ssh.AuthMethod{ssh.PublicKeys(signer)}, nil
	default:
		return nil, fmt.Errorf("unsupported auth method %q", target.Method)
	}
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}
