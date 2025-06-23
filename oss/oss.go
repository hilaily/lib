package oss

type OSSClient = S3Client
type OSSConfig = S3Config

func New(conf *OSSConfig) (*OSSClient, error) {
	return NewS3Client(conf)
}

func NewFromConfig(conf IConfig) (*OSSClient, error) {
	v, err := NewS3ClientFromConfig(conf)
	if err != nil {
		return nil, err
	}
	return v.(*OSSClient), nil
}
