package store

import (
	"fmt"
	"sort"

	"example.com/permission-selector/internal/domain"
)

func (s *Store) SaveNode(node domain.OrganizationNode) error {
	if err := domain.ValidateOrganizationNode(node); err != nil {
		return err
	}
	return s.put(BucketNodes, node.ID, node)
}

func (s *Store) FindNode(id string) (domain.OrganizationNode, error) {
	var node domain.OrganizationNode
	err := s.get(BucketNodes, id, &node)
	return node, err
}

func (s *Store) ListNodes() ([]domain.OrganizationNode, error) {
	return listValues(s, BucketNodes, func(data []byte) (domain.OrganizationNode, error) {
		var node domain.OrganizationNode
		if err := decode(data, &node); err != nil {
			return node, err
		}
		return node, nil
	})
}

func (s *Store) SaveAccount(account domain.Account) error {
	if err := domain.ValidateAccount(account); err != nil {
		return err
	}
	return s.put(BucketAccounts, account.ID, account)
}

func (s *Store) FindAccount(id string) (domain.Account, error) {
	var account domain.Account
	err := s.get(BucketAccounts, id, &account)
	return account, err
}

func (s *Store) ListAccounts() ([]domain.Account, error) {
	return listValues(s, BucketAccounts, func(data []byte) (domain.Account, error) {
		var account domain.Account
		if err := decode(data, &account); err != nil {
			return account, err
		}
		return account, nil
	})
}

func (s *Store) SaveRequest(request domain.AuthorizationRequest) error {
	if err := domain.ValidateRequest(request); err != nil {
		return err
	}
	return s.put(BucketRequests, request.ID, request)
}

func (s *Store) FindRequest(id string) (domain.AuthorizationRequest, error) {
	var request domain.AuthorizationRequest
	err := s.get(BucketRequests, id, &request)
	return request, err
}

func (s *Store) ListRequests() ([]domain.AuthorizationRequest, error) {
	return listValues(s, BucketRequests, func(data []byte) (domain.AuthorizationRequest, error) {
		var request domain.AuthorizationRequest
		if err := decode(data, &request); err != nil {
			return request, err
		}
		return request, nil
	})
}

func (s *Store) SaveObject(object domain.AuthorizationObject) error {
	if err := domain.ValidateSelection(object); err != nil {
		return err
	}
	return s.put(BucketObjects, object.ID, object)
}

func (s *Store) FindObject(id string) (domain.AuthorizationObject, error) {
	var object domain.AuthorizationObject
	err := s.get(BucketObjects, id, &object)
	return object, err
}

func (s *Store) ListObjects(requestID string) ([]domain.AuthorizationObject, error) {
	values, err := listValues(s, BucketObjects, func(data []byte) (domain.AuthorizationObject, error) {
		var object domain.AuthorizationObject
		if err := decode(data, &object); err != nil {
			return object, err
		}
		return object, nil
	})
	if err != nil {
		return nil, err
	}
	filtered := values[:0]
	for _, object := range values {
		if requestID == "" || object.RequestID == requestID {
			filtered = append(filtered, object)
		}
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].ID < filtered[j].ID })
	return filtered, nil
}

func (s *Store) DeleteObject(id string) error {
	if id == "" {
		return fmt.Errorf("%w: object id is required", domain.ErrInvalidInput)
	}
	return s.remove(BucketObjects, id)
}
