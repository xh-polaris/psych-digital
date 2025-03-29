package history

import (
	"github.com/xh-polaris/psych-digital/biz/infrastructure/config"
	"github.com/zeromicro/go-zero/core/stores/monc"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/net/context"
	"sync"
)

const (
	prefixUserCacheKey = "cache:history"
	CollectionName     = "history"
)

var Mapper *MongoMapper
var once sync.Once

type IMongoMapper interface {
	Insert(ctx context.Context, his History) error
}

type MongoMapper struct {
	conn *monc.Model
}

func GetMongoMapper() *MongoMapper {
	once.Do(func() {
		c := config.GetConfig()
		conn := monc.MustNewModel(c.Mongo.URL, c.Mongo.DB, CollectionName, c.Cache)
		Mapper = &MongoMapper{
			conn: conn,
		}
	})
	return Mapper
}

func (m *MongoMapper) Insert(ctx context.Context, his *History) error {
	if his.ID.IsZero() {
		his.ID = primitive.NewObjectID()
	}
	_, err := m.conn.InsertOneNoCache(ctx, his)
	return err
}
