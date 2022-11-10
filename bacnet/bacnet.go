package bacnet

import (
	dm "github.com/baetyl/baetyl-go/v2/dmcontext"
	v2log "github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/spec/v1"
)

type Bacnet struct {
	ctx dm.Context
	log *v2log.Logger
	ws  map[string]*Worker
}

func NewBacnet(ctx dm.Context, cfg Config) (*Bacnet, error) {
	// TODO
	log := ctx.Log().With(v2log.Any("module", "bacnet"))
	bac := &Bacnet{
		ctx: ctx,
		log: log,
	}
	if err := ctx.RegisterDeltaCallback(bac.DeltaCallback); err != nil {
		return nil, err
	}
	if err := ctx.RegisterPropertyGetCallback(bac.PropertyGetCallback); err != nil {
		return nil, err
	}
	return bac, nil
}

func (bac *Bacnet) Start() {
	// TODO
}

func (bac *Bacnet) Close() error {
	// TODO
	return nil
}

func (bac *Bacnet) DeltaCallback(info *dm.DeviceInfo, prop v1.Delta) error {
	// TODO
	return nil
}

func (bac *Bacnet) PropertyGetCallback(info *dm.DeviceInfo, properties []string) error {
	// TODO
	return nil
}
