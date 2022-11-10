package bacnet

import (
	"testing"

	dm "github.com/baetyl/baetyl-go/v2/dmcontext"
	"github.com/stretchr/testify/assert"
)

func TestWorker(t *testing.T) {
	ctx := &dm.DmCtx{}
	worker := NewWorker(ctx, nil)
	assert.NotNil(t, worker)
}
