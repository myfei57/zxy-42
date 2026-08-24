package ns

import "time"

// Seed fills the namespace with a default hospital, department and two wards
// so a fresh deployment has somewhere to place devices.
func Seed(r *WardRegistry) error {
	if len(r.Wards()) > 0 {
		return nil
	}
	at := time.Now()
	hospital := NewHospital("第一人民医院", "H001")
	if err := r.AddHospital(hospital); err != nil {
		return err
	}
	department := NewDepartment(hospital.ID, "重症医学科", "D001")
	if err := r.AddDepartment(department); err != nil {
		return err
	}
	icuWard, err := NewWard(department.ID, "ICU 病区", "W001", "icu", 16, at)
	if err != nil {
		return err
	}
	if err := r.AddWard(icuWard); err != nil {
		return err
	}
	generalWard, err := NewWard(department.ID, "普通病区", "W002", "standard", 48, at)
	if err != nil {
		return err
	}
	return r.AddWard(generalWard)
}
