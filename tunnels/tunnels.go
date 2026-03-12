package tunnels

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/ssh"
)

type TunnelType string

const (
	LocalForward   TunnelType = "local"
	RemoteForward  TunnelType = "remote"
	DynamicForward TunnelType = "dynamic"
)

type Tunnel struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	HostKey       string     `json:"host_key"`
	LocalAddress  string     `json:"local_address"`
	RemoteAddress string     `json:"remote_address"`
	Type          TunnelType `json:"type"`
	Active        bool       `json:"-"`
	cancel        context.CancelFunc
}

type Manager struct {
	Tunnels map[string]*Tunnel
	mu      sync.Mutex
	active  map[string]*ssh.Client
}

func NewManager() *Manager {
	return &Manager{
		Tunnels: make(map[string]*Tunnel),
		active:  make(map[string]*ssh.Client),
	}
}

func GetConfigPath() string {
	usrHome, _ := os.UserHomeDir()
	return filepath.Join(usrHome, ".ssh", "aura_tunnels.json")
}

func (m *Manager) Load() error {
	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	return json.Unmarshal(data, &m.Tunnels)
}

func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := json.MarshalIndent(m.Tunnels, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(GetConfigPath(), data, 0644)
}

func (m *Manager) Add(t *Tunnel) {
	m.mu.Lock()
	m.Tunnels[t.ID] = t
	m.mu.Unlock()
}

func (m *Manager) Remove(id string) {
	m.mu.Lock()
	delete(m.Tunnels, id)
	m.mu.Unlock()
}

func (m *Manager) Start(t *Tunnel, sshConfig *ssh.ClientConfig, hostname string) error {
	client, err := ssh.Dial("tcp", hostname, sshConfig)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.active[t.ID] = client
	m.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel
	t.Active = true

	go func() {
		listener, err := net.Listen("tcp", t.LocalAddress)
		if err != nil {
			fmt.Printf("Tunnel error: %v\n", err)
			return
		}
		defer listener.Close()

		go func() {
			<-ctx.Done()
			listener.Close()
		}()

		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					continue
				}
			}

			go m.handleForward(ctx, client, conn, t.RemoteAddress)
		}
	}()

	return nil
}

func (m *Manager) handleForward(ctx context.Context, client *ssh.Client, localConn net.Conn, remoteAddr string) {
	defer localConn.Close()

	remoteConn, err := client.Dial("tcp", remoteAddr)
	if err != nil {
		return
	}
	defer remoteConn.Close()

	done := make(chan struct{}, 2)

	go func() {
		io.Copy(localConn, remoteConn)
		done <- struct{}{}
	}()

	go func() {
		io.Copy(remoteConn, localConn)
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}
}

func (m *Manager) Stop(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if t, ok := m.Tunnels[id]; ok && t.cancel != nil {
		t.cancel()
		t.Active = false
	}

	if client, ok := m.active[id]; ok {
		client.Close()
		delete(m.active, id)
	}
}
