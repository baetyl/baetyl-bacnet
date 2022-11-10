package bacnet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlave(t *testing.T) {
	cfg := SlaveConfig{}
	slave, err := NewSlave(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, slave)
}
