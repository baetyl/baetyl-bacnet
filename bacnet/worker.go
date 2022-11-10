package bacnet

import (
	dm "github.com/baetyl/baetyl-go/v2/dmcontext"
	"github.com/baetyl/baetyl-go/v2/log"
)

type Worker struct {
	ctx dm.Context
	log *log.Logger
}

func NewWorker(ctx dm.Context, log *log.Logger) *Worker {
	return &Worker{
		ctx: ctx,
		log: log,
	}
}
