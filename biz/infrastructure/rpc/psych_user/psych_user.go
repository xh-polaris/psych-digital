package psych_user

import (
	"github.com/google/wire"
	"github.com/xh-polaris/gopkg/kitex/client"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/config"
	"github.com/xh-polaris/psych-idl/kitex_gen/user/psychuserservice"
)

type IPsychUser interface {
	psychuserservice.Client
}

type PsychUser struct {
	psychuserservice.Client
}

var PsychUserSet = wire.NewSet(
	NewPsychUser,
	wire.Struct(new(PsychUser), "*"),
	wire.Bind(new(IPsychUser), new(*PsychUser)),
)

func NewPsychUser(config *config.Config) psychuserservice.Client {
	return client.NewClient(config.Name, "psych.user", psychuserservice.NewClient)
}
