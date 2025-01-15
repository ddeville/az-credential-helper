package pkg

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

func GetAzureAccessToken(scope string) (*azcore.AccessToken, error) {
	cred, err := getAzureChainedTokenCredential()
	if err != nil {
		return nil, fmt.Errorf("failed to create ChainedTokenCredential: %w", err)
	}

	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{scope},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve token: %w", err)
	}

	return &token, nil
}

func getAzureChainedTokenCredential() (*azidentity.ChainedTokenCredential, error) {
	var creds []azcore.TokenCredential
	var errs []error

	// Create the chain manually so that we can exclude managed identity credentials since it seems to
	// return a useless token even when run locally from a docker base image...

	envCred, err := azidentity.NewEnvironmentCredential(nil)
	if err == nil {
		creds = append(creds, envCred)
	} else {
		errs = append(errs, fmt.Errorf("failed to create EnvironmentCredential: %w", err))
	}

	wiCred, err := azidentity.NewWorkloadIdentityCredential(nil)
	if err == nil {
		creds = append(creds, wiCred)
	} else {
		errs = append(errs, fmt.Errorf("failed to create WorkloadIdentityCredential: %w", err))
	}

	azCred, err := azidentity.NewAzureCLICredential(nil)
	if err == nil {
		creds = append(creds, azCred)
	} else {
		errs = append(errs, fmt.Errorf("failed to create AzureCLICredential: %w", err))
	}

	azdCred, err := azidentity.NewAzureDeveloperCLICredential(nil)
	if err == nil {
		creds = append(creds, azdCred)
	} else {
		errs = append(errs, fmt.Errorf("failed to create AzureDeveloperCLICredential: %w", err))
	}

	if len(creds) == 0 {
		return nil, errors.Join(errs...)
	}

	chain, err := azidentity.NewChainedTokenCredential(creds, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create ChainedTokenCredential: %w", err)
	}

	return chain, nil
}
