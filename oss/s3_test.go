package oss

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestOss(t *testing.T) {
	conf := &S3Config{
		Endpoint:   "",
		AccessKey:  "",
		SecretKey:  "",
		Region:     "",
		BucketName: "",
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
