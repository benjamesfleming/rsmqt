package rsmq

import (
	"fmt"
	"io/ioutil"
	"net"

	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	Host       string
	Port       string
	User       string
	AuthType   string // "password" or "key"
	Password   string
	KeyPath    string
	Passphrase string // Optional, for encrypted keys
}

// DialSSH establishes an SSH connection and returns a dialer function
// compatible with redis.Options.Dialer
func DialSSH(cfg SSHConfig) (func(network, addr string) (net.Conn, error), error) {
	var authMethods []ssh.AuthMethod

	if cfg.AuthType == "password" {
		authMethods = append(authMethods, ssh.Password(cfg.Password))
	} else if cfg.AuthType == "key" {
		key, err := ioutil.ReadFile(cfg.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read private key: %v", err)
		}

		var signer ssh.Signer
		if cfg.Passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(cfg.Passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey(key)
		}

		if err != nil {
			// If parsing fails, it might be an encrypted key without a passphrase
			// or just an invalid key. The caller should handle the prompt logic
			// if they suspect a missing passphrase, or we can check the error type
			// but x/crypto/ssh doesn't always make it easy to distinguish.
			// For now, return the error.
			return nil, fmt.Errorf("unable to parse private key: %v", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	sshConfig := &ssh.ClientConfig{
		User: cfg.User,
		Auth: authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Support host key verification?
		Timeout:         0, // Default timeout
	}

	sshAddr := net.JoinHostPort(cfg.Host, cfg.Port)
	client, err := ssh.Dial("tcp", sshAddr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial ssh: %v", err)
	}

	// The dialer function that Redis client will use
	return func(network, addr string) (net.Conn, error) {
		return client.Dial(network, addr)
	}, nil
}
