package main

import (
	"errors"
	"log"

	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/spf13/cobra"

	"github.com/ddeville/az-credential-helper/pkg"
)

var rootCmd = &cobra.Command{
	Use: "docker-credential-acr-login",
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "retrieve docker credentials for the ACR passed via stdin.",
	Run: func(cmd *cobra.Command, args []string) {
		credentials.Serve(AzureContainerRegistryCredentialsHelper{})
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}

type AzureContainerRegistryCredentialsHelper struct{}

func (h AzureContainerRegistryCredentialsHelper) Get(serverURL string) (string, string, error) {
	creds, err := pkg.GetDockerCredentials(serverURL)
	if err != nil {
		return "", "", err
	}
	return creds.Username, creds.Password, nil
}

func (h AzureContainerRegistryCredentialsHelper) Add(creds *credentials.Credentials) error {
	return errors.New("not implemented")
}

func (h AzureContainerRegistryCredentialsHelper) Delete(serverURL string) error {
	return errors.New("not implemented")
}

func (h AzureContainerRegistryCredentialsHelper) List() (map[string]string, error) {
	return nil, errors.New("not implemented")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Failed to run command: %v", err)
	}
}
