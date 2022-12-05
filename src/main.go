package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type PushPayload struct {
	Ref        string     `json:"ref"`
	Repository Repository `json:"repository"`
}

type Repository struct {
	FullName string `json:"full_name"`
}

func handlerWithConfig(secretKey, portainerUrl string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		event := r.Header.Get("X-GitHub-Event")
		if event == "ping" {
			w.WriteHeader(http.StatusNoContent)
			return
		} else if event != "push" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		hash := hmac.New(sha256.New, []byte(secretKey))
		buf, err := io.ReadAll(r.Body)

		if err != nil {
			log.Printf("Error reading request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err = hash.Write(buf)
		if err != nil {
			log.Printf("Error hashing request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		signature := hex.EncodeToString(hash.Sum(nil))
		if !hmac.Equal([]byte(signature), []byte(r.Header.Get("X-Hub-Signature-256"))) {
			log.Printf("Invalid signature %v %v", string(append([]byte("sha256="), signature...)), r.Header.Get("X-Hub-Signature-256"))
			w.WriteHeader(http.StatusTeapot)
			log.Printf("%v", signature == r.Header.Get("X-Hub-Signature-256"))
			log.Printf("%v", signature)
			log.Printf("%v", r.Header.Get("X-Hub-Signature-256"))
			return
		}

		var payload PushPayload
		err = json.Unmarshal(buf, &payload)
		if err != nil {
			log.Printf("Error unmarshalling payload: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !strings.EqualFold(r.URL.Query().Get("branch"), strings.TrimPrefix(payload.Ref, "refs/heads/")) {
			log.Printf("Not the desired branch: %v", payload.Ref)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		uuid := r.URL.Query().Get("uuid")
		res, err := http.Post(fmt.Sprintf("%s/api/stacks/webhooks/%s", portainerUrl, uuid), "", nil)
		if err != nil || res.StatusCode != http.StatusNoContent {
			log.Printf("Error creating request: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}

}

func main() {
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		fmt.Fprintf(os.Stderr, "SECRET_KEY environment variable not set")
		os.Exit(1)
	}
	portainerUrl := os.Getenv("PORTAINER_URL")
	if portainerUrl == "" {
		fmt.Fprintf(os.Stderr, "PORTAINER_URL environment variable not set")
		os.Exit(1)
	}
	http.HandleFunc("/", handlerWithConfig(secretKey, portainerUrl))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3473"
	}
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Fprintf(os.Stderr, "server closed\n")
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "error starting server: %s\n", err)
		os.Exit(1)
	}
}
