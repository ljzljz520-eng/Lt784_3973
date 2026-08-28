package org

import (
	"fmt"
	"strings"
	"time"

	"example.com/permission-selector/internal/domain"
)

func (s *Service) RenameDepartment(id, name string) (domain.OrganizationNode, error) {
	if strings.TrimSpace(name) == "" {
		return domain.OrganizationNode{}, domain.ErrInvalidInput
	}
	node, err := s.store.FindNode(id)
	if err != nil {
		return domain.OrganizationNode{}, err
	}
	node.Name = strings.TrimSpace(name)
	node.UpdatedAt = time.Now().UTC()
	if err := s.store.SaveNode(node); err != nil {
		return domain.OrganizationNode{}, err
	}
	return node, nil
}

func (s *Service) MoveDepartment(id, parentID string) error {
	node, err := s.store.FindNode(id)
	if err != nil {
		return err
	}
	if id == parentID {
		return fmt.Errorf("%w: department cannot contain itself", domain.ErrInvalidInput)
	}
	parent, err := s.store.FindNode(parentID)
	if err != nil {
		return err
	}
	isBelow, err := s.IsDescendant(parentID, id)
	if err != nil {
		return err
	}
	if isBelow {
		return fmt.Errorf("%w: department cannot move below its descendant", domain.ErrConflict)
	}
	oldPath := node.Path
	newPath := parent.Path + "/" + node.ID
	node.ParentID = parent.ID
	node.Depth = parent.Depth + 1
	node.Path = newPath
	node.UpdatedAt = time.Now().UTC()
	if err := s.store.SaveNode(node); err != nil {
		return err
	}
	descendants, err := s.store.ListNodes()
	if err != nil {
		return err
	}
	for _, descendant := range descendants {
		if !strings.HasPrefix(descendant.Path, oldPath+"/") {
			continue
		}
		relative := strings.TrimPrefix(descendant.Path, oldPath)
		descendant.Path = newPath + relative
		descendant.Depth = node.Depth + strings.Count(relative, "/")
		descendant.UpdatedAt = time.Now().UTC()
		if err := s.store.SaveNode(descendant); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) DepartmentMembers(id string, includeInactive bool) ([]domain.Account, error) {
	filter := domain.AccountFilter{NodeID: id, OnlyActive: !includeInactive}
	return s.AccountsForNode(filter)
}

func (s *Service) ValidateTree() error {
	nodes, err := s.store.ListNodes()
	if err != nil {
		return err
	}
	byID := make(map[string]domain.OrganizationNode, len(nodes))
	for _, node := range nodes {
		if _, exists := byID[node.ID]; exists {
			return fmt.Errorf("%w: duplicate node %s", domain.ErrConflict, node.ID)
		}
		byID[node.ID] = node
	}
	for _, node := range nodes {
		if node.ParentID == "" {
			continue
		}
		parent, exists := byID[node.ParentID]
		if !exists || !strings.HasPrefix(node.Path, parent.Path+"/") || node.Depth != parent.Depth+1 {
			return fmt.Errorf("%w: invalid path for %s", domain.ErrConflict, node.ID)
		}
	}
	return nil
}
