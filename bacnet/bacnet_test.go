package bacnet

import (
	"testing"

	dm "github.com/baetyl/baetyl-go/v2/dmcontext"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/stretchr/testify/assert"
)

func TestBacnet(t *testing.T) {
	bac := Bacnet{}

	err := bac.Close()
	assert.NoError(t, err)

	devInfo := &dm.DeviceInfo{}
	err = bac.DeltaCallback(devInfo, v1.Delta{})
	assert.NoError(t, err)
	err = bac.PropertyGetCallback(devInfo, []string{})
	assert.NoError(t, err)
}
