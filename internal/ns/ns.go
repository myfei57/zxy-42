package ns

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrWardNotFound       = errors.New("ns: ward not found")
	ErrDepartmentNotFound = errors.New("ns: department not found")
	ErrHospitalNotFound   = errors.New("ns: hospital not found")
	ErrDuplicateCode      = errors.New("ns: duplicate code")
	ErrEmptyCode          = errors.New("ns: code is required")
)

// Hospital is the top level of the namespace tree.
type Hospital struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// Department belongs to one hospital.
type Department struct {
	ID         string `json:"id"`
	HospitalID string `json:"hospital_id"`
	Name       string `json:"name"`
	Code       string `json:"code"`
}

// Ward is a physical ward inside a department. Devices are assigned to a ward
// and QC dispatches follow the current ward mapping.
type Ward struct {
	ID           string    `json:"id"`
	DepartmentID string    `json:"department_id"`
	Name         string    `json:"name"`
	Code         string    `json:"code"`
	Isolation    string    `json:"isolation"`
	Capacity     int       `json:"capacity"`
	CreatedAt    time.Time `json:"created_at"`
}

func NewHospital(name, code string) Hospital {
	return Hospital{ID: uuid.NewString(), Name: name, Code: strings.TrimSpace(code)}
}

func NewDepartment(hospitalID, name, code string) Department {
	return Department{ID: uuid.NewString(), HospitalID: hospitalID, Name: name, Code: strings.TrimSpace(code)}
}

func NewWard(departmentID, name, code, isolation string, capacity int, at time.Time) (Ward, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return Ward{}, ErrEmptyCode
	}
	if capacity <= 0 {
		return Ward{}, errors.New("ns: ward capacity must be positive")
	}
	return Ward{
		ID:           uuid.NewString(),
		DepartmentID: departmentID,
		Name:         name,
		Code:         code,
		Isolation:    isolation,
		Capacity:     capacity,
		CreatedAt:    at,
	}, nil
}
