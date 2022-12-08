package baetyl_bacnet

import "time"

type Config struct {
	Slaves []SlaveConfig `yaml:"slaves" json:"slaves"`
	Jobs   []Job         `yaml:"jobs" json:"jobs"`
}

type SlaveConfig struct {
	Device   string        `yaml:"device" json:"device"`
	DeviceId uint32        `yaml:"deviceId" json:"deviceId"`
	Interval time.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`
	Address  string        `yaml:"address" json:"address"`
	Port     int           `yaml:"port" json:"port"`
}

type Job struct {
	Device        string              `yaml:"device" json:"device"`
	DeviceId      uint32              `yaml:"deviceId" json:"deviceId"`
	AddressOffset uint                `yaml:"addressOffset" json:"addressOffset"`
	Properties    map[string]Property `yaml:"properties" json:"properties"`
	Interval      time.Duration       `yaml:"interval" json:"interval" default:"15s"`
}

type Property struct {
	Id            string `yaml:"id" json:"id"`
	Name          string `yaml:"name" json:"name"`
	Type          string `yaml:"type" json:"type"`
	Mode          string `yaml:"mode" json:"mode"`
	BacnetType    uint   `yaml:"bacnetType" json:"bacnetType"`
	BacnetAddress uint   `yaml:"bacnetAddress" json:"bacnetAddress"`
}
