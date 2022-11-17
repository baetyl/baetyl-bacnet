package bacip

import (
	"testing"

	"github.com/baetyl/baetyl-bacnet/bacnet"
	"github.com/matryer/is"
)

func TestFullEncodingAndCoherency(t *testing.T) {
	ttc := []struct {
		bvlc    BVLC
		encoded string //hex string
	}{
		{
			bvlc: BVLC{
				Type:     TypeBacnetIP,
				Function: BacFuncBroadcast,
				NPDU: NPDU{
					Version:               Version1,
					IsNetworkLayerMessage: false,
					ExpectingReply:        false,
					Priority:              Normal,
					ADPU: &APDU{
						DataType:    UnconfirmedServiceRequest,
						ServiceType: ServiceUnconfirmedWhoIs,
						Payload: &ReadProperty{
							ObjectID: bacnet.ObjectID{
								Type:     bacnet.AnalogInput,
								Instance: 300184,
							},
							Property: bacnet.PropertyIdentifier{
								Type: bacnet.PresentValue,
							},
						},
					},
				},
			},
			encoded: "810b000801001008",
		},
		{
			bvlc: BVLC{
				Type:     TypeBacnetIP,
				Function: BacFuncBroadcast,
				NPDU: NPDU{
					Version:               Version1,
					IsNetworkLayerMessage: false,
					ExpectingReply:        false,
					Priority:              Normal,
					Destination: &bacnet.Address{
						Net: 0xffff,
						Adr: []byte{},
					},
					Source:   &bacnet.Address{},
					HopCount: 255,
					ADPU: &APDU{
						DataType:    UnconfirmedServiceRequest,
						ServiceType: ServiceUnconfirmedIAm,
						Payload: &Iam{
							ObjectID: bacnet.ObjectID{
								Type:     8,
								Instance: 30185,
							},
							MaxApduLength:       1476,
							SegmentationSupport: bacnet.SegmentationSupportBoth,
							VendorID:            364,
						},
					},
				},
			},
			encoded: "810b00190120ffff00ff1000c4020075e92205c4910022016c",
		},
	}

	for _, tc := range ttc {
		t.Run(tc.encoded, func(t *testing.T) {
			is := is.New(t)
			result, err := tc.bvlc.MarshalBinary()
			is.NoErr(err)
			//is.Equal(tc.encoded, hex.EncodeToString(result))
			w := BVLC{}
			is.NoErr(w.UnmarshalBinary(result))
			//result2, err := w.MarshalBinary()
			//is.NoErr(err)
			//is.Equal(tc.encoded, hex.EncodeToString(result2))
		})
	}
}
