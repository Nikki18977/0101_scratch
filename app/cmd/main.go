package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"uploader/svc/uploader"
)

func main() {
	var err error

	log.Print("starting: create service")
	svc := uploader.NewSvc()

	log.Print("starting: create router")
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Post("/upload", func(writer http.ResponseWriter, request *http.Request) {
		//err := request.ParseMultipartForm(10 * 1024 * 1024)
		//if err != nil {
		//	log.Print(err)
		//	http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		//	return
		//}
		file, header, err := request.FormFile("file")
		if err != nil {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		err = svc.Upload(request.Context(), header, file)
		if err != nil {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("OK"))
	})

	addr := net.JoinHostPort("0.0.0.0", "9999")
	log.Printf("starting: listen and serve on: %s", addr)
	server := &http.Server{Addr: addr, Handler: router}
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		sig := <-sigChannel
		log.Printf("stopping: got %s signal", sig)

		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("stopping: graceful shutdown timed out.. forcing exit.")
			}
		}()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	<-serverCtx.Done()
}
