package provider

import (
	"github.com/google/wire"
	"github.com/xh-polaris/psych-digital/biz/application/service"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/config"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/mapper/history"
)

var provider *Provider

func Init() {
	var err error
	provider, err = NewProvider()
	if err != nil {
		panic(err)
	}
}

// Provider 提供controller依赖的对象
type Provider struct {
	Config         *config.Config
	HistoryService service.HistoryService
}

func Get() *Provider {
	return provider
}

var RpcSet = wire.NewSet()

var ApplicationSet = wire.NewSet(
	service.HistoryServiceSet,
)

var InfrastructureSet = wire.NewSet(
	config.NewConfig,
	history.NewMongoMapper,
	RpcSet,
)

var AllProvider = wire.NewSet(
	ApplicationSet,
	InfrastructureSet,
)
