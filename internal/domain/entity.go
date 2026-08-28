package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

type ObjectType string

const (
	ObjectTypeDepartment ObjectType = "department"
	ObjectTypePerson     ObjectType = "person"
)

type Status string

const (
	StatusDraft      Status = "draft"
	StatusConfirmed  Status = "confirmed"
	StatusDispatched Status = "dispatched"
	StatusPublished  Status = "published"
	StatusRejected   Status = "rejected"
)

type OrganizationNode struct {
	ID        string    `json:"id"`
	ParentID  string    `json:"parent_id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Depth     int       `json:"depth"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Account struct {
	ID        string    `json:"id"`
	NodeID    string    `json:"node_id"`
	Username  string    `json:"username"`
	Display   string    `json:"display"`
	Email     string    `json:"email"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuthorizationObject struct {
	ID         string     `json:"id"`
	RequestID  string     `json:"request_id"`
	RefID      string     `json:"ref_id"`
	Type       ObjectType `json:"type"`
	Name       string     `json:"name"`
	SelectedAt time.Time  `json:"selected_at"`
}

type AuthorizationRequest struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Applicant   string    `json:"applicant"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ConfirmedAt time.Time `json:"confirmed_at"`
}

type AuditEvent struct {
	ID        string    `json:"id"`
	RequestID string    `json:"request_id"`
	Action    string    `json:"action"`
	Actor     string    `json:"actor"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}

type DispatchRecord struct {
	ID          string    `json:"id"`
	RequestID   string    `json:"request_id"`
	Sequence    int       `json:"sequence"`
	ObjectCount int       `json:"object_count"`
	Summary     string    `json:"summary"`
	CreatedAt   time.Time `json:"created_at"`
}

type ResultSnapshot struct {
	ID              string                `json:"id"`
	RequestID       string                `json:"request_id"`
	DispatchID      string                `json:"dispatch_id"`
	Objects         []AuthorizationObject `json:"objects"`
	DepartmentCount int                   `json:"department_count"`
	PersonCount     int                   `json:"person_count"`
	CreatedAt       time.Time             `json:"created_at"`
}

type Page[T any] struct {
	Items       []T  `json:"items"`
	Page        int  `json:"page"`
	PageSize    int  `json:"page_size"`
	Total       int  `json:"total"`
	HasNext     bool `json:"has_next"`
	HasPrevious bool `json:"has_previous"`
}

type SelectionSummary struct {
	Total       int `json:"total"`
	Departments int `json:"departments"`
	People      int `json:"people"`
}

var (
	ErrNotFound       = errors.New("record not found")
	ErrInvalidInput   = errors.New("invalid input")
	ErrConflict       = errors.New("state conflict")
	ErrDuplicate      = errors.New("duplicate selection")
	ErrInvalidPage    = errors.New("invalid page")
	ErrInactiveRecord = errors.New("record is inactive")
)

func ValidateOrganizationNode(node OrganizationNode) error {
	if strings.TrimSpace(node.ID) == "" || strings.TrimSpace(node.Name) == "" {
		return fmt.Errorf("%w: node id and name are required", ErrInvalidInput)
	}
	if node.Depth < 0 {
		return fmt.Errorf("%w: node depth cannot be negative", ErrInvalidInput)
	}
	return nil
}

func ValidateAccount(account Account) error {
	if strings.TrimSpace(account.ID) == "" || strings.TrimSpace(account.NodeID) == "" {
		return fmt.Errorf("%w: account id and node are required", ErrInvalidInput)
	}
	if strings.TrimSpace(account.Username) == "" || !strings.Contains(account.Email, "@") {
		return fmt.Errorf("%w: account identity is incomplete", ErrInvalidInput)
	}
	return nil
}

func ValidateRequest(request AuthorizationRequest) error {
	if strings.TrimSpace(request.ID) == "" || strings.TrimSpace(request.Title) == "" {
		return fmt.Errorf("%w: request id and title are required", ErrInvalidInput)
	}
	if strings.TrimSpace(request.Applicant) == "" {
		return fmt.Errorf("%w: applicant is required", ErrInvalidInput)
	}
	if request.Status == "" {
		return fmt.Errorf("%w: status is required", ErrInvalidInput)
	}
	return nil
}

func ValidateSelection(object AuthorizationObject) error {
	if strings.TrimSpace(object.RequestID) == "" || strings.TrimSpace(object.RefID) == "" {
		return fmt.Errorf("%w: request and reference are required", ErrInvalidInput)
	}
	if object.Type != ObjectTypeDepartment && object.Type != ObjectTypePerson {
		return fmt.Errorf("%w: unsupported object type", ErrInvalidInput)
	}
	if strings.TrimSpace(object.Name) == "" {
		return fmt.Errorf("%w: object name is required", ErrInvalidInput)
	}
	return nil
}

func SortObjects(objects []AuthorizationObject) {
	sort.SliceStable(objects, func(i, j int) bool {
		if objects[i].Type != objects[j].Type {
			return objects[i].Type < objects[j].Type
		}
		return objects[i].Name < objects[j].Name
	})
}

func SummarizeObjects(objects []AuthorizationObject) SelectionSummary {
	result := SelectionSummary{}
	for _, object := range objects {
		result.Total++
		switch object.Type {
		case ObjectTypeDepartment:
			result.Departments++
		case ObjectTypePerson:
			result.People++
		}
	}
	return result
}

func PageBounds(page, pageSize, total int) (int, int, error) {
	if page < 1 || pageSize < 1 {
		return 0, 0, ErrInvalidPage
	}
	if total == 0 {
		return 0, 0, nil
	}
	start := (page - 1) * pageSize
	if start >= total {
		return 0, 0, ErrInvalidPage
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return start, end, nil
}
