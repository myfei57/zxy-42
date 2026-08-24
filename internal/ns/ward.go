package ns

// WardRegistry is the in-memory namespace tree used to resolve ward and
// department identities for device placement and QC dispatch.
type WardRegistry struct {
	hospitals   []Hospital
	departments []Department
	wards       []Ward
}

func NewWardRegistry() *WardRegistry {
	return &WardRegistry{}
}

func (r *WardRegistry) AddHospital(h Hospital) error {
	for _, existing := range r.hospitals {
		if existing.Code == h.Code {
			return ErrDuplicateCode
		}
	}
	r.hospitals = append(r.hospitals, h)
	return nil
}

func (r *WardRegistry) AddDepartment(d Department) error {
	if !r.hasHospital(d.HospitalID) {
		return ErrHospitalNotFound
	}
	for _, existing := range r.departments {
		if existing.HospitalID == d.HospitalID && existing.Code == d.Code {
			return ErrDuplicateCode
		}
	}
	r.departments = append(r.departments, d)
	return nil
}

func (r *WardRegistry) AddWard(w Ward) error {
	if !r.hasDepartment(w.DepartmentID) {
		return ErrDepartmentNotFound
	}
	for _, existing := range r.wards {
		if existing.Code == w.Code {
			return ErrDuplicateCode
		}
	}
	r.wards = append(r.wards, w)
	return nil
}

func (r *WardRegistry) Ward(id string) (Ward, bool) {
	for _, ward := range r.wards {
		if ward.ID == id {
			return ward, true
		}
	}
	return Ward{}, false
}

func (r *WardRegistry) WardByCode(code string) (Ward, bool) {
	for _, ward := range r.wards {
		if ward.Code == code {
			return ward, true
		}
	}
	return Ward{}, false
}

func (r *WardRegistry) Wards() []Ward {
	return append([]Ward(nil), r.wards...)
}

func (r *WardRegistry) Department(id string) (Department, bool) {
	for _, department := range r.departments {
		if department.ID == id {
			return department, true
		}
	}
	return Department{}, false
}

func (r *WardRegistry) Hospital(id string) (Hospital, bool) {
	for _, hospital := range r.hospitals {
		if hospital.ID == id {
			return hospital, true
		}
	}
	return Hospital{}, false
}

func (r *WardRegistry) hasHospital(id string) bool {
	_, ok := r.Hospital(id)
	return ok
}

func (r *WardRegistry) hasDepartment(id string) bool {
	_, ok := r.Department(id)
	return ok
}
