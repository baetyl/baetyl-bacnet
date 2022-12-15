package baetyl_bacnet

import (
	"context"
	"time"

	dm "github.com/baetyl/baetyl-go/v2/dmcontext"
	"github.com/baetyl/baetyl-go/v2/log"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"

	"github.com/baetyl/baetyl-bacnet/bacip"
	"github.com/baetyl/baetyl-bacnet/bacnet"
	"github.com/baetyl/baetyl-bacnet/dmp"
)

type Worker struct {
	job   Job
	slave *Slave
	ctx   dm.Context
	log   *log.Logger
}

func NewWorker(job Job, slave *Slave, ctx dm.Context, log *log.Logger) *Worker {
	return &Worker{
		job:   job,
		slave: slave,
		ctx:   ctx,
		log:   log,
	}
}

func (w *Worker) Execute() error {
	temp := make(map[string]interface{})
	for _, prop := range w.job.Properties {
		objID := bacnet.ObjectID{
			Type:     bacnet.ObjectType(prop.BacnetType),
			Instance: bacnet.ObjectInstance(prop.BacnetAddress),
		}
		d, err := readValue(w.slave.bacnetClient, w.slave.device, objID)
		if err != nil {
			w.log.Error("failed to read", log.Error(err))
			continue
		}
		temp[prop.Name] = d
	}

	r := v1.Report{}
	accessTemplate, err := w.ctx.GetAccessTemplates(w.slave.info)
	if err != nil {
		return err
	}
	for _, model := range accessTemplate.Mappings {
		args := make(map[string]interface{})
		params, err := dm.ParseExpression(model.Expression)
		if err != nil {
			w.log.Warn("parse expression failed", log.Any("expression", model.Expression), log.Error(err))
			continue
		}
		for _, param := range params {
			id := param[1:]
			mappingName, err := dmp.GetMappingName(id, accessTemplate)
			if err != nil {
				w.log.Warn("get mapping failed", log.Any("id", id), log.Error(err))
				continue
			}
			value, ok := temp[mappingName]
			if !ok {
				w.log.Warn("mapping name not exist", log.Any("name", mappingName))
				continue
			}
			args[param] = value
		}
		modelValue, err := dm.ExecExpression(model.Expression, args, model.Type)
		if err != nil {
			w.log.Warn("exec expression failed", log.Any("expression", model.Expression),
				log.Any("args", args), log.Any("mapping type", model.Type), log.Error(err))
			continue
		}
		r[model.Attribute] = modelValue
	}

	if err := w.ctx.ReportDeviceProperties(w.slave.info, r); err != nil {
		return err
	}
	return nil
}

func readValue(c *bacip.Client, device bacnet.Device, object bacnet.ObjectID) (interface{}, error) {
	rp := bacip.ReadProperty{
		ObjectID: object,
		Property: bacnet.PropertyIdentifier{
			Type: bacnet.PresentValue,
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	d, err := c.ReadProperty(ctx, device, rp)
	if err != nil {
		return nil, err
	}
	return d, nil
}
