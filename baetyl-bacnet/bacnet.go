package baetyl_bacnet

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/baetyl/baetyl-bacnet/bacip"
	"github.com/baetyl/baetyl-bacnet/bacnet"
	"github.com/baetyl/baetyl-bacnet/dmp"
	dm "github.com/baetyl/baetyl-go/v2/dmcontext"
	v2log "github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/spec/v1"
)

type Bacnet struct {
	ctx dm.Context
	log *v2log.Logger
	cfg *Config
	ws  map[string]*Worker
}

func NewBacnet(ctx dm.Context, cfg *Config) (*Bacnet, error) {
	log := ctx.Log().With(v2log.Any("module", "baetyl-bacnet"))
	infos := make(map[string]dm.DeviceInfo)
	for _, info := range ctx.GetAllDevices() {
		infos[info.Name] = info
	}
	slaves := make(map[string]*Slave)
	for _, dCfg := range cfg.Slaves {
		if info, ok := infos[dCfg.Device]; ok {
			dev, err := NewSlave(&info, dCfg)
			if err != nil {
				log.Error("failed to create device instance", v2log.Any("device", dCfg.Device), v2log.Error(err))
				continue
			}
			ctx.Online(&info)
			slaves[dCfg.Device] = dev
		}
	}
	ws := make(map[string]*Worker)
	for _, job := range cfg.Jobs {
		if dev := slaves[job.Device]; dev != nil {
			ws[job.Device] = NewWorker(job, dev, ctx, log)
		} else {
			log.Error("device of job not exist", v2log.Any("device id", job.Device))
		}
	}
	bac := &Bacnet{
		ctx: ctx,
		log: log,
		cfg: cfg,
		ws:  ws,
	}
	if err := ctx.RegisterDeltaCallback(bac.DeltaCallback); err != nil {
		return nil, err
	}
	if err := ctx.RegisterPropertyGetCallback(bac.PropertyGetCallback); err != nil {
		return nil, err
	}
	for _, worker := range ws {
		go bac.working(worker)
	}
	return bac, nil
}

func (bac *Bacnet) working(w *Worker) {
	ticker := time.NewTicker(w.job.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := w.Execute()
			if err != nil {
				bac.log.Error("failed to execute job", v2log.Error(err))
			}
		case <-bac.ctx.WaitChan():
			bac.log.Warn("worker stopped", v2log.Any("worker", w))
			return
		}
	}
}

func (bac *Bacnet) Close() error {
	return nil
}

func (bac *Bacnet) DeltaCallback(info *dm.DeviceInfo, delta v1.Delta) error {
	w, ok := bac.ws[info.Name]
	if !ok {
		bac.log.Warn("worker not exist according to device", v2log.Any("device", info.Name))
		return errors.New("worker not exist")
	}
	accessTemplate, err := w.ctx.GetAccessTemplates(info)
	if err != nil {
		bac.log.Warn("get access template err", v2log.Any("device", info.Name))
		return err
	}
	for key, val := range delta {
		id, err := dmp.GetConfigIdByModelName(key, accessTemplate)
		if id == "" || err != nil {
			bac.log.Warn("prop not exist", v2log.Any("name", key))
			continue
		}
		propName, err := dmp.GetMappingName(id, accessTemplate)
		if err != nil {
			bac.log.Warn("prop name not exist", v2log.Any("id", id))
			continue
		}
		propVal, err := dmp.GetPropValueByModelName(key, val, accessTemplate)
		println(propVal)
		if err != nil {
			bac.log.Warn("get prop value err", v2log.Any("name", propName))
			continue
		}

		for _, prop := range w.job.Properties {
			if propName == prop.Name {
				var value interface{}
				switch propVal.(type) {
				case float64:
					value, err = dmp.ParsePropertyValue(prop.Type, propVal.(float64))
				default:
					value = propVal
				}
				if err != nil {
					return err
				}

				objID := bacnet.ObjectID{
					Type:     bacnet.ObjectType(prop.BacnetType),
					Instance: bacnet.ObjectInstance(prop.BacnetAddress),
				}
				err = writeValue(w.slave.bacnetClient, w.slave.device, objID, bacnet.PropertyValue{
					Type:  bacnet.TypeBoolean,
					Value: true,
				}, bacnet.PropertyIdentifier{
					Type: bacnet.OutOfService,
				})
				if err != nil {
					return err
				}
				var ty bacnet.PropertyValueType
				switch value.(type) {
				case bool:
					ty = bacnet.TypeBoolean
				default:
					ty = bacnet.TypeReal
				}
				if err != nil {
					return err
				}
				err = writeValue(w.slave.bacnetClient, w.slave.device, objID, bacnet.PropertyValue{
					Type:  ty,
					Value: value,
				}, bacnet.PropertyIdentifier{
					Type: bacnet.PresentValue,
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func writeValue(c *bacip.Client, device bacnet.Device, object bacnet.ObjectID, propertyValue bacnet.PropertyValue,
	property bacnet.PropertyIdentifier) error {
	wp := bacip.WriteProperty{
		ObjectID:      object,
		Property:      property,
		PropertyValue: propertyValue,
		Priority:      bacnet.Available16,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := c.WriteProperty(ctx, device, wp)
	if err != nil {
		fmt.Printf("%v\t", err)
		return err
	}
	return nil
}

func (bac *Bacnet) PropertyGetCallback(info *dm.DeviceInfo, properties []string) error {
	w, ok := bac.ws[info.Name]
	if !ok {
		bac.log.Warn("worker not exist according to device", v2log.Any("device", info.Name))
		return errors.New("worker not exist")
	}
	if err := w.Execute(); err != nil {
		return err
	}
	return nil
}
