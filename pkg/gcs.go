package pkg

import (
	"context"
	"errors"
	"io"
	"cloud.google.com/go/storage"
)

type GcsClient struct {
	Client     *storage.Client
	BucketName string
}

func NewGcsClient(client *storage.Client, bucketName string) *GcsClient {
	return &GcsClient{
		Client:     client,
		BucketName: bucketName,
	}
}

func (g *GcsClient) UploadResume(ctx context.Context, file io.Reader, objectName string) (string, error) {
	if g.Client == nil {
		return "", errors.New("GCS client is not initialized")
	}

	writer := g.Client.Bucket(g.BucketName).Object(objectName).NewWriter(ctx)
	defer writer.Close()

	if _, err := io.Copy(writer, file); err != nil {
		return "", err
	}

	return "gs://" + g.BucketName + "/" + objectName, nil
}

