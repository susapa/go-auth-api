package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
)

type BlobClient struct {
	client    *azblob.Client
	container string
	accountName string
}

var defaultClient *BlobClient

// Init initialises the Azure Blob Storage client.
// connStr is a storage account connection string.
// container is the blob container name.
func Init(connStr, container, accountName string) error {
	client, err := azblob.NewClientFromConnectionString(connStr, nil)
	if err != nil {
		return fmt.Errorf("blob: NewClientFromConnectionString: %w", err)
	}
	defaultClient = &BlobClient{
		client:      client,
		container:   container,
		accountName: accountName,
	}
	return nil
}

// Upload streams r to Azure Blob Storage under the given blobName.
// Returns the public URL of the uploaded blob.
func Upload(ctx context.Context, blobName string, r io.Reader, contentType string) (string, error) {
	if defaultClient == nil {
		return "", fmt.Errorf("blob: client not initialised — call storage.Init first")
	}

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	opts := &azblob.UploadStreamOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: &contentType,
		},
	}

	_, err := defaultClient.client.UploadStream(ctx, defaultClient.container, blobName, r, opts)
	if err != nil {
		return "", fmt.Errorf("blob: UploadStream: %w", err)
	}

	url := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s",
		defaultClient.accountName, defaultClient.container, blobName)
	return url, nil
}
