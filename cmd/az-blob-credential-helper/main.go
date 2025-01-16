package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ddeville/az-credential-helper/pkg"
)

const (
	scope      string = "https://storage.azure.com/.default"
	hostSuffix string = ".blob.core.windows.net"
)

var rootCmd = &cobra.Command{
	Use: "az-blob-credential-helper",
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "retrieve credentials for the storage account passed as json via stdin.",
	Run:   func(cmd *cobra.Command, args []string) { get() },
}

func init() {
	rootCmd.AddCommand(getCmd)
}

func get() {
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
		log.Fatalf("Unexpected host: %s should end with %s", parsed.Host, hostSuffix)
	}

	accessToken, err := pkg.GetAzureAccessToken(scope)
	if err != nil {
		log.Fatalf("Failed to retrieve token: %v", err)
	}

	res := map[string]map[string][]string{
		"headers": {
			"Authorization": {"Bearer " + accessToken.Token},
			"x-ms-version":  {"2024-05-04"},
		},
	}
	if err := json.NewEncoder(os.Stdout).Encode(res); err != nil {
		log.Fatalf("Failed to write response: %v", err)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Failed to run command: %v", err)
	}
}
