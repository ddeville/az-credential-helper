package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/ddeville/az-credential-helper/pkg"
)

const scope string = "https://storage.azure.com/.default"
const hostSuffix string = ".blob.core.windows.net"

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Missing command")
	}
	if strings.TrimSpace(os.Args[1]) != "get" {
		log.Fatalf("Invalid command: %s. Allowed commands: get", os.Args[1])
	}

	var req struct {
		URI string `json:"uri"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}
	if req.URI == "" {
		log.Fatal("Missing uri in input")
	}

	parsed, err := url.Parse(req.URI)
	if err != nil {
		log.Fatalf("Invalid uri: %v", err)
	}
	if !strings.HasSuffix(parsed.Host, hostSuffix) {
		log.Fatalf("Unexpected host in URI: %s. Should be: *%s", parsed.Host, hostSuffix)
	}

	accessToken, err := pkg.GetAzureAccessToken(scope)
	if err != nil {
		log.Fatalf("Failed to retrieve token from azure: %v", err)
	}

	res := map[string]map[string][]string{
		"headers": {
			"Authorization": {"Bearer " + accessToken.Token},
			"x-ms-version":  {"2024-05-04"},
		},
	}
	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(res); err != nil {
		log.Fatalf("Failed to write response: %v", err)
	}
}
