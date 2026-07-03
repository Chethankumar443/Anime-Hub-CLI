package persistence

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type AppState struct {
	LastAnimeID   string `json:"last_anime_id"`
	LastEpisodeID string `json:"last_episode_id"`
	Volume        int    `json:"volume"`
}

type Store struct {
	mu       sync.RWMutex
	filePath string
	state    AppState
}

func NewStore(filePath string) (*Store, error) {
	s := &Store{
		filePath: filePath,
	}
	if err := s.Load(); err != nil {
		if os.IsNotExist(err) {
			s.state = AppState{Volume: 50}
			if err := s.Save(); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return s, nil
}

func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.state)
}

func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(dir, "store-*.tmp")
	if err != nil {
		return err
	}
	defer func() {
		tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()

	if _, err := tmpFile.Write(data); err != nil {
		return err
	}
	if err := tmpFile.Sync(); err != nil {
		return err
	}
	tmpFile.Close()

	return os.Rename(tmpFile.Name(), s.filePath)
}

func (s *Store) GetState() AppState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

func (s *Store) SetState(state AppState) error {
	s.mu.Lock()
	s.state = state
	s.mu.Unlock()
	return s.Save()
}
