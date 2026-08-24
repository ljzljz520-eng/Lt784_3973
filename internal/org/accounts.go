package org

import (
	"fmt"
	"strings"
	"time"

	"example.com/permission-selector/internal/domain"
)

func (s *Service) AddAccount(id, nodeID, username, display, email string) (domain.Account, error) {
	if strings.TrimSpace(display) == "" {
		display = username
	}
	node, err := s.store.FindNode(nodeID)
	if err != nil {
		return domain.Account{}, err
	}
	if !node.Active {
		return domain.Account{}, domain.ErrInactiveRecord
	}
	now := time.Now().UTC()
	account := domain.Account{ID: id, NodeID: nodeID, Username: username, Display: display, Email: email, Active: true, CreatedAt: now, UpdatedAt: now}
	if err := s.store.SaveAccount(account); err != nil {
		return domain.Account{}, err
	}
	return account, nil
}

func (s *Service) DisableAccount(id string) error {
	account, err := s.store.FindAccount(id)
	if err != nil {
		return err
	}
	account.Active = false
	account.UpdatedAt = time.Now().UTC()
	return s.store.SaveAccount(account)
}

func (s *Service) AccountsForNode(filter domain.AccountFilter) ([]domain.Account, error) {
	filter = domain.NormalizeFilter(filter)
	if _, err := s.store.FindNode(filter.NodeID); err != nil {
		return nil, fmt.Errorf("find selected node: %w", err)
	}
	nodes, err := s.Descendants(filter.NodeID, true)
	if err != nil {
		return nil, err
	}
	allowed := make(map[string]bool, len(nodes))
	for _, node := range nodes {
		allowed[node.ID] = true
	}
	accounts, err := s.store.ListAccounts()
	if err != nil {
		return nil, err
	}
	result := make([]domain.Account, 0, len(accounts))
	for _, account := range accounts {
		if !allowed[account.NodeID] {
			continue
		}
		if filter.OnlyActive && !account.Active {
			continue
		}
		if filter.Query != "" && !strings.Contains(strings.ToLower(account.Display), strings.ToLower(filter.Query)) {
			continue
		}
		result = append(result, account)
	}
	domain.SortAccounts(result)
	return result, nil
}
