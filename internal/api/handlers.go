package api

import (
	"net/http"
	"strconv"

	"example.com/permission-selector/internal/domain"
	"example.com/permission-selector/internal/report"
	"example.com/permission-selector/internal/selector"
)

func (s *Server) tree(writer http.ResponseWriter, request *http.Request) {
	nodes, err := s.org.Tree()
	if err != nil {
		writeError(writer, err)
		return
	}
	if err := s.org.ValidateTree(); err != nil {
		writeError(writer, err)
		return
	}
	insights, err := s.org.Insights()
	if err != nil {
		writeError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"nodes": nodes, "insights": insights})
}

func (s *Server) accounts(writer http.ResponseWriter, request *http.Request) {
	page, pageSize := queryPage(request)
	filter := domain.AccountFilter{NodeID: request.URL.Query().Get("node"), Query: request.URL.Query().Get("q"), OnlyActive: true}
	result, err := s.selector.LoadAccounts(filter, page, pageSize)
	if err != nil {
		writeError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, result)
}

func (s *Server) createRequest(writer http.ResponseWriter, request *http.Request) {
	var command domain.RequestCommand
	if err := decodeBody(request, &command); err != nil {
		writeError(writer, err)
		return
	}
	created, err := s.workflow.CreateRequest(command)
	if err != nil {
		writeError(writer, err)
		return
	}
	writeJSON(writer, http.StatusCreated, created)
}

func (s *Server) request(writer http.ResponseWriter, request *http.Request) {
	id := pathValue(request.URL.Path, "/api/requests/")
	if id == "" {
		writeError(writer, domain.ErrNotFound)
		return
	}
	result, err := s.workflow.GetRequest(id)
	if err != nil {
		writeError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, result)
}

func (s *Server) requestAction(writer http.ResponseWriter, request *http.Request) {
	path := pathValue(request.URL.Path, "/api/requests/")
	parts := splitPath(path)
	if len(parts) != 2 {
		writeError(writer, domain.ErrNotFound)
		return
	}
	requestID, action := parts[0], parts[1]
	switch action {
	case "select":
		var command domain.SelectionCommand
		if err := decodeBody(request, &command); err != nil {
			writeError(writer, err)
			return
		}
		command.RequestID = requestID
		object, err := s.selector.Select(command)
		if err != nil {
			writeError(writer, err)
			return
		}
		writeJSON(writer, http.StatusCreated, object)
	case "confirm":
		actor := request.URL.Query().Get("actor")
		updated, summary, err := s.selector.Confirm(requestID, actor)
		if err != nil {
			writeError(writer, err)
			return
		}
		writeJSON(writer, http.StatusOK, map[string]any{"request": updated, "summary": summary})
	case "dispatch":
		page, pageSize := queryPage(request)
		actor := request.URL.Query().Get("actor")
		record, err := s.selector.Dispatch(requestID, actor, page, pageSize)
		if err != nil {
			writeError(writer, err)
			return
		}
		writeJSON(writer, http.StatusOK, record)
	case "result":
		if request.URL.Query().Get("q") != "" || request.URL.Query().Get("type") != "" {
			result, err := s.selector.SearchResult(selector.ResultQuery{RequestID: requestID, Text: request.URL.Query().Get("q"), Type: domain.ObjectType(request.URL.Query().Get("type")), Sort: request.URL.Query().Get("sort")})
			if err != nil {
				writeError(writer, err)
				return
			}
			writeJSON(writer, http.StatusOK, map[string]any{"items": result, "total": len(result)})
			return
		}
		page, pageSize := queryPage(request)
		result, err := s.selector.ResultPage(requestID, page, pageSize)
		if err != nil {
			writeError(writer, err)
			return
		}
		writeJSON(writer, http.StatusOK, result)
	case "publish":
		actor := request.URL.Query().Get("actor")
		published, err := s.workflow.Publish(requestID, actor)
		if err != nil {
			writeError(writer, err)
			return
		}
		writeJSON(writer, http.StatusOK, published)
	case "export":
		format := request.URL.Query().Get("format")
		if format == "json" {
			data, err := report.JSON(s.store, requestID)
			if err != nil {
				writeError(writer, err)
				return
			}
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(data)
			return
		}
		data, err := report.CSV(s.store, requestID)
		if err != nil {
			writeError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "text/csv")
		writer.Header().Set("Content-Disposition", "attachment; filename=authorization.csv")
		_, _ = writer.Write(data)
	default:
		writeError(writer, domain.ErrNotFound)
	}
}

func queryPage(request *http.Request) (int, int) {
	page, _ := strconv.Atoi(request.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(request.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return page, pageSize
}

func splitPath(value string) []string {
	parts := make([]string, 0, 2)
	start := 0
	for index := 0; index <= len(value); index++ {
		if index == len(value) || value[index] == '/' {
			if index > start {
				parts = append(parts, value[start:index])
			}
			start = index + 1
		}
	}
	return parts
}
