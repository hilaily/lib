package oss

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hilaily/kit/stringx"
	"github.com/sirupsen/logrus"
)

var (
	_ IS3 = &S3Client{}
)

type IS3 interface {
	UploadFile(ctx context.Context, key string, data []byte) error
	GenURL(ctx context.Context, key string) string
	GetFileSize(ctx context.Context, key string) (int64, error)
	GeneratePresignedURL(ctx context.Context, filename string, expires time.Duration) (*PreSignedInfo, error)
	UploadFileWithPresignedURL(method, url string, data io.Reader) error
	GetClient() *s3.Client
}

type S3Config struct {
	Endpoint   string `json:"endpoint" yaml:"endpoint" mapstructure:"endpoint"`
	AccessKey  string `json:"access_key" yaml:"access_key" mapstructure:"access_key"`
	SecretKey  string `json:"secret_key" yaml:"secret_key" mapstructure:"secret_key"`
	Region     string `json:"region" yaml:"region" mapstructure:"region"`
	BucketName string `json:"bucket_name" yaml:"bucket_name" mapstructure:"bucket_name"`
	Domain     string `json:"domain" yaml:"domain" mapstructure:"domain"`
}

type IConfig interface {
	Unmarshal(ptr any) error
}

func NewS3ClientFromConfig(conf IConfig) (IS3, error) {
	var config S3Config
	err := conf.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("get s3 config failed: %w", err)
	}
	return NewS3Client(&config)
}

// NewS3Client 初始化并返回一个 S3Client 实例
func NewS3Client(conf *S3Config) (*S3Client, error) {
	endpoint := conf.Endpoint
	accessKey := conf.AccessKey
	secretKey := conf.SecretKey
	region := conf.Region
	logrus.Debugf("oss config: endpoint: %s, accessKey: %s, region: %s", endpoint, accessKey, region)

	// 加载 AWS 配置，包括凭证和自定义端点
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion(region),
		config.WithEndpointResolver(
			aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
				if service == s3.ServiceID && endpoint != "" {
					return aws.Endpoint{
						URL:           endpoint,
						SigningRegion: region, // 使用配置中的区域
					}, nil
				}
				// 使用默认解析器
				return aws.Endpoint{}, &aws.EndpointNotFoundError{}
			}),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("can not load aws config: %w", err)
	}

	// 创建 S3 客户端
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &S3Client{
		Client: client,
		cfg:    conf,
	}, nil
}

// S3Client 定义了支持上传和获取文件大小的结构体
type S3Client struct {
	cfg    *S3Config
	Client *s3.Client
}

func (s *S3Client) GetClient() *s3.Client {
	return s.Client
}

// UploadFile 上传文件到指定的 S3 存储桶
// key 是文件在 S3 中的路径，data 是文件的内容
func (s *S3Client) UploadFile(ctx context.Context, key string, data []byte) error {
	// 创建上传输入参数
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.cfg.BucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	}

	// 执行上传操作
	_, err := s.Client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("upload file failed: %w", err)
	}

	return nil
}

// GetFileSize 获取指定文件的大小（以字节为单位）
// key 是文件在 S3 中的路径
func (s *S3Client) GetFileSize(ctx context.Context, key string) (int64, error) {
	// 创建 HeadObject 输入参数
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.cfg.BucketName),
		Key:    aws.String(key),
	}

	// 执行 HeadObject 操作
	output, err := s.Client.HeadObject(ctx, input)
	if err != nil {
		var notFound *types.NotFound
		if ok := errors.As(err, &notFound); ok {
			return 0, fmt.Errorf("file %s not found", key)
		}
		return 0, fmt.Errorf("get file size failed: %w", err)
	}

	return *output.ContentLength, nil
}

// GeneratePresignedURL 生成预签名 URL
func (s *S3Client) GeneratePresignedURL(ctx context.Context, filename string, expires time.Duration) (*PreSignedInfo, error) {
	presignClient := s3.NewPresignClient(s.Client)

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.cfg.BucketName),
		Key:    aws.String(filename),
	}

	presignedRequest, err := presignClient.PresignPutObject(ctx, input, s3.WithPresignExpires(expires))
	if err != nil {
		return nil, fmt.Errorf("generate presigned url failed: %w", err)
	}

	header := make(map[string]string, len(presignedRequest.SignedHeader))
	for k, v := range presignedRequest.SignedHeader {
		header[k] = v[0]
	}

	return &PreSignedInfo{
		URL:    presignedRequest.URL,
		Header: header,
		Method: presignedRequest.Method,
	}, nil
}

// UploadFileWithPresignedURL 根据预签名的 URL 上传文件
func (s *S3Client) UploadFileWithPresignedURL(method, url string, data io.Reader) error {
	// 创建 HTTP PUT 请求
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return fmt.Errorf("create upload request failed: %w", err)
	}

	// 设置必要的头部信息，根据需要设置 Content-Type
	req.Header.Set("Content-Type", "application/octet-stream")

	// 初始化 HTTP 客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 执行请求
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("execute upload request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体（可选）
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response failed: %w", err)
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("upload failed, status code: %d, response: %s", resp.StatusCode, string(body))
	}

	log.Printf("file uploaded successfully through presigned url")
	return nil
}

func (s *S3Client) GenURL(ctx context.Context, key string) string {
	domain := s.cfg.Domain
	if domain == "" {
		domain = s.cfg.Endpoint
	}
	return stringx.URLJoin(domain, s.cfg.BucketName, key)
}

type PreSignedInfo struct {
	URL    string
	Header map[string]string
	Method string
}
