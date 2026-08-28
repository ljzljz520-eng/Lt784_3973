package domain

import "sort"

type AccountFilter struct {
	NodeID     string
	Query      string
	OnlyActive bool
}

func NormalizeFilter(filter AccountFilter) AccountFilter {
	if filter.NodeID == "" {
		filter.NodeID = "root"
	}
	if !filter.OnlyActive {
		filter.OnlyActive = true
	}
	return filter
}

func SortAccounts(accounts []Account) {
	sort.SliceStable(accounts, func(i, j int) bool {
		if accounts[i].Display == accounts[j].Display {
			return accounts[i].ID < accounts[j].ID
		}
		return accounts[i].Display < accounts[j].Display
	})
}

func BuildAccountPage(accounts []Account, page, pageSize int) (Page[Account], error) {
	SortAccounts(accounts)
	start, end, err := PageBounds(page, pageSize, len(accounts))
	if err != nil {
		return Page[Account]{Page: page, PageSize: pageSize, Total: len(accounts)}, err
	}
	items := append([]Account(nil), accounts[start:end]...)
	return Page[Account]{
		Items: items, Page: page, PageSize: pageSize, Total: len(accounts),
		HasNext: end < len(accounts), HasPrevious: start > 0,
	}, nil
}

func BuildObjectPage(objects []AuthorizationObject, page, pageSize int) (Page[AuthorizationObject], error) {
	SortObjects(objects)
	start, end, err := PageBounds(page, pageSize, len(objects))
	if err != nil {
		return Page[AuthorizationObject]{Page: page, PageSize: pageSize, Total: len(objects)}, err
	}
	items := append([]AuthorizationObject(nil), objects[start:end]...)
	return Page[AuthorizationObject]{
		Items: items, Page: page, PageSize: pageSize, Total: len(objects),
		HasNext: end < len(objects), HasPrevious: start > 0,
	}, nil
}
