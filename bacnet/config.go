package bacnet

type Config struct {
	Slaves []SlaveConfig `yaml:"slaves" json:"slaves"`
	Jobs   []Job         `yaml:"jobs" json:"jobs"`
}

type SlaveConfig struct {
	Address string `yaml:"address" json:"address"`
	Port    int    `yaml:"port" json:"port"`
}

type Job struct {
	DeviceId string      `yaml:"deviceId" json:"deviceId"`
	Maps     []MapConfig `yaml:"maps" json:"maps"`
}

type MapConfig struct {
	Id            string `yaml:"id" json:"id"`
	Name          string `yaml:"name" json:"name"`
	Type          string `yaml:"type" json:"type"`
	BacnetType    uint   `yaml:"bacnetType" json:"bacnetType"`
	BacnetAddress uint   `yaml:"bacnetAddress" json:"bacnetAddress"`
}
