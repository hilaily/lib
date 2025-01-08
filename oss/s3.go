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
	"github.com/sirupsen/logrus"
)

// S3Client 定义了支持上传和获取文件大小的结构体
type S3Client struct {
	Client     *s3.Client
	BucketName string
}

type S3Config struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	Region     string
	BucketName string
}

// NewS3Client 初始化并返回一个 S3Client 实例
func NewS3Client(conf *S3Config) (*S3Client, error) {
	endpoint := conf.Endpoint
	accessKey := conf.AccessKey
	secretKey := conf.SecretKey
	region := conf.Region
	logrus.Infof("oss config: endpoint %s, accessKey %s, secretKey(len) %d, region %s", endpoint, accessKey, len(secretKey), region)

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
		return nil, fmt.Errorf("无法加载 AWS 配置: %w", err)
	}

	// 创建 S3 客户端
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &S3Client{
		Client:     client,
		BucketName: conf.BucketName,
	}, nil
}

// UploadFile 上传文件到指定的 S3 存储桶
// key 是文件在 S3 中的路径，data 是文件的内容
func (s *S3Client) UploadFile(ctx context.Context, key string, data []byte) error {
	// 创建上传输入参数
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	}

	// 执行上传操作
	_, err := s.Client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("上传文件失败: %w", err)
	}

	return nil
}

// GetFileSize 获取指定文件的大小（以字节为单位）
// key 是文件在 S3 中的路径
func (s *S3Client) GetFileSize(ctx context.Context, key string) (int64, error) {
	// 创建 HeadObject 输入参数
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(key),
	}

	// 执行 HeadObject 操作
	output, err := s.Client.HeadObject(ctx, input)
	if err != nil {
		var notFound *types.NotFound
		if ok := errors.As(err, &notFound); ok {
			return 0, fmt.Errorf("文件 %s 未找到", key)
		}
		return 0, fmt.Errorf("获取文件大小失败: %w", err)
	}

	return *output.ContentLength, nil
}

// GeneratePresignedURL 生成预签名 URL
func (s *S3Client) GeneratePresignedURL(ctx context.Context, filename string, expires time.Duration) (*PreSignedInfo, error) {
	presignClient := s3.NewPresignClient(s.Client)

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(filename),
	}

	presignedRequest, err := presignClient.PresignPutObject(ctx, input, s3.WithPresignExpires(expires))
	if err != nil {
		return nil, fmt.Errorf("生成预签名 URL 失败: %w", err)
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
		return fmt.Errorf("创建上传请求失败: %w", err)
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
		return fmt.Errorf("执行上传请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体（可选）
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("上传失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	log.Printf("文件通过预签名 URL 上传成功")
	return nil
}

type PreSignedInfo struct {
	URL    string
	Header map[string]string
	Method string
}
