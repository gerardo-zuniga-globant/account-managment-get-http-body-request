package main

import (
	"context"
	"encoding/json"
	"fmt"
	bun "github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/reqlog"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	ContentTypeHeaderKey                  = "Content-Type"
	ContentTypeApplicationJsonHeaderValue = "application/json"
	V1BaseEndpoint                        = "/v1"
	UserPathGroup                         = "/users"
	ServerPort                            = "8080"
)

func main() {
	router := bun.New(bun.Use(reqlog.NewMiddleware()))
	setAPIRoutes(router)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", ServerPort),
		Handler: router,
	}
	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Println("Server Shutdown:", err)
	}
	<-ctx.Done()
}

func decodeSaveTransactionRequest(body io.ReadCloser) (fuwdn *FindUserWithDisplayNameCommand, apiEr error) {
	err := json.NewDecoder(body).Decode(&fuwdn)
	if err != nil {
		fmt.Println(err)
	}
	return
}

type FindUserWithDisplayNameCommand struct {
	DisplayName string
}

func setAPIRoutes(router *bun.Router) {
	// Transaction group middleware
	v1Group := router.NewGroup(V1BaseEndpoint).Use(func(next bun.HandlerFunc) bun.HandlerFunc {
		return func(w http.ResponseWriter, req bun.Request) error {
			w.Header().Add(ContentTypeHeaderKey, ContentTypeApplicationJsonHeaderValue)
			return next(w, req)
		}
	})

	// group for 'users' routes
	v1Group.WithGroup(UserPathGroup, func(g *bun.Group) {
		// mutes/bans user from chat
		g.GET("/:id", HandleGetByDisplayName)
	})

}

func HandleGetByDisplayName(w http.ResponseWriter, req bun.Request) error {
	defer req.Body.Close()
	reqFindUserWithDisplayNameCommand, err := decodeSaveTransactionRequest(req.Body)
	if err != nil {
		return err
	}
	fmt.Println(*reqFindUserWithDisplayNameCommand)
	return nil
}
