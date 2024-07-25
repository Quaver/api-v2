package azure

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Quaver/api2/config"
	"github.com/sirupsen/logrus"
	"io"
	"io/fs"
	"io/ioutil"
	"net/url"
	"os"
)

type StorageClient struct {
	accountName string
	accountKey  string
	credential  azblob.Credential
	pipe        pipeline.Pipeline
}

// Client Global storage client used throughout the application.
var Client StorageClient

// InitializeClient  Initializes the azure storage client
func InitializeClient() {
	if Client != (StorageClient{}) {
		return
	}

	var err error

	client := StorageClient{
		accountName: config.Instance.Azure.AccountName,
		accountKey:  config.Instance.Azure.AccountKey,
	}

	client.credential, err = azblob.NewSharedKeyCredential(client.accountName, client.accountKey)

	if err != nil {
		panic(err)
	}

	client.pipe = azblob.NewPipeline(client.credential, azblob.PipelineOptions{})
	Client = client

	logrus.Info("Successfully created Azure client!")
}

// Create a ContainerURL object to be able to make requests on that container
func (c *StorageClient) createContainerURL(container string) azblob.ContainerURL {
	URL, _ := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/%s", c.accountName, container))
	return azblob.NewContainerURL(*URL, c.pipe)
}

// UploadFile Uploads a file to a given container
func (c *StorageClient) UploadFile(container string, fileName string, data []byte) error {
	containerURL := c.createContainerURL(container)
	blobURL := containerURL.NewBlockBlobURL(fileName)
	ctx := context.Background()

	_, err := azblob.UploadBufferToBlockBlob(ctx, data, blobURL, azblob.UploadToBlockBlobOptions{
		BlockSize:   4 * 1024 * 1024,
		Parallelism: 16,
	})

	if err != nil {
		return err
	}

	return nil
}

// UploadFileFromDisk Uploads a file to azure from disk
func (c *StorageClient) UploadFileFromDisk(container string, name string, path string, progress pipeline.ProgressReceiver) error {
	containerURL := c.createContainerURL(container)
	blobURL := containerURL.NewBlockBlobURL(name)
	ctx := context.WithoutCancel(context.Background())

	file, err := os.Open(path)

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = azblob.UploadFileToBlockBlob(ctx, file, blobURL, azblob.UploadToBlockBlobOptions{
		BlockSize:   4 * 1024 * 1024,
		Parallelism: 16,
		Progress:    progress,
	})

	if err != nil {
		return err
	}

	return nil
}

// DownloadFile Downloads a blob from a given container
func (c *StorageClient) DownloadFile(container string, name string, path string) (bytes.Buffer, error) {
	containerURL := c.createContainerURL(container)
	blobURL := containerURL.NewBlockBlobURL(name)
	ctx := context.Background()

	resp, err := blobURL.Download(ctx, 0, azblob.CountToEnd, azblob.BlobAccessConditions{},
		false, azblob.ClientProvidedKeyOptions{})

	if err != nil {
		return bytes.Buffer{}, err
	}

	bodyStream := resp.Body(azblob.RetryReaderOptions{MaxRetryRequests: 20})

	defer func(bodyStream io.ReadCloser) {
		err := bodyStream.Close()
		if err != nil {
			return
		}
	}(bodyStream)

	// Read the body into a buffer
	downloadedData := bytes.Buffer{}
	_, err = downloadedData.ReadFrom(bodyStream)

	if err != nil {
		return bytes.Buffer{}, err
	}

	err = ioutil.WriteFile(path, downloadedData.Bytes(), fs.ModePerm)

	if err != nil {
		return bytes.Buffer{}, err
	}

	return downloadedData, nil
}

// ListBlobs Lists blobs in a given container
func (c *StorageClient) ListBlobs(container string) ([]string, error) {
	containerURL := c.createContainerURL(container)
	ctx := context.Background()

	blobs := make([]string, 0)

	for marker := (azblob.Marker{}); marker.NotDone(); {
		listBlob, err := containerURL.ListBlobsFlatSegment(ctx, marker, azblob.ListBlobsSegmentOptions{})

		if err != nil {
			return nil, err
		}

		marker = listBlob.NextMarker

		// Process the blobs returned in this result segment (if the segment is empty, the loop body won't execute)
		for _, blobInfo := range listBlob.Segment.BlobItems {
			blobs = append(blobs, blobInfo.Name)
		}
	}

	return blobs, nil
}

// DeleteBlob Delete a blob from a given container
func (c *StorageClient) DeleteBlob(container string, fileName string) error {
	containerURL := c.createContainerURL(container)
	blobURL := containerURL.NewBlockBlobURL(fileName)
	ctx := context.Background()

	_, err := blobURL.Delete(ctx, azblob.DeleteSnapshotsOptionNone, azblob.BlobAccessConditions{})

	if err != nil {
		return nil
	}

	return nil
}
