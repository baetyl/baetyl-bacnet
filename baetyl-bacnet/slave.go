package baetyl_bacnet

import (
	"time"

	dm "github.com/baetyl/baetyl-go/v2/dmcontext"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"

	"github.com/baetyl/baetyl-bacnet/bacip"
	"github.com/baetyl/baetyl-bacnet/bacnet"
)

var bacnetClient *bacip.Client
var devices []bacnet.Device

type Slave struct {
	info         *dm.DeviceInfo
	bacnetClient *bacip.Client
	device       bacnet.Device
	cfg          SlaveConfig
}

func NewSlave(info *dm.DeviceInfo, cfg SlaveConfig) (*Slave, error) {
	if bacnetClient == nil {
		c, err := bacip.NewClientByIp(cfg.Address, cfg.Port)
		if err != nil {
			return nil, errors.Trace(err)
		}
		bacnetClient = c
	}
	if devices != nil && len(devices) > 0 {
		slave, err := generateSlave(info, cfg)
		if err == nil {
			return slave, nil
		}
	}
	devs, err := bacnetClient.WhoIs(bacip.WhoIs{}, 10*time.Second)
	if err != nil {
		return nil, errors.Trace(err)
	}
	log.L().Info("Find devices:", log.Any("devices", devs))
	devices = devs
	return generateSlave(info, cfg)
}

func generateSlave(info *dm.DeviceInfo, cfg SlaveConfig) (*Slave, error) {
	slave := &Slave{
		info:         info,
		cfg:          cfg,
		bacnetClient: bacnetClient,
	}
	for _, device := range devices {
		if uint32(device.ID.Instance) == cfg.DeviceId {
			slave.device = device
			return slave, nil
		}
	}
	return nil, errors.New("device not find")
}
