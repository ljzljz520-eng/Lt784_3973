package metrics

import (
	"sort"
	"time"

	"example.com/permission-selector/internal/domain"
	"example.com/permission-selector/internal/store"
)

type Report struct {
	Requests       int            `json:"requests"`
	Confirmed      int            `json:"confirmed"`
	Published      int            `json:"published"`
	Accounts       int            `json:"accounts"`
	Departments    int            `json:"departments"`
	SelectionTotal int            `json:"selection_total"`
	ByType         map[string]int `json:"by_type"`
	GeneratedAt    time.Time      `json:"generated_at"`
}

func BuildReport(database *store.Store) (Report, error) {
	requests, err := database.ListRequests()
	if err != nil {
		return Report{}, err
	}
	totalRequests, err := database.Count(store.BucketRequests)
	if err != nil {
		return Report{}, err
	}
	accounts, err := database.ActiveAccounts()
	if err != nil {
		return Report{}, err
	}
	nodes, err := database.ListNodes()
	if err != nil {
		return Report{}, err
	}
	objects, err := database.ListObjects("")
	if err != nil {
		return Report{}, err
	}
	report := Report{Requests: totalRequests, Accounts: len(accounts), Departments: len(nodes), ByType: map[string]int{}, GeneratedAt: time.Now().UTC()}
	for _, request := range requests {
		switch request.Status {
		case domain.StatusConfirmed, domain.StatusDispatched:
			report.Confirmed++
		case domain.StatusPublished:
			report.Published++
		}
	}
	for _, object := range objects {
		report.SelectionTotal++
		report.ByType[string(object.Type)]++
	}
	return report, nil
}

func RecentRequests(database *store.Store, limit int) ([]domain.AuthorizationRequest, error) {
	requests, err := database.ListRequests()
	if err != nil {
		return nil, err
	}
	sort.Slice(requests, func(i, j int) bool { return requests[i].UpdatedAt.After(requests[j].UpdatedAt) })
	if limit < 1 || limit >= len(requests) {
		return requests, nil
	}
	return requests[:limit], nil
}
