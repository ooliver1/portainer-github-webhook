package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
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
		var body []byte
		_, err := r.Body.Read(body)

		if err != nil {
			log.Printf("Error reading request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		hash.Write(body)
		signature := hash.Sum(nil)
		if !hmac.Equal(append([]byte("sha256="), signature...), []byte(r.Header.Get("X-Hub-Signature-256"))) {
			log.Printf("Invalid signature %v %v", string(signature), r.Header.Get("X-Hub-Signature-256"))
			w.WriteHeader(http.StatusTeapot)
			return
		}

		var payload PushPayload
		err = json.Unmarshal(body, &payload)
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
		if err != nil || res.StatusCode != http.StatusOK {
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
