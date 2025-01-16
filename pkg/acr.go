package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	scope          string = "https://containerregistry.azure.net/.default"
	registrySuffix string = ".azurecr.io"
	username       string = "00000000-0000-0000-0000-000000000000"
)

type DockerCredentials struct {
	Username string
	Password string
}

func GetDockerCredentials(serverURL string) (*DockerCredentials, error) {
	if !strings.Contains(serverURL, "://") && !strings.HasPrefix(serverURL, "//") {
		serverURL = "//" + serverURL
	}

	registryURL, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}

	registry := registryURL.Hostname()
	if !strings.HasSuffix(registry, registrySuffix) {
		return nil, fmt.Errorf("non-acr registry: %s", registry)
	}

	accessToken, err := GetAzureAccessToken(scope)
	if err != nil {
		log.Fatalf("failed to get access token: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	refreshToken, err := getACRRefreshToken(ctx, registry, accessToken.Token)
	if err != nil {
		log.Fatalf("failed to get ACR refresh token: %v", err)
	}

	return &DockerCredentials{Username: username, Password: refreshToken}, nil
}

func getACRRefreshToken(ctx context.Context, registry string, aadAccessToken string) (string, error) {
	exchangeURL := fmt.Sprintf("https://%s/oauth2/exchange", registry)

	form := url.Values{}
	form.Set("grant_type", "access_token")
	form.Set("service", registry)
	form.Set("access_token", aadAccessToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, exchangeURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("exchange failed: %s", body)
	}

	var parsed struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}

	return parsed.RefreshToken, nil
}
