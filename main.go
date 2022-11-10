package main

import (
	dm "github.com/baetyl/baetyl-go/v2/dmcontext"
	"github.com/baetyl/baetyl-go/v2/utils"
	"gopkg.in/yaml.v2"

	"github.com/baetyl/baetyl-bacnet/bacnet"
)

func main() {
	dm.Run(func(ctx dm.Context) error {
		cfg, err := genConfig(ctx)
		if err != nil {
			return err
		}
		bac, err := bacnet.NewBacnet(ctx, *cfg)
		if err != nil {
			return err
		}
		bac.Start()
		defer bac.Close()
		ctx.Wait()
		return nil
	})
}

func genConfig(ctx dm.Context) (*bacnet.Config, error) {
	cfg := &bacnet.Config{}
	var slaves []bacnet.SlaveConfig
	var jobs []bacnet.Job

	// generate slave
	var slave bacnet.SlaveConfig
	if err := yaml.Unmarshal([]byte(ctx.GetDriverConfig()), &slave); err != nil {
		return nil, err
	}
	slaves = append(slaves, slave)

	// generate job
	for _, devInfo := range ctx.GetAllDevices() {
		accessConfig := devInfo.AccessConfig
		if accessConfig.Custom == nil {
			continue
		}
		var job bacnet.Job
		if err := yaml.Unmarshal([]byte(*accessConfig.Custom), &job); err != nil {
			return nil, err
		}

		// generate jobMap
		var jobMaps []bacnet.MapConfig
		devTpl, err := ctx.GetAccessTemplates(&devInfo)
		if err != nil {
			return nil, err
		}
		if devTpl != nil && devTpl.Properties != nil && len(devTpl.Properties) > 0 {
			for _, prop := range devTpl.Properties {
				if visitor := prop.Visitor.Custom; visitor != nil {
					var jobMap bacnet.MapConfig
					if err := yaml.Unmarshal([]byte(*visitor), &jobMap); err != nil {
						return nil, err
					}
					jobMap.Id = prop.Id
					jobMap.Name = prop.Name
					jobMap.Type = prop.Type
					jobMaps = append(jobMaps, jobMap)
				}
			}
		}
		job.Maps = jobMaps
		jobs = append(jobs, job)
	}
	cfg.Jobs = jobs
	cfg.Slaves = slaves
	if err := utils.SetDefaults(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
