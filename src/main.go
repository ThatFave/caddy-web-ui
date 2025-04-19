package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("/static")))

	caddy_api := os.Getenv("CADDY_API")

	_, err := os.ReadFile("/config/Caddyfile")
	if err != nil {
		os.Create("/config/Caddyfile")
		return
	}

	// Serve Caddyfile
	http.HandleFunc("/api/caddyfile", func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile("/config/Caddyfile")
		if err != nil {
			http.Error(w, "Failed to read Caddyfile: "+err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write(data)
	})

	http.HandleFunc("/api/caddyfile/validate", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to validate Caddyfile: "+err.Error(), 400)
		}

		req, err := http.NewRequest("POST", caddy_api+"/adapt", bytes.NewReader(body))
		if err != nil {
			http.Error(w, "Failed to create validation request: "+err.Error(), 500)
			return
		}
		req.Header.Set("Content-Type", "text/caddyfile")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, "Failed to send validation request: "+err.Error(), 500)
			return
		}
		defer resp.Body.Close()

		log.Println(w, "response: "+strconv.Itoa(resp.StatusCode), resp.StatusCode)

		if resp.StatusCode != 200 {
			msg, _ := io.ReadAll(resp.Body)
			http.Error(w, "Validation failed:\n"+string(msg), resp.StatusCode)
			log.Println(w, "Validation failed:\n"+string(msg), resp.StatusCode)
			return
		}

		w.Write([]byte("Validation successful"))
	})

	// Update and reload Caddyfile
	http.HandleFunc("/api/caddyfile/update", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body: "+err.Error(), 400)
			return
		}

		// Save it to disk
		if err := os.WriteFile("/config/Caddyfile", body, 0644); err != nil {
			http.Error(w, "Failed to write file: "+err.Error(), 500)
			return
		}

		// Reload it into Caddy
		req, err := http.NewRequest("POST", caddy_api+"/load", bytes.NewReader(body))
		if err != nil {
			http.Error(w, "Failed to create reload request: "+err.Error(), 500)
			return
		}
		req.Header.Set("Content-Type", "text/caddyfile")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, "Failed to send reload request: "+err.Error(), 500)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			msg, _ := io.ReadAll(resp.Body)
			http.Error(w, "Reload failed:\n"+string(msg), resp.StatusCode)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Reload successful"))
	})

	http.ListenAndServe(":8080", nil)
}
