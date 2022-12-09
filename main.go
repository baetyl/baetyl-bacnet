package main

import (
	"strconv"

	"github.com/baetyl/baetyl-bacnet/baetyl-bacnet"
	dm "github.com/baetyl/baetyl-go/v2/dmcontext"
	"github.com/baetyl/baetyl-go/v2/utils"
	"gopkg.in/yaml.v2"
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

	// generate slave
	var slave baetyl_bacnet.SlaveConfig
	if err := yaml.Unmarshal([]byte(ctx.GetDriverConfig()), &slave); err != nil {
		return nil, err
	}

	// generate job
	for _, devInfo := range ctx.GetAllDevices() {
		accessConfig := devInfo.AccessConfig
		if accessConfig.Custom == nil {
			continue
		}
		slave.Device = devInfo.Name
		var job baetyl_bacnet.Job
		if err := yaml.Unmarshal([]byte(*accessConfig.Custom), &job); err != nil {
			return nil, err
		}
		job.Interval = slave.Interval
		job.Device = slave.Device

		// generate jobMap
		jobMaps := make(map[string]baetyl_bacnet.Property)
		devTpl, err := ctx.GetAccessTemplates(&devInfo)
		if err != nil {
			return nil, err
		}
		if devTpl != nil && devTpl.Properties != nil && len(devTpl.Properties) > 0 {
			for _, prop := range devTpl.Properties {
				if visitor := prop.Visitor.Custom; visitor != nil {
					var jobMap baetyl_bacnet.Property
					if err := yaml.Unmarshal([]byte(*visitor), &jobMap); err != nil {
						return nil, err
					}
					jobMap.Id = prop.Id
					jobMap.Name = prop.Name
					jobMap.Type = prop.Type
					jobMap.Mode = prop.Mode
					jobMap.BacnetAddress = jobMap.BacnetAddress + job.AddressOffset
					jobMaps[strconv.FormatUint(uint64(jobMap.BacnetType), 10)+":"+
						strconv.FormatUint(uint64(jobMap.BacnetAddress), 10)] = jobMap
				}
			}
		}
		slave.Device = devInfo.Name
		slave.DeviceId = job.DeviceId
		slaves = append(slaves, slave)
		job.Properties = jobMaps
		jobs = append(jobs, job)
	}
	cfg.Jobs = jobs
	cfg.Slaves = slaves
	if err := utils.SetDefaults(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
