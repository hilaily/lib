package mysql

import (
	"fmt"
	"strings"
	"time"

	"github.com/hilaily/lib/configx"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type IMySQL interface {
	GetWriteDB() *gorm.DB
	GetReadDB() *gorm.DB
}

type DBConfig struct {
	User     string
	Pass     string
	Host     string
	Port     int
	Database string
}

func NewFromConfig(conf configx.IConfig) (*_mysql, error) {
	var writeConf *DBConfig
	ok, err := conf.Get("mysql_write", writeConf)
	if !ok {
		return nil, fmt.Errorf("mysql write config not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get mysql write config failed: %w", err)
	}

	var readConf *DBConfig
	hasRead, err := conf.Get("mysql_read", readConf)
	if hasRead {
		if err != nil {
			return nil, fmt.Errorf("get mysql read config failed: %w", err)
		}
	} else {
		logrus.Warn("mysql read config not found")
	}
	return New(writeConf, readConf)
}

func New(writeConf, readConf *DBConfig) (*_mysql, error) {
	writeDB, writeDSN, err := connectDB(writeConf.User, writeConf.Pass, writeConf.Host, writeConf.Port, writeConf.Database)
	if err != nil {
		return nil, fmt.Errorf("write db connect failed:%w", err)
	}
	res := &_mysql{
		writeDB: writeDB,
	}
	if readConf != nil {
		// 读库配置
		readDB, readDSN, err := connectDB(readConf.User, readConf.Pass, readConf.Host, readConf.Port, readConf.Database)
		if err != nil {
			return nil, fmt.Errorf("read db connect failed:%w", err)
		}
		err = writeDB.Use(dbresolver.Register(dbresolver.Config{
			Sources:  []gorm.Dialector{mysql.Open(writeDSN)}, // 写操作使用的数据源
			Replicas: []gorm.Dialector{mysql.Open(readDSN)},  // 读操作使用的数据源
			Policy:   dbresolver.RandomPolicy{},              // 负载均衡策略
		}).SetConnMaxIdleTime(time.Hour).
			SetConnMaxLifetime(24 * time.Hour).
			SetMaxIdleConns(100).
			SetMaxOpenConns(200))
		if err != nil {
			panic("can not config read write split: " + err.Error())
		}
		res.readDB = readDB
	}

	logrus.Info("database connect success")
	return res, nil
}

func connectDB(user, pass, host string, port int, dbName string) (*gorm.DB, string, error) {
	// 主库（写库）配置
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s", user, pass, host, port, dbName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err == nil {
		return db, dsn, nil
	}

	if !strings.Contains(err.Error(), "Unknown database") {
		return nil, "", fmt.Errorf("connect to database failed: %w", err)
	}

	tmpDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s", user, pass, host, port, "mysql")

	db, err = gorm.Open(mysql.Open(tmpDSN), &gorm.Config{})
	if err != nil {
		return nil, "", fmt.Errorf("connect to database failed: %w", err)
	}

	// 创建数据库
	err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s DEFAULT CHARACTER SET utf8mb4", dbName)).Error
	if err != nil {
		return nil, "", fmt.Errorf("create database failed: %w", err)
	}

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	return db, dsn, err
}

type _mysql struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func (m *_mysql) GetWriteDB() *gorm.DB {
	return m.writeDB
}

func (m *_mysql) GetReadDB() *gorm.DB {
	return m.readDB
}
