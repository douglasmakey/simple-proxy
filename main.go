package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

var defaultHTTPClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		TLSHandshakeTimeout:   3 * time.Second,
		ResponseHeaderTimeout: 3 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

func main() {
	r := mux.NewRouter().SkipClean(true)
	r.HandleFunc("/proxy/{url:.*}", proxy).Methods(http.MethodGet, http.MethodPost)

	server := http.Server{
		Addr:    ":8000",
		Handler: r,
	}

	go func() {
		log.Printf("Starting HTTP Proxy Server. Listening at %s", server.Addr)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		} else {
			log.Println("Server closed!")
		}
	}()

	// Check for a closing signal
	// Graceful shutdown
	sigquit := make(chan os.Signal, 1)
	signal.Notify(sigquit, os.Interrupt, syscall.SIGTERM)
	sig := <-sigquit
	log.Printf("caught sig: %+v", sig)
	log.Printf("Gracefully shutting down server...")

	if err := server.Shutdown(context.Background()); err != nil {
		log.Printf("Unable to shut down server: %v", err)
	} else {
		log.Println("Server stopped")
	}
}

func proxy(w http.ResponseWriter, r *http.Request) {
	targetURL := mux.Vars(r)["url"]
	if targetURL == "" {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}

	if _, err := url.ParseRequestURI(targetURL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// COPY headers origen to destiny
	req.Header = r.Header.Clone()

	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		var nerr net.Error
		if errors.As(err, &nerr) {
			if nerr.Timeout() {
				http.Error(w, err.Error(), http.StatusGatewayTimeout)
				return
			}
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	// COPY headers destiny to origen
	copyHeaders(w.Header(), resp.Header)

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	return
}

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
