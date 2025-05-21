package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/hilaily/lib/configx"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBConfig struct {
	URI    string `yaml:"uri"`
	User   string `yaml:"user"`
	Pass   string `yaml:"pass"`
	Port   int    `yaml:"port"`
	DBName string `yaml:"db_name"`
}

func NewFromConfig(conf configx.IConfig) (*mongo.Client, error) {
	var config *MongoDBConfig
	err := conf.Get("mongo", config)
	if err != nil {
		return nil, fmt.Errorf("get mongo config failed: %w", err)
	}
	return NewClient(config)
}

func NewClient(conf *MongoDBConfig) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(conf.URI).
		SetAuth(options.Credential{
			Username: conf.User,
			Password: conf.Pass,
		}).
		SetMaxPoolSize(100).                 // 设置最大连接池大小
		SetMinPoolSize(20).                  // 设置最小连接池大小
		SetMaxConnIdleTime(time.Minute * 5). // 设置最大空闲时间
		SetTimeout(time.Second * 30)         // 设置操作超时时间

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to MongoDB: %v", err)
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to ping MongoDB: %v", err)
	}

	return client, nil
}
