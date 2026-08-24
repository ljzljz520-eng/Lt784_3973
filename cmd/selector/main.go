package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/permission-selector/internal/api"
	"example.com/permission-selector/internal/audit"
	"example.com/permission-selector/internal/config"
	"example.com/permission-selector/internal/domain"
	"example.com/permission-selector/internal/org"
	"example.com/permission-selector/internal/selector"
	"example.com/permission-selector/internal/store"
	"example.com/permission-selector/internal/workflow"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
	database, err := store.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	organization := org.NewService(database)
	auditLog := audit.NewService(database)
	selection := selector.NewService(database, organization, auditLog)
	flows := workflow.NewService(database, organization, selection, auditLog)
	if cfg.SeedDemo {
		if err := seedDemo(organization, database); err != nil {
			log.Fatal(err)
		}
	}
	server := api.NewServer(cfg, database, organization, selection, flows)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != context.Canceled {
			log.Printf("server stopped: %v", err)
		}
	}()
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}

func seedDemo(organization *org.Service, database *store.Store) error {
	if _, err := database.FindNode("root"); err == nil {
		return nil
	}
	if _, err := organization.AddDepartment("root", "", "全组织"); err != nil {
		return err
	}
	engineering, err := organization.AddDepartment("engineering", "root", "研发中心")
	if err != nil {
		return err
	}
	people := []struct{ id, username, display string }{{"alice", "alice", "Alice"}, {"bob", "bob", "Bob"}, {"chen", "chen", "Chen"}}
	for _, person := range people {
		if _, err := organization.AddAccount(person.id, engineering.ID, person.username, person.display, person.username+"@example.com"); err != nil {
			return err
		}
	}
	return nil
}

func defaultRequest(id string) domain.RequestCommand {
	return domain.RequestCommand{RequestID: id, Actor: "admin", Title: "Default access", Reason: "Initial access selection"}
}
