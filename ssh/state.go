package ssh

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

type State struct {
	LastConnected map[string]time.Time `json:"last_connected"`
	Favorites     map[string]bool      `json:"favorites"`
}

var CurrentState *State

func LoadState() error {
	usr, _ := user.Current()
	stateFile := filepath.Join(usr.HomeDir, ".aura_state.json")
	
	CurrentState = &State{
		LastConnected: make(map[string]time.Time),
		Favorites:     make(map[string]bool),
	}

	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(stateFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, CurrentState)
}

func SaveState() error {
	usr, _ := user.Current()
	stateFile := filepath.Join(usr.HomeDir, ".aura_state.json")

	data, err := json.MarshalIndent(CurrentState, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(stateFile, data, 0644)
}

func RecordConnection(key string) {
	if CurrentState == nil {
		_ = LoadState()
	}
	CurrentState.LastConnected[key] = time.Now()
	_ = SaveState()
}

func ToggleFavorite(key string) {
	if CurrentState == nil {
		_ = LoadState()
	}
	CurrentState.Favorites[key] = !CurrentState.Favorites[key]
	_ = SaveState()
}
