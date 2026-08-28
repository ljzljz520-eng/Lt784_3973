package org

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"example.com/permission-selector/internal/domain"
	"example.com/permission-selector/internal/store"
)

type Service struct {
	store *store.Store
}

func NewService(database *store.Store) *Service { return &Service{store: database} }

func (s *Service) AddDepartment(id, parentID, name string) (domain.OrganizationNode, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(name) == "" {
		return domain.OrganizationNode{}, fmt.Errorf("%w: department identity is required", domain.ErrInvalidInput)
	}
	depth := 0
	path := id
	if parentID != "" {
		parent, err := s.store.FindNode(parentID)
		if err != nil {
			return domain.OrganizationNode{}, fmt.Errorf("find parent: %w", err)
		}
		if !parent.Active {
			return domain.OrganizationNode{}, domain.ErrInactiveRecord
		}
		depth = parent.Depth + 1
		path = parent.Path + "/" + id
	}
	now := time.Now().UTC()
	node := domain.OrganizationNode{ID: id, ParentID: parentID, Name: name, Path: path, Depth: depth, Active: true, CreatedAt: now, UpdatedAt: now}
	if err := s.store.SaveNode(node); err != nil {
		return domain.OrganizationNode{}, err
	}
	return node, nil
}

func (s *Service) DeactivateDepartment(id string) error {
	node, err := s.store.FindNode(id)
	if err != nil {
		return err
	}
	node.Active = false
	node.UpdatedAt = time.Now().UTC()
	return s.store.SaveNode(node)
}

func (s *Service) Tree() ([]domain.OrganizationNode, error) {
	nodes, err := s.store.ListNodes()
	if err != nil {
		return nil, err
	}
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Depth == nodes[j].Depth {
			return nodes[i].Name < nodes[j].Name

		}
		return nodes[i].Depth < nodes[j].Depth
	})
	return nodes, nil
}

func (s *Service) Descendants(id string, includeSelf bool) ([]domain.OrganizationNode, error) {
	root, err := s.store.FindNode(id)
	if err != nil {
		return nil, err
	}
	nodes, err := s.store.ListNodes()
	if err != nil {
		return nil, err
	}
	result := make([]domain.OrganizationNode, 0)
	for _, node := range nodes {
		isSelf := node.ID == root.ID
		isChild := strings.HasPrefix(node.Path, root.Path+"/")
		if (isSelf && includeSelf || isChild) && node.Active {
			result = append(result, node)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result, nil
}

func (s *Service) IsDescendant(nodeID, ancestorID string) (bool, error) {
	node, err := s.store.FindNode(nodeID)
	if err != nil {
		return false, err
	}
	ancestor, err := s.store.FindNode(ancestorID)
	if err != nil {
		return false, err
	}
	return node.ID == ancestor.ID || strings.HasPrefix(node.Path, ancestor.Path+"/"), nil
}
