package baetyl_bacnet

import (
	"context"
	"strconv"
	"time"

	"github.com/baetyl/baetyl-bacnet/bacip"
	"github.com/baetyl/baetyl-bacnet/bacnet"
	"github.com/baetyl/baetyl-bacnet/dmp"
	dm "github.com/baetyl/baetyl-go/v2/dmcontext"
	"github.com/baetyl/baetyl-go/v2/log"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
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
	prop := bacnet.PropertyIdentifier{Type: bacnet.ObjectList, ArrayIndex: new(uint32)}
	*prop.ArrayIndex = uint32(0)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	d, err := w.slave.bacnetClient.ReadProperty(ctx, w.slave.device, bacip.ReadProperty{
		ObjectID: w.slave.device.ID,
		Property: prop,
	})
	cancel()
	if err != nil {
		return err
	}
	temp := make(map[string]interface{})
	for i := 1; i <= int(d.(uint32)); i++ {
		*prop.ArrayIndex = uint32(i)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		d, err := w.slave.bacnetClient.ReadProperty(ctx, w.slave.device, bacip.ReadProperty{
			ObjectID: w.slave.device.ID,
			Property: prop,
		})
		cancel()
		if err != nil {
			w.log.Error("failed to read", log.Error(err))
			continue
		}
		objID := d.(bacnet.ObjectID)
		if objID.Type == bacnet.BacnetDevice {
			continue
		}
		d, err = readValue(w.slave.bacnetClient, w.slave.device, objID)
		if err != nil {
			w.log.Error("failed to read", log.Error(err))
			continue
		}
		if pro, ok := w.job.Properties[strconv.FormatUint(uint64(objID.Type),
			10)+":"+strconv.FormatUint(uint64(objID.Instance), 10)]; ok {
			temp[pro.Name] = d
		}
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
			return err
		}
		for _, param := range params {
			id := param[1:]
			mappingName, err := dmp.GetMappingName(id, accessTemplate)
			if err != nil {
				return err
			}
			args[param] = temp[mappingName]
		}
		modelValue, err := dm.ExecExpression(model.Expression, args, model.Type)
		if err != nil {
			return err
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
