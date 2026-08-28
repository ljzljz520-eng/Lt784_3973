package store

import (
	"fmt"

	"example.com/permission-selector/internal/domain"
)

type RequestBundle struct {
	Request    domain.AuthorizationRequest
	Objects    []domain.AuthorizationObject
	Dispatches []domain.DispatchRecord
	Snapshots  []domain.ResultSnapshot
	Audits     []domain.AuditEvent
}

func (s *Store) LoadRequestBundle(requestID string) (RequestBundle, error) {
	if requestID == "" {
		return RequestBundle{}, domain.ErrInvalidInput
	}
	request, err := s.FindRequest(requestID)
	if err != nil {
		return RequestBundle{}, err
	}
	objects, err := s.ListObjects(requestID)
	if err != nil {
		return RequestBundle{}, err
	}
	dispatches, err := s.ListDispatches(requestID)
	if err != nil {
		return RequestBundle{}, err
	}
	snapshots, err := s.ListSnapshots(requestID)
	if err != nil {
		return RequestBundle{}, err
	}
	audits, err := s.ListAudits(requestID)
	if err != nil {
		return RequestBundle{}, err
	}
	return RequestBundle{Request: request, Objects: objects, Dispatches: dispatches, Snapshots: snapshots, Audits: audits}, nil
}

func (s *Store) EnsureBundleConsistent(bundle RequestBundle) error {
	if bundle.Request.ID == "" {
		return domain.ErrInvalidInput
	}
	for _, object := range bundle.Objects {
		if object.RequestID != bundle.Request.ID {
			return fmt.Errorf("%w: object belongs to another request", domain.ErrConflict)
		}
	}
	for _, dispatch := range bundle.Dispatches {
		if dispatch.RequestID != bundle.Request.ID || dispatch.ObjectCount < 1 {
			return fmt.Errorf("%w: dispatch record is inconsistent", domain.ErrConflict)
		}
	}
	for _, snapshot := range bundle.Snapshots {
		if snapshot.RequestID != bundle.Request.ID || snapshot.DispatchID == "" {
			return fmt.Errorf("%w: snapshot record is inconsistent", domain.ErrConflict)
		}
	}
	return nil
}

func (s *Store) ActiveNodes() ([]domain.OrganizationNode, error) {
	nodes, err := s.ListNodes()
	if err != nil {
		return nil, err
	}
	active := nodes[:0]
	for _, node := range nodes {
		if node.Active {
			active = append(active, node)
		}
	}
	return active, nil
}
