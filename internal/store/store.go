package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrNotFound = errors.New("store: record not found")
	ErrEscape   = errors.New("store: relative path escapes root")
)

// FileStore is a small file-backed persistence layer. Every write goes
// through a temporary file and a rename so readers never observe a
// half-written record.
type FileStore struct {
	Root string
}

func NewFileStore(root string) (*FileStore, error) {
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("store: empty root")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("store: create root: %w", err)
	}
	return &FileStore{Root: root}, nil
}

func (s *FileStore) resolve(rel string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(rel))
	if clean == "." || filepath.IsAbs(clean) {
		return "", ErrEscape
	}
	target := filepath.Join(s.Root, clean)
	relative, err := filepath.Rel(s.Root, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", ErrEscape
	}
	return target, nil
}

func (s *FileStore) Exists(rel string) bool {
	target, err := s.resolve(rel)
	if err != nil {
		return false
	}
	_, statErr := os.Stat(target)
	return statErr == nil
}

func (s *FileStore) WriteJSON(rel string, value interface{}) error {
	target, err := s.resolve(rel)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("store: mkdir: %w", err)
	}
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("store: encode: %w", err)
	}
	return atomicWrite(target, payload)
}

func (s *FileStore) ReadJSON(rel string, value interface{}) error {
	target, err := s.resolve(rel)
	if err != nil {
		return err
	}
	payload, err := os.ReadFile(target)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return err
	}
	if err := json.Unmarshal(payload, value); err != nil {
		return fmt.Errorf("store: decode %s: %w", rel, err)
	}
	return nil
}

func (s *FileStore) ReadFile(rel string) ([]byte, error) {
	target, err := s.resolve(rel)
	if err != nil {
		return nil, err
	}
	payload, err := os.ReadFile(target)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return payload, nil
}

func (s *FileStore) AppendLine(rel string, line []byte) error {
	target, err := s.resolve(rel)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("store: mkdir: %w", err)
	}
	handle, err := os.OpenFile(target, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer handle.Close()
	if _, err := handle.Write(append(line, '\n')); err != nil {
		return err
	}
	return nil
}

func (s *FileStore) ReadLines(rel string) ([]string, error) {
	target, err := s.resolve(rel)
	if err != nil {
		return nil, err
	}
	payload, err := os.ReadFile(target)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	lines := strings.Split(strings.TrimRight(string(payload), "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, nil
	}
	return lines, nil
}

func (s *FileStore) List(rel string) ([]string, error) {
	target, err := s.resolve(rel)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}

func (s *FileStore) Remove(rel string) error {
	target, err := s.resolve(rel)
	if err != nil {
		return err
	}
	if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func atomicWrite(target string, payload []byte) error {
	tmp := target + ".tmp"
	if err := os.WriteFile(tmp, payload, 0o644); err != nil {
		return fmt.Errorf("store: write temp: %w", err)
	}
	if err := os.Rename(tmp, target); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("store: commit %s: %w", filepath.Base(target), err)
	}
	return nil
}
