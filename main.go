package main

import (
	"strconv"

	dm "github.com/baetyl/baetyl-go/v2/dmcontext"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/jinzhu/copier"

	"github.com/baetyl/baetyl-bacnet/baetyl-bacnet"
)

func main() {
	dm.Run(func(ctx dm.Context) error {
		cfg, err := genConfig(ctx)
		if err != nil {
			return err
		}
		bac, err := baetyl_bacnet.NewBacnet(ctx, cfg)
		if err != nil {
			return err
		}
		defer bac.Close()
		ctx.Wait()
		return nil
	})
}

func genConfig(ctx dm.Context) (*baetyl_bacnet.Config, error) {
	cfg := &baetyl_bacnet.Config{}
	var slaves []baetyl_bacnet.SlaveConfig
	var jobs []baetyl_bacnet.Job

	// generate job
	for _, devInfo := range ctx.GetAllDevices() {
		accessConfig := devInfo.AccessConfig
		if accessConfig.Bacnet == nil {
			continue
		}
		slave := baetyl_bacnet.SlaveConfig{
			Device: devInfo.Name,
		}
		if err := copier.Copy(&slave, accessConfig.Bacnet); err != nil {
			return nil, err
		}
		slaves = append(slaves, slave)

		// generate jobMap
		jobMaps := make(map[string]baetyl_bacnet.Property)
		devTpl, err := ctx.GetAccessTemplates(&devInfo)
		if err != nil {
			return nil, err
		}
		if devTpl != nil && devTpl.Properties != nil && len(devTpl.Properties) > 0 {
			for _, prop := range devTpl.Properties {
				if visitor := prop.Visitor.Bacnet; visitor != nil {
					var jobMap baetyl_bacnet.Property
					jobMap.Id = prop.Id
					jobMap.Name = prop.Name
					jobMap.Type = prop.Type
					jobMap.Mode = prop.Mode
					jobMap.BacnetType = visitor.BacnetType
					jobMap.ApplicationTagNumber = visitor.ApplicationTagNumber
					jobMap.BacnetAddress = visitor.BacnetAddress + accessConfig.Bacnet.AddressOffset
					jobMaps[strconv.FormatUint(uint64(jobMap.BacnetType), 10)+":"+
						strconv.FormatUint(uint64(jobMap.BacnetAddress), 10)] = jobMap
				}
			}
		}
		job := baetyl_bacnet.Job{
			Device:        devInfo.Name,
			Interval:      accessConfig.Bacnet.Interval,
			Properties:    jobMaps,
			DeviceId:      accessConfig.Bacnet.DeviceId,
			AddressOffset: accessConfig.Bacnet.AddressOffset,
		}
		jobs = append(jobs, job)
	}
	cfg.Jobs = jobs
	cfg.Slaves = slaves
	if err := utils.SetDefaults(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
