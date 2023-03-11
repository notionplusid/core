package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	gohttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	autocounter "github.com/notionplusid/core/app"
	"github.com/notionplusid/core/app/handler/http"
	"github.com/notionplusid/core/app/provider/notion"
	"github.com/notionplusid/core/app/service"
	"github.com/notionplusid/core/app/storage/datastore"
	"github.com/notionplusid/core/app/storage/inmemcache"
)

const (
	shutdownTO = 10 * time.Second
)

func main() {
	log.Print("Autocounter API starting")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go listenSig(cancel)

	env, err := NewEnv(ctx)
	if err != nil {
		log.Fatalf("Env: %s", err)
		return
	}
	log.Print("Env: OK")

	ds, err := datastore.New(ctx, env.GCloud.ProjectID)
	if err != nil {
		log.Fatalf("Datastore: %s", err)
		return
	}
	log.Print("Datastore: OK")

	inmem, err := inmemcache.New(ds)
	if err != nil {
		log.Fatalf("In-mem cache: %s", err)
		return
	}
	log.Print("In-mem cache: OK")
	if err := inmem.Sync(ctx); err != nil {
		log.Fatalf("In-mem cache: couldn't sync: %s", err)
		return
	}
	log.Print("In-mem cache: synced")

	tenant, err := service.NewTenant(inmem, notion.ExtConfig{
		ClientID:     env.Notion.ClientID,
		ClientSecret: env.Notion.ClientSecret,
		RedirectURI:  env.Notion.RedirectURI,
	})
	if err != nil {
		log.Fatalf("Tenant Service: %s", err)
		return
	}
	log.Print("Tenant Service: OK")

	table, err := service.NewTable(inmem)
	if err != nil {
		log.Fatalf("Table Service: %s", err)
		return
	}
	log.Print("Table Service: OK")

	// in case of the internal Notion extension - precreate the workspace.
	if env.Notion.ExtMode == NotionExtModeInternal {
		ws, err := autocounter.NewWorkspace(
			env.Notion.ClientID,
			env.Notion.ClientSecret,
		)
		if err != nil {
			log.Fatalf("Notion Ext Mode: %s", err)
			return
		}
		_, err = tenant.RegisterWorkspace(ctx, ws)
		if err != nil {
			log.Fatalf("Notion Ext Mode: Internal initial register: %s", err)
			return
		}
	}

	go func(ctx context.Context, procWssCount int64) {
		log.Printf("Worker: started")
		for {
			if err := tenant.ProcOldestUpdated(ctx, procWssCount, table.ProcWs); err != nil {
				log.Printf("Worker: couldn't process tables: %s", err)
			}

			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}(ctx, env.Notion.ProcWss)

	h, err := http.New(ctx, http.Dep{
		Tenant:     tenant,
		Table:      table,
		IsInternal: env.Notion.ExtMode == NotionExtModeInternal,
	})
	if err != nil {
		log.Fatalf("HTTP Handler: %s", err)
		return
	}
	log.Print("HTTP Handler: OK")

	host := fmt.Sprintf(":%s", env.HTTP.Port)

	log.Printf("Server: listening and serving on host %s", host)
	if err := ListenAndServe(ctx, host, h); err != nil {
		log.Printf("Server exited with an error: %s", err)
	}

	log.Print("Bye.")
}

func listenSig(callback func()) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	log.Printf("Service: Received SIGINT/SIGTERM. Allowing %s to shutdown gracefully.", shutdownTO)
	time.Sleep(shutdownTO)
	callback()
}

// ListenAndServe the HTTP traffic.
// Handles cancellation of the http request
func ListenAndServe(ctx context.Context, addr string, h *http.Handler) error {
	switch {
	case addr == "":
		return errors.New("address is required")
	case h == nil:
		return errors.New("handler is required")
	}

	httpServer := &gohttp.Server{
		Addr:    addr,
		Handler: h,
	}
	go func() {
		<-ctx.Done()
		log.Print("Server: exit signal received: exiting")
		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Printf("Server: couldn't close: %s", err)
		}
	}()

	err := httpServer.ListenAndServe()
	switch {
	case err == gohttp.ErrServerClosed:
		log.Print("Server: closed")
	case err != nil:
		log.Printf("Server: exited: %s", err)
	}

	return err
}
