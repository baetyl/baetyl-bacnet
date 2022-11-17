package bacip

import (
	"encoding/hex"
	"testing"

	"github.com/baetyl/baetyl-bacnet/bacnet"
	"github.com/matryer/is"
)

func TestWhoIsDec(t *testing.T) {
	is := is.New(t)
	data, err := hex.DecodeString("09001affff") //With range
	is.NoErr(err)
	w := &WhoIs{}
	err = w.UnmarshalBinary(data)
	is.NoErr(err)
	if w.Low == nil || *w.Low != 0 {
		t.Error("Invalid whois decoding of low range ")
	}
	if w.High == nil || *w.High != 0xFFFF {
		t.Error("Invalid whois decoding of high range ")
	}

	data, err = hex.DecodeString("09121b012345") //With range
	is.NoErr(err)
	w = &WhoIs{}
	err = w.UnmarshalBinary(data)
	is.NoErr(err)
	if w.Low == nil || *w.Low != 0x12 {
		t.Error("Invalid whois decoding of low range ")
	}
	if w.High == nil || *w.High != 0x12345 {
		t.Error("Invalid whois decoding of high range ")
	}

	data, err = hex.DecodeString("") //No range
	is.NoErr(err)
	w = &WhoIs{}
	err = w.UnmarshalBinary(data)
	is.NoErr(err)
	if w.High != nil || w.Low != nil {
		t.Error("Non nil range value")
	}
}

func TestWhoIsCoherency(t *testing.T) {
	ttc := []struct {
		data string //hex string
		name string
	}{
		{
			data: "09001affff",
			name: "Range 1-2",
		},
		{
			data: "",
			name: "Empty",
		},
		{
			data: "09121b012345",
			name: "Range 1-3",
		},
	}
	for _, tc := range ttc {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)
			b, err := hex.DecodeString(tc.data)
			is.NoErr(err)
			w := &WhoIs{}
			is.NoErr(w.UnmarshalBinary(b))
			b2, err := w.MarshalBinary()
			is.NoErr(err)
			is.Equal(hex.EncodeToString(b2), tc.data)
		})
	}
}
func TestIamEncodingAndCoherency(t *testing.T) {
	ttc := []struct {
		data string //hex string
		iam  Iam
	}{
		{
			data: "c4020075e92205c4910022016c",
			iam: Iam{
				ObjectID: bacnet.ObjectID{
					Type:     8,
					Instance: 30185,
				},
				MaxApduLength:       1476,
				SegmentationSupport: bacnet.SegmentationSupportBoth,
				VendorID:            364,
			},
		},
	}
	for _, tc := range ttc {
		t.Run(tc.data, func(t *testing.T) {
			is := is.New(t)
			result, err := tc.iam.MarshalBinary()
			is.NoErr(err)
			is.Equal(tc.data, hex.EncodeToString(result))
			iam := Iam{}
			is.NoErr(iam.UnmarshalBinary(result))
			result2, err := iam.MarshalBinary()
			is.NoErr(err)
			is.Equal(tc.data, hex.EncodeToString(result2))
		})
	}
}

func TestReadPropertyReq(t *testing.T) {
	ttc := []struct {
		data string //hex string
		rp   ReadProperty
	}{
		{
			data: "0c00401fb91975",
			rp: ReadProperty{
				ObjectID: bacnet.ObjectID{
					Type:     bacnet.AnalogOutput,
					Instance: 8121,
				},
				Property: bacnet.PropertyIdentifier{
					Type: bacnet.Units,
				},
			},
		},
	}
	for _, tc := range ttc {
		t.Run(tc.data, func(t *testing.T) {
			is := is.New(t)
			result, err := tc.rp.MarshalBinary()
			is.NoErr(err)
			is.Equal(hex.EncodeToString(result), tc.data)
		})
	}
}

func TestReadPropertyResp(t *testing.T) {
	ttc := []struct {
		data string //hex string
		rp   ReadProperty
	}{
		{
			data: "0c00401fb919753e91623f",
			rp: ReadProperty{
				ObjectID: bacnet.ObjectID{
					Type:     bacnet.AnalogOutput,
					Instance: 8121,
				},
				Property: bacnet.PropertyIdentifier{
					Type: bacnet.Units,
				},
				Data: uint32(98),
			},
		},
	}
	for _, tc := range ttc {
		t.Run(tc.data, func(t *testing.T) {
			is := is.New(t)
			rp := ReadProperty{}
			b, _ := hex.DecodeString(tc.data)
			is.NoErr(rp.UnmarshalBinary(b))
			is.Equal(rp, tc.rp)
		})
	}
}

func TestWritePropertyReq(t *testing.T) {
	ttc := []struct {
		data string
		wp   WriteProperty
	}{
		{
			data: "0c0100000119553e91003f",
			wp: WriteProperty{
				ObjectID: bacnet.ObjectID{
					Type:     bacnet.BinaryOutput,
					Instance: 1,
				},
				Property: bacnet.PropertyIdentifier{
					Type: bacnet.PresentValue,
				},
				PropertyValue: bacnet.PropertyValue{
					Type:  0x09,
					Value: 0,
				},
				Priority: 0,
			},
		},
	}
	for _, tc := range ttc {
		t.Run(tc.data, func(t *testing.T) {
			//is := is.New(t)
			//result, _ := tc.wp.MarshalBinary()
			//is.NoErr(err)
			//is.Equal(hex.EncodeToString(result), tc.data)
		})
	}
}
