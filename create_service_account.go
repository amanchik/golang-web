package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/firestore"
	iam "google.golang.org/api/iam/v1"
)

func createClient(ctx context.Context) *firestore.Client {
	// Sets your Google Cloud Platform project ID.
	projectID := "golang-370407"

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	// Close client when done with
	// defer client.Close()
	return client
}

// createServiceAccount creates a service account.
func createServiceAccount(w io.Writer, projectID, name, displayName string) (*iam.ServiceAccount, error) {
	ctx := context.Background()
	service, err := iam.NewService(ctx)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("iam.NewService: %v", err)
	}

	request := &iam.CreateServiceAccountRequest{
		AccountId: name,
		ServiceAccount: &iam.ServiceAccount{
			DisplayName: displayName,
		},
	}
	account, err := service.Projects.ServiceAccounts.Create("projects/"+projectID, request).Do()
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("Projects.ServiceAccounts.Create: %v", err)
	}
	fmt.Fprintf(w, "Created service account: %v", account)
	return account, nil
}

// func main() {
// 	f := bufio.NewWriter(os.Stdout)
// 	defer f.Flush()
// 	// account, _ := createServiceAccount(f, "golang-370407", "first-service", "First Service Account")
// 	// fmt.Println(account)
// 	ctx := context.Background()

//		client := createClient(ctx)
//		_, _, err := client.Collection("users").Add(ctx, map[string]interface{}{
//			"first": "Ada",
//			"last":  "Lovelace",
//			"born":  1815,
//		})
//		if err != nil {
//			log.Fatalf("Failed adding alovelace: %v", err)
//		}
//	}
type Entity struct {
	Value string
}

func main() {
	ctx := context.Background()

	// Create a datastore client. In a typical application, you would create
	// a single client which is reused for every datastore operation.
	dsClient, err := datastore.NewClient(ctx, "golang-370407")
	if err != nil {
		// Handle error.
	}
	defer dsClient.Close()

	k := datastore.NameKey("Entity", "stringID", nil)
	e := new(Entity)
	if err := dsClient.Get(ctx, k, e); err != nil {
		// Handle error.
	}

	old := e.Value
	e.Value = "Hello World!"

	if _, err := dsClient.Put(ctx, k, e); err != nil {
		// Handle error.
	}

	fmt.Printf("Updated value from %q to %q\n", old, e.Value)
}
