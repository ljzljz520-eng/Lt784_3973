package org

import (
	"sort"
	"strings"

	"example.com/permission-selector/internal/domain"
)

type NodeInsight struct {
	NodeID          string `json:"node_id"`
	Name            string `json:"name"`
	DescendantCount int    `json:"descendant_count"`
	AccountCount    int    `json:"account_count"`
	ActiveAccounts  int    `json:"active_accounts"`
	PathLabel       string `json:"path_label"`
}

func (s *Service) Insight(id string) (NodeInsight, error) {
	node, err := s.store.FindNode(id)
	if err != nil {
		return NodeInsight{}, err
	}
	descendants, err := s.Descendants(id, true)
	if err != nil {
		return NodeInsight{}, err
	}
	allowed := map[string]bool{}
	for _, descendant := range descendants {
		allowed[descendant.ID] = true
	}
	accounts, err := s.store.ListAccounts()
	if err != nil {
		return NodeInsight{}, err
	}
	insight := NodeInsight{NodeID: node.ID, Name: node.Name, DescendantCount: len(descendants), PathLabel: strings.ReplaceAll(node.Path, "/", " / ")}
	for _, account := range accounts {
		if allowed[account.NodeID] {
			insight.AccountCount++
			if account.Active {
				insight.ActiveAccounts++
			}
		}
	}
	return insight, nil
}

func (s *Service) Insights() ([]NodeInsight, error) {
	nodes, err := s.Tree()
	if err != nil {
		return nil, err
	}
	result := make([]NodeInsight, 0, len(nodes))
	for _, node := range nodes {
		insight, insightErr := s.Insight(node.ID)
		if insightErr != nil {
			return nil, insightErr
		}
		result = append(result, insight)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].PathLabel < result[j].PathLabel })
	return result, nil
}

func (s *Service) ResolveSelection(object domain.AuthorizationObject) (string, error) {
	if err := domain.ValidateSelection(object); err != nil {
		return "", err
	}
	if object.Type == domain.ObjectTypePerson {
		account, err := s.store.FindAccount(object.RefID)
		if err != nil {
			return "", err
		}
		return account.Display, nil
	}
	node, err := s.store.FindNode(object.RefID)
	if err != nil {
		return "", err
	}
	return node.Path, nil
}

func (s *Service) ActiveNodeIDs() (map[string]bool, error) {
	nodes, err := s.store.ListNodes()
	if err != nil {
		return nil, err
	}
	active := make(map[string]bool, len(nodes))
	for _, node := range nodes {
		if node.Active {
			active[node.ID] = true
		}
	}
	return active, nil
}
