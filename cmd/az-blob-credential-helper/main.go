package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

const scope string = "https://storage.azure.com/.default"
const hostSuffix string = ".blob.core.windows.net"

func getAzureAccessToken() (*azcore.AccessToken, error) {
	var creds []azcore.TokenCredential
	var errs []error

	// Create the chain manually so that we can exclude managed identity credentials since it seems to
	// return a useless token even when run locally from a docker base image...

	envCred, err := azidentity.NewEnvironmentCredential(nil)
	if err == nil {
		creds = append(creds, envCred)
	} else {
		errs = append(errs, fmt.Errorf("failed to create EnvironmentCredential: %v", err))
	}

	wiCred, err := azidentity.NewWorkloadIdentityCredential(nil)
	if err == nil {
		creds = append(creds, wiCred)
	} else {
		errs = append(errs, fmt.Errorf("failed to create WorkloadIdentityCredential: %v", err))
	}

	azCred, err := azidentity.NewAzureCLICredential(nil)
	if err == nil {
		creds = append(creds, azCred)
	} else {
		errs = append(errs, fmt.Errorf("failed to create AzureCLICredential: %v", err))
	}

	azdCred, err := azidentity.NewAzureDeveloperCLICredential(nil)
	if err == nil {
		creds = append(creds, azdCred)
	} else {
		errs = append(errs, fmt.Errorf("failed to create AzureDeveloperCLICredential: %v", err))
	}

	if len(creds) == 0 {
		return nil, errors.Join(errs...)
	}

	chain, err := azidentity.NewChainedTokenCredential(creds, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create ChainedTokenCredential: %w", err)
	}

	token, err := chain.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{scope},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve token: %w", err)
	}

	return &token, nil
}

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
	if !strings.HasSuffix(parsed.Host, hostSuffix) {
		fail(fmt.Sprintf("Unexpected host in URI: %s. Should be: *%s", parsed.Host, hostSuffix))
	}

	accessToken, err := getAzureAccessToken()
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
