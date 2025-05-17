package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewWrapper[T any](client *mongo.Client, dbName string) *Wrapper[T] {
	return &Wrapper[T]{
		Client: client,
		DBName: dbName,
	}
}

type Wrapper[T any] struct {
	Client *mongo.Client
	DBName string
}

// InsertOne 插入单个新文档
func (w *Wrapper[T]) InsertOne(collectionName string, data T) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := w.Client.Database(w.DBName).Collection(collectionName)

	return collection.InsertOne(ctx, data)
}

// UpdateOne 更新单个文档
func (w *Wrapper[T]) UpdateOne(collectionName string, filter bson.M, data T) (*mongo.UpdateResult, error) {
	if len(filter) == 0 {
		return nil, fmt.Errorf("更新操作必须提供有效的过滤条件")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := w.Client.Database(w.DBName).Collection(collectionName)

	opts := options.Replace().SetUpsert(false) // 默认不启用 upsert
	return collection.ReplaceOne(ctx, filter, data, opts)
}

// UpsertOne 更新文档，如果不存在则插入
func (w *Wrapper[T]) UpsertOne(collectionName string, filter bson.M, data T) (*mongo.UpdateResult, error) {
	if len(filter) == 0 {
		return nil, fmt.Errorf("upsert 操作必须提供有效的过滤条件")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := w.Client.Database(w.DBName).Collection(collectionName)

	opts := options.Replace().SetUpsert(true)
	return collection.ReplaceOne(ctx, filter, data, opts)
}

// FindOne 查询单个文档，支持 pipeline
func (w *Wrapper[T]) FindOne(collectionName string, filter bson.M, pipeline []bson.D) (*T, error) {
	var result *T
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := w.Client.Database(w.DBName).Collection(collectionName)

	// 如果提供了 pipeline，使用 Aggregate
	if len(pipeline) > 0 {
		// 如果有 filter，将其添加到 pipeline 开头
		if filter != nil {
			matchStage := bson.D{{Key: "$match", Value: filter}}
			pipeline = append([]bson.D{matchStage}, pipeline...)
		}

		// 添加 $limit 1 到 pipeline 末尾确保只返回一个文档
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: int64(1)}})

		cursor, err := collection.Aggregate(ctx, pipeline)
		if err != nil {
			return result, err
		}
		defer cursor.Close(ctx)

		if cursor.Next(ctx) {
			err = cursor.Decode(&result)
			return result, err
		}
		return result, mongo.ErrNoDocuments
	}

	// 没有 pipeline 时使用普通的 FindOne
	if filter == nil {
		filter = bson.M{}
	}
	err := collection.FindOne(ctx, filter).Decode(&result)
	return result, err
}

// FindMany 查询多个文档，支持 pipeline
func (w *Wrapper[T]) FindMany(collectionName string, filter bson.M, pipeline []bson.D) ([]*T, error) {
	var results []*T
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := w.Client.Database(w.DBName).Collection(collectionName)

	var cursor *mongo.Cursor
	var err error

	// 如果提供了 pipeline，使用 Aggregate
	if len(pipeline) > 0 {
		// 如果有 filter，将其添加到 pipeline 开头
		if filter != nil {
			matchStage := bson.D{{Key: "$match", Value: filter}}
			pipeline = append([]bson.D{matchStage}, pipeline...)
		}

		cursor, err = collection.Aggregate(ctx, pipeline)
		if err != nil {
			return nil, err
		}
		defer cursor.Close(ctx)
	} else {
		// 没有 pipeline 时使用普通的 Find
		if filter == nil {
			filter = bson.M{}
		}

		cursor, err = collection.Find(ctx, filter)
		if err != nil {
			return nil, err
		}
		defer cursor.Close(ctx)
	}

	logrus.Debugf("cursor: %v", cursor)
	err = cursor.All(ctx, &results)
	if err != nil {
		return nil, fmt.Errorf("cursor.All fail: %v\n", err)
	}

	return results, nil
}
