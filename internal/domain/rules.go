package domain

import (
	"fmt"
	"strings"
)

type RequestCommand struct {
	RequestID string `json:"request_id"`
	Actor     string `json:"actor"`
	Title     string `json:"title"`
	Reason    string `json:"reason"`
}

type SelectionCommand struct {
	RequestID string     `json:"request_id"`
	Actor     string     `json:"actor"`
	RefID     string     `json:"ref_id"`
	Type      ObjectType `json:"type"`
	Name      string     `json:"name"`
}

type DispatchCommand struct {
	RequestID string `json:"request_id"`
	Actor     string `json:"actor"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
}

type WorkflowReceipt struct {
	Request      AuthorizationRequest `json:"request"`
	Summary      SelectionSummary     `json:"summary"`
	Dispatch     DispatchRecord       `json:"dispatch"`
	SnapshotID   string               `json:"snapshot_id"`
	AuditEntries int                  `json:"audit_entries"`
}

func (command RequestCommand) Validate() error {
	if strings.TrimSpace(command.RequestID) == "" || strings.TrimSpace(command.Actor) == "" {
		return fmt.Errorf("%w: request id and actor are required", ErrInvalidInput)
	}
	if len(strings.TrimSpace(command.Title)) < 3 {
		return fmt.Errorf("%w: title must contain at least three characters", ErrInvalidInput)
	}
	if len(strings.TrimSpace(command.Reason)) < 5 {
		return fmt.Errorf("%w: reason must contain at least five characters", ErrInvalidInput)
	}
	return nil
}

func (command SelectionCommand) Validate() error {
	if strings.TrimSpace(command.RequestID) == "" || strings.TrimSpace(command.Actor) == "" {
		return fmt.Errorf("%w: request id and actor are required", ErrInvalidInput)
	}
	if strings.TrimSpace(command.RefID) == "" || strings.TrimSpace(command.Name) == "" {
		return fmt.Errorf("%w: selection reference and name are required", ErrInvalidInput)
	}
	if command.Type != ObjectTypeDepartment && command.Type != ObjectTypePerson {
		return fmt.Errorf("%w: selection type is invalid", ErrInvalidInput)
	}
	return nil
}

func (command DispatchCommand) Validate() error {
	if strings.TrimSpace(command.RequestID) == "" || strings.TrimSpace(command.Actor) == "" {
		return fmt.Errorf("%w: request id and actor are required", ErrInvalidInput)
	}
	if command.Page < 1 || command.PageSize < 1 || command.PageSize > 100 {
		return fmt.Errorf("%w: page must be positive and page size must be 1..100", ErrInvalidInput)
	}
	return nil
}

func CanSelect(status Status) bool {
	return status == StatusDraft || status == StatusRejected
}

func CanConfirm(status Status, count int) error {
	if status != StatusDraft {
		return fmt.Errorf("%w: only draft requests can be confirmed", ErrConflict)
	}
	if count == 0 {
		return fmt.Errorf("%w: at least one object is required", ErrInvalidInput)
	}
	return nil
}

func CanDispatch(status Status) error {
	if status != StatusConfirmed && status != StatusDispatched {
		return fmt.Errorf("%w: request must be confirmed before dispatch", ErrConflict)
	}
	return nil
}

func CanPublish(status Status) error {
	if status != StatusDispatched {
		return fmt.Errorf("%w: request must be dispatched before publishing", ErrConflict)
	}
	return nil
}
