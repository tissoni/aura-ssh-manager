package snippets

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Snippet struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

type Snippets struct {
	Items map[string]Snippet `json:"items"`
}

func GetPath() string {
	usrHome, _ := os.UserHomeDir()
	return filepath.Join(usrHome, ".ssh", "aura_snippets.json")
}

func Load() *Snippets {
	path := GetPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return Default()
	}

	s := &Snippets{}
	err = json.Unmarshal(data, s)
	if err != nil {
		return Default()
	}
	return s
}

func Default() *Snippets {
	return &Snippets{
		Items: map[string]Snippet{
			"docker-ps": {
				Name:    "docker-ps",
				Command: "docker ps --format \"table {{.Names}}\\t{{.Status}}\\t{{.Image}}\"",
			},
			"k8s-pods": {
				Name:    "k8s-pods",
				Command: "kubectl get pods -A",
			},
			"sys-status": {
				Name:    "sys-status",
				Command: "systemctl status",
			},
			"nginx-reload": {
				Name:    "nginx-reload",
				Command: "sudo nginx -t && sudo systemctl reload nginx",
			},
		},
	}
}

func (s *Snippets) Save() error {
	path := GetPath()
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *Snippets) Add(name, command string) {
	if s.Items == nil {
		s.Items = make(map[string]Snippet)
	}
	s.Items[name] = Snippet{Name: name, Command: command}
}

func (s *Snippets) Remove(name string) {
	delete(s.Items, name)
}
