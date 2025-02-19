package configs

import (
	"context"
	"encoding/json"
	"log"
	"os"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

func GetFirebaseCloudCredentials() string {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create secret manager client: %v", err)
	}
	defer client.Close()

	secretName := os.Getenv("SECRET_NAME")
	if secretName == "" {
		log.Fatalf("SECRET_NAME environment variable not set")
	}

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretName,
	}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		log.Fatalf("Failed to access secret version: %v", err)
	}

	var credentials map[string]interface{}
	if err := json.Unmarshal(result.Payload.Data, &credentials); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	credentialsPath := os.Getenv("CREDENTIALS_PATH")
	if credentialsPath == "" {
		log.Fatalf("CREDENTIALS_PATH environment variable not set")
	}

	if err := os.WriteFile(credentialsPath, result.Payload.Data, 0644); err != nil {
		log.Fatalf("Failed to write credentials file: %v", err)
	}

	return credentialsPath
}

func GetFirebaseLocalCredentials() string {
	return os.Getenv("CREDENTIALS_PATH")
}
