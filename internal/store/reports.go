package store

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// ReportStore persists generated QC reports and committed results as files.
type ReportStore struct {
	fs *FileStore
}

func NewReportStore(fs *FileStore) *ReportStore {
	return &ReportStore{fs: fs}
}

// Write persists a report payload and returns its relative path.
func (r *ReportStore) Write(planID, name string, payload interface{}) (string, error) {
	rel := "reports/" + planID + "/" + name + ".json"
	if err := r.fs.WriteJSON(rel, payload); err != nil {
		return "", err
	}
	return rel, nil
}

func (r *ReportStore) Read(planID, name string) ([]byte, error) {
	return r.fs.ReadFile("reports/" + planID + "/" + name + ".json")
}

func (r *ReportStore) List(planID string) ([]string, error) {
	dir := filepath.Join(r.fs.Root, filepath.FromSlash("reports"), planID)
	entries, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, strings.TrimSuffix(filepath.Base(entry), ".json"))
	}
	sort.Strings(names)
	return names, nil
}

func (r *ReportStore) Path(planID, name string) string {
	return fmt.Sprintf("reports/%s/%s.json", planID, name)
}
