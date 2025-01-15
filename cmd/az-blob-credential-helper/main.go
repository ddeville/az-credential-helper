package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/ddeville/az-credential-helper/pkg"
)

const Scope string = "https://storage.azure.com/.default"
const HostSuffix string = ".blob.core.windows.net"

func fail(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		fail("Missing command")
	}
	if strings.TrimSpace(os.Args[1]) != "get" {
		fail(fmt.Sprintf("Invalid command: %s. Allowed commands: get", os.Args[1]))
	}

	var req struct {
		URI string `json:"uri"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		fail(fmt.Sprintf("Failed to parse JSON: %v", err))
	}
	if req.URI == "" {
		fail("Missing uri in input")
	}

	parsed, err := url.Parse(req.URI)
	if err != nil {
		fail(fmt.Sprintf("Invalid uri: %v", err))
	}
	if !strings.HasSuffix(parsed.Host, HostSuffix) {
		fail(fmt.Sprintf("Unexpected host in URI: %s. Should be: *%s", parsed.Host, HostSuffix))
	}

	accessToken, err := pkg.GetAzureAccessToken(Scope)
	if err != nil {
		fail(fmt.Sprintf("Failed to retrieve token from azure: %v", err))
	}

	res := map[string]map[string][]string{
		"headers": {
			"Authorization": {"Bearer " + accessToken.Token},
			"x-ms-version":  {"2024-05-04"},
		},
	}
	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(res); err != nil {
		fail(fmt.Sprintf("Failed to write response: %v", err))
	}
}
