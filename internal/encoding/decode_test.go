package encoding

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/baetyl/baetyl-bacnet/bacnet"
	"github.com/matryer/is"
)

func TestValidAppData(t *testing.T) {
	ttc := []struct {
		data        string //hex string
		expected    interface{}
		expectedBis interface{} //nil if same as one
	}{
		{
			data: "c4020075e9",
			expected: bacnet.ObjectID{
				Type:     8,
				Instance: 30185,
			},
		},
		{
			data:     "2205c4",
			expected: uint32(1476),
		},
		{
			data:        "9100",
			expected:    bacnet.SegmentationSupportBoth,
			expectedBis: uint32(0),
		},
		{
			data:     "22016c",
			expected: uint32(364),
		},
		{
			data:     "7511004543592d53313030302d413437383035",
			expected: "ECY-S1000-A47805",
		},
		{
			data:     "4400000000",
			expected: float32(0),
		},
		{
			data:     "00",
			expected: nil,
		},
	}
	for _, tc := range ttc {
		t.Run(fmt.Sprintf("AppData decode %s (%T)", tc.data, tc.expected), func(t *testing.T) {
			is := is.New(t)
			b, err := hex.DecodeString(tc.data)
			is.NoErr(err)
			decoder := NewDecoder(b)
			//Ensure that it work when passed the concrete type
			switch tc.expected.(type) {
			case bacnet.ObjectID:
				x := bacnet.ObjectID{}
				decoder.AppData(&x)
				is.NoErr(decoder.err)
				is.Equal(x, tc.expected)
			case uint32:
				var x uint32
				decoder.AppData(&x)
				is.NoErr(decoder.err)
				is.Equal(x, tc.expected)
			case bacnet.SegmentationSupport:
				var x bacnet.SegmentationSupport
				decoder.AppData(&x)
				is.NoErr(decoder.err)
				is.Equal(x, tc.expected)
			case string:
				var x string
				decoder.AppData(&x)
				is.NoErr(decoder.err)
				is.Equal(x, tc.expected)
			case float32:
				var x float32
				decoder.AppData(&x)
				is.NoErr(decoder.err)
				is.Equal(x, tc.expected)
			default:
				if tc.expected != nil { //This is for NullTag to pass
					t.Errorf("Invalid from type %T", tc.expected)
				}

			}
			//Ensure that it work when passed an empty interface
			var v interface{}
			decoder = NewDecoder(b)
			decoder.AppData(&v)
			is.NoErr(decoder.err)
			if tc.expectedBis != nil {
				is.Equal(v, tc.expectedBis)
			} else {
				is.Equal(v, tc.expected)
			}

		})
		t.Run(fmt.Sprintf("AppData encode %s (%T)", tc.data, tc.expected), func(t *testing.T) {
			is := is.New(t)
			enc := NewEncoder()
			enc.AppData(tc.expected)
			is.NoErr(enc.Error())
			is.Equal(hex.EncodeToString(enc.Bytes()), tc.data)
		})
	}
}
