package report

import (
	"strings"

	"example.com/permission-selector/internal/domain"
)

func FilterObjects(objects []domain.AuthorizationObject, query string, objectType domain.ObjectType) []domain.AuthorizationObject {
	query = strings.ToLower(strings.TrimSpace(query))
	filtered := make([]domain.AuthorizationObject, 0, len(objects))
	for _, object := range objects {
		if objectType != "" && object.Type != objectType {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(object.Name), query) && !strings.Contains(strings.ToLower(object.RefID), query) {
			continue
		}
		filtered = append(filtered, object)
	}
	return filtered
}

func GroupByType(objects []domain.AuthorizationObject) map[domain.ObjectType][]domain.AuthorizationObject {
	groups := map[domain.ObjectType][]domain.AuthorizationObject{domain.ObjectTypeDepartment: {}, domain.ObjectTypePerson: {}}
	for _, object := range objects {
		groups[object.Type] = append(groups[object.Type], object)
	}
	return groups
}
