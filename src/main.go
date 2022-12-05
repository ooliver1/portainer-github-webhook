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

	"gopkg.in/yaml.v3"
)

type Config struct {
	PortainerUrl string    `yaml:"portainer_url"`
	SecretKey    string    `yaml:"secret_key"`
	Webhooks     []Webhook `yaml:"webhooks"`
}

type Webhook struct {
	Uuid   string `yaml:"uuid"`
	Repo   string `yaml:"repo"`
	Branch string `yaml:"branch"`
}

type PushPayload struct {
	Ref        string     `json:"ref"`
	Repository Repository `json:"repository"`
}

type Repository struct {
	FullName string `json:"full_name"`
}

func handlerWithConfig(config Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		hash := hmac.New(sha256.New, []byte(config.SecretKey))
		var body []byte
		_, err := r.Body.Read(body)

		if err != nil {
			log.Printf("Error reading request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		hash.Write(body)
		signature := hash.Sum(nil)
		if !hmac.Equal(signature, []byte(r.Header.Get("X-Hub-Signature"))) {
			log.Printf("Invalid signature")
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

		for _, webhook := range config.Webhooks {
			if strings.EqualFold(webhook.Repo, payload.Repository.FullName) &&
				strings.EqualFold(webhook.Branch, strings.TrimPrefix(payload.Ref, "refs/heads/")) {

				log.Printf("Found webhook for %s", payload.Repository.FullName)
				res, err := http.Post(fmt.Sprintf("%s/api/stacks/webhooks/%s", config.PortainerUrl, webhook.Uuid), "", nil)
				if err != nil || res.StatusCode != http.StatusOK {
					log.Printf("Error creating request: %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusNoContent)
			}
		}
	}
}

func main() {
	configFile, err := os.ReadFile("config.yaml")
	if err != nil {
		fmt.Printf("Error reading config file: %v", err)
		os.Exit(1)
	}

	var config Config
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		fmt.Printf("Error parsing config file: %v", err)
		os.Exit(1)
	}

	http.HandleFunc("/", handlerWithConfig(config))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3473"
	}
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
