package history

import (
	"github.com/xh-polaris/psych-digital/biz/adaptor/cmd"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/config"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/consts"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/util"
	"github.com/zeromicro/go-zero/core/stores/monc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	FindMany(ctx context.Context, p *cmd.Paging) (data []*History, total int64, err error)
}

type MongoMapper struct {
	conn *monc.Model
}

func NewMongoMapper(config *config.Config) *MongoMapper {
	conn := monc.MustNewModel(config.Mongo.URL, config.Mongo.DB, CollectionName, config.Cache)
	return &MongoMapper{conn: conn}
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

func (m *MongoMapper) FindMany(ctx context.Context, p *cmd.Paging) (data []*History, total int64, err error) {
	skip, limit := util.ParsePaging(p)
	data = make([]*History, 0, limit)
	err = m.conn.Find(ctx, &data,
		bson.M{}, &options.FindOptions{
			Skip:  &skip,
			Limit: &limit,
			Sort:  bson.M{consts.StartTime: -1},
		})
	if err != nil {
		return nil, 0, err
	}

	total, err = m.conn.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}
	return data, total, nil
}
