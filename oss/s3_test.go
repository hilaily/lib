package oss

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hilaily/lib/env"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestOss(t *testing.T) {
	env.MustLoad(".env.test")
	conf := &S3Config{
		Endpoint:   os.Getenv("S3_ENDPOINT"),
		AccessKey:  os.Getenv("S3_ACCESS_KEY"),
		SecretKey:  os.Getenv("S3_SECRET_KEY"),
		Region:     os.Getenv("S3_REGION"),
		BucketName: os.Getenv("S3_BUCKET_NAME"),
	}

	client, err := NewS3Client(conf)
	if err != nil {
		t.Fatal(err)
	}

	urlInfo, err := client.GeneratePresignedURL(context.Background(), "test.jpg", 3600*time.Second)
	assert.NoError(t, err)
	logrus.Info(urlInfo)

	data, err := os.ReadFile(filepath.Join(os.Getenv("HOME"), "Downloads/test.jpg"))
	assert.NoError(t, err)
	if len(data) == 0 {
		t.Fatal("data is empty")
	}
	err = client.UploadFileWithPresignedURL(urlInfo.Method, urlInfo.URL, bytes.NewReader(data))
	assert.NoError(t, err)
}
