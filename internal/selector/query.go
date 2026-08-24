package selector

import (
	"sort"
	"strings"

	"example.com/permission-selector/internal/domain"
)

type ResultQuery struct {
	RequestID  string
	Text       string
	Type       domain.ObjectType
	Sort       string
	Descending bool
}

func (s *Service) SearchResult(query ResultQuery) ([]domain.AuthorizationObject, error) {
	if strings.TrimSpace(query.RequestID) == "" {
		return nil, domain.ErrInvalidInput
	}
	objects, err := s.store.ListObjects(query.RequestID)
	if err != nil {
		return nil, err
	}
	text := strings.ToLower(strings.TrimSpace(query.Text))
	filtered := make([]domain.AuthorizationObject, 0, len(objects))
	for _, object := range objects {
		if query.Type != "" && object.Type != query.Type {
			continue
		}
		if text != "" && !strings.Contains(strings.ToLower(object.Name), text) && !strings.Contains(strings.ToLower(object.RefID), text) {
			continue
		}
		filtered = append(filtered, object)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		if query.Sort == "type" && filtered[i].Type != filtered[j].Type {
			return filtered[i].Type < filtered[j].Type
		}
		if query.Sort == "selected" {
			return filtered[i].SelectedAt.Before(filtered[j].SelectedAt)
		}
		return filtered[i].Name < filtered[j].Name
	})
	if query.Descending {
		reverseObjects(filtered)
	}
	return filtered, nil
}

func reverseObjects(objects []domain.AuthorizationObject) {
	for left, right := 0, len(objects)-1; left < right; left, right = left+1, right-1 {
		objects[left], objects[right] = objects[right], objects[left]
	}
}

func (s *Service) TypeCounts(requestID string) (map[domain.ObjectType]int, error) {
	objects, err := s.SearchResult(ResultQuery{RequestID: requestID})
	if err != nil {
		return nil, err
	}
	counts := map[domain.ObjectType]int{}
	for _, object := range objects {
		counts[object.Type]++
	}
	return counts, nil
}

func MergeSelections(existing, incoming []domain.AuthorizationObject) []domain.AuthorizationObject {
	merged := make(map[string]domain.AuthorizationObject, len(existing)+len(incoming))
	for _, object := range existing {
		merged[string(object.Type)+":"+object.RefID] = object
	}
	for _, object := range incoming {
		key := string(object.Type) + ":" + object.RefID
		if _, exists := merged[key]; !exists {
			merged[key] = object
		}
	}
	result := make([]domain.AuthorizationObject, 0, len(merged))
	for _, object := range merged {
		result = append(result, object)
	}
	domain.SortObjects(result)
	return result
}
