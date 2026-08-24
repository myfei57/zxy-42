package ns

import "medops/internal/store"

// Snapshot is the serializable form of the whole namespace tree.
type Snapshot struct {
	Hospitals   []Hospital   `json:"hospitals"`
	Departments []Department `json:"departments"`
	Wards       []Ward       `json:"wards"`
}

func (r *WardRegistry) Snapshot() Snapshot {
	return Snapshot{
		Hospitals:   append([]Hospital(nil), r.hospitals...),
		Departments: append([]Department(nil), r.departments...),
		Wards:       append([]Ward(nil), r.wards...),
	}
}

func (r *WardRegistry) Restore(s Snapshot) error {
	r.hospitals = nil
	r.departments = nil
	r.wards = nil
	r.hospitals = append(r.hospitals, s.Hospitals...)
	r.departments = append(r.departments, s.Departments...)
	r.wards = append(r.wards, s.Wards...)
	return nil
}

// Persist writes the namespace snapshot through the file store.
func (r *WardRegistry) Persist(fs *store.FileStore) error {
	return fs.WriteJSON("ns/snapshot.json", r.Snapshot())
}

// Load restores the namespace from the file store, seeding defaults when no
// snapshot exists yet.
func (r *WardRegistry) Load(fs *store.FileStore) error {
	var snap Snapshot
	if err := fs.ReadJSON("ns/snapshot.json", &snap); err != nil {
		if err := Seed(r); err != nil {
			return err
		}
		return r.Persist(fs)
	}
	return r.Restore(snap)
}
