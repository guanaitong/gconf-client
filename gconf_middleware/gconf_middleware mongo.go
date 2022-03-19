package gconf_middleware

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/guanaitong/gconf-go-client"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

const defaultMongoConfigKey = "mongo-config.json"

type MongoType int

const (
	MongoStandalone MongoType = iota
	MongoReplicaSet
	MongoShardCluster
)

type MongoClient struct {
	Context context.Context
	Client  *mongo.Client
	DBCon   *mongo.Database
	Session mongo.Session
}

type MongoConfig struct {
	Type       MongoType `json:"type"`
	URI        string    `json:"uri"`
	DBName     string    `json:"dbName"`
	Standalone struct{}  `json:"standalone"`
	ReplicaSet struct {
		ReplicaName string `json:"replicaName"`
	} `json:"replicaSet"`
	ShardCluster    struct{}      `json:"shardCluster"`
	MaxPoolSize     uint64        `json:"maxPoolSize"`
	MinPoolSize     uint64        `json:"minPoolSize"`
	MaxConnIdleTime time.Duration `json:"maxConnIdleTime"`
	SocketTimeout   time.Duration `json:"socketTimeout"`
}

func (mongoConfig *MongoConfig) NewClient() *MongoClient {
	ctx := context.Background()
	mc := &MongoClient{Context: ctx}

	if mongoConfig.MaxPoolSize == 0 {
		mongoConfig.MaxPoolSize = 10
	}
	if mongoConfig.MinPoolSize == 0 {
		mongoConfig.MinPoolSize = 1
	}
	if mongoConfig.MaxConnIdleTime == time.Second*0 {
		mongoConfig.MaxConnIdleTime = time.Second * 60
	}
	if mongoConfig.SocketTimeout == time.Second*0 {
		mongoConfig.SocketTimeout = time.Second * 60
	}

	conStr, err := connstring.Parse(mongoConfig.URI)
	if err != nil {
		panic(err)
	}
	if mongoConfig.DBName == "" {
		mongoConfig.DBName = conStr.Database
	}

	//parse
	if mongoConfig.Type == MongoStandalone {

	} else if mongoConfig.Type == MongoReplicaSet {
		if mongoConfig.ReplicaSet.ReplicaName == "" {
			mongoConfig.ReplicaSet.ReplicaName = conStr.ReplicaSet
		}
	} else if mongoConfig.Type == MongoShardCluster {

	} else {
		panic("unsupported type")
	}

	clientOption := options.Client().
		ApplyURI(mongoConfig.URI).
		SetMaxPoolSize(mongoConfig.MaxPoolSize).
		SetMinPoolSize(mongoConfig.MinPoolSize).
		SetMaxConnIdleTime(mongoConfig.MaxConnIdleTime).
		SetSocketTimeout(mongoConfig.SocketTimeout)

	client, err := mongo.Connect(ctx, clientOption)
	if err != nil {
		panic(err)
	}
	mc.Client = client
	mc.DBCon = client.Database(conStr.Database)

	return mc
}

func GetDefaultMongoConfig() *MongoConfig {
	return GetMongoConfig(defaultMongoConfigKey)
}

func GetMongoConfig(key string) *MongoConfig {
	if key == "" {
		panic(errors.New("mongo config is null"))
	}

	mongoConfig := new(MongoConfig)
	configValue := gconf.GetCurrentConfigCollection().GetValue(key).Raw()
	if err := json.Unmarshal([]byte(configValue), mongoConfig); err != nil {
		panic(err)
	}
	return mongoConfig
}
