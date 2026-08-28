package report

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"example.com/permission-selector/internal/domain"
	"example.com/permission-selector/internal/store"
)

type RequestReport struct {
	Request     domain.AuthorizationRequest  `json:"request"`
	Objects     []domain.AuthorizationObject `json:"objects"`
	Summary     domain.SelectionSummary      `json:"summary"`
	Dispatches  []domain.DispatchRecord      `json:"dispatches"`
	Audit       []domain.AuditEvent          `json:"audit"`
	GeneratedAt time.Time                    `json:"generated_at"`
}

type Row struct {
	RequestID  string
	Title      string
	Status     string
	ObjectID   string
	ObjectType string
	ObjectName string
	SelectedAt string
}

func Build(database *store.Store, requestID string) (RequestReport, error) {
	if strings.TrimSpace(requestID) == "" {
		return RequestReport{}, domain.ErrInvalidInput
	}
	bundle, err := database.LoadRequestBundle(requestID)
	if err != nil {
		return RequestReport{}, err
	}
	if err := database.EnsureBundleConsistent(bundle); err != nil {
		return RequestReport{}, err
	}
	request, objects, dispatches, audit := bundle.Request, bundle.Objects, bundle.Dispatches, bundle.Audits
	domain.SortObjects(objects)
	return RequestReport{Request: request, Objects: objects, Summary: domain.SummarizeObjects(objects), Dispatches: dispatches, Audit: audit, GeneratedAt: time.Now().UTC()}, nil
}

func Rows(report RequestReport) []Row {
	rows := make([]Row, 0, len(report.Objects))
	for _, object := range report.Objects {
		rows = append(rows, Row{RequestID: report.Request.ID, Title: report.Request.Title, Status: string(report.Request.Status), ObjectID: object.RefID, ObjectType: string(object.Type), ObjectName: object.Name, SelectedAt: object.SelectedAt.Format(time.RFC3339)})
	}
	if len(rows) == 0 {
		rows = append(rows, Row{RequestID: report.Request.ID, Title: report.Request.Title, Status: string(report.Request.Status)})
	}
	return rows
}

func CSV(database *store.Store, requestID string) ([]byte, error) {
	result, err := Build(database, requestID)
	if err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	if err := writer.Write([]string{"request_id", "title", "status", "object_id", "object_type", "object_name", "selected_at"}); err != nil {
		return nil, err
	}
	for _, row := range Rows(result) {
		if err := writer.Write([]string{row.RequestID, row.Title, row.Status, row.ObjectID, row.ObjectType, row.ObjectName, row.SelectedAt}); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func JSON(database *store.Store, requestID string) ([]byte, error) {
	result, err := Build(database, requestID)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(result, "", "  ")
}

func DispatchSummary(report RequestReport) string {
	parts := []string{fmt.Sprintf("request=%s", report.Request.ID), fmt.Sprintf("objects=%d", report.Summary.Total), fmt.Sprintf("departments=%d", report.Summary.Departments), fmt.Sprintf("people=%d", report.Summary.People), fmt.Sprintf("dispatches=%d", len(report.Dispatches))}
	return strings.Join(parts, " ")
}

func Timeline(report RequestReport) []string {
	entries := make([]string, 0, len(report.Audit)+len(report.Dispatches)+1)
	entries = append(entries, report.Request.CreatedAt.Format(time.RFC3339)+" created")
	for _, event := range report.Audit {
		entries = append(entries, event.CreatedAt.Format(time.RFC3339)+" "+event.Action+" by "+event.Actor)
	}
	for _, dispatch := range report.Dispatches {
		entries = append(entries, dispatch.CreatedAt.Format(time.RFC3339)+" dispatch #"+strconv.Itoa(dispatch.Sequence))
	}
	sort.Strings(entries)
	return entries
}
