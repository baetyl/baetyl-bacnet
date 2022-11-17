package bacip

import (
	"errors"
	"fmt"

	"github.com/baetyl/baetyl-bacnet/bacnet"
	"github.com/baetyl/baetyl-bacnet/internal/encoding"
)

type WhoIs struct {
	Low, High *uint32 //may be null if we want to check all range
}

func (w WhoIs) MarshalBinary() ([]byte, error) {
	encoder := encoding.NewEncoder()
	if w.Low != nil && w.High != nil {
		if *w.Low > bacnet.MaxInstance || *w.High > bacnet.MaxInstance {
			return nil, fmt.Errorf("invalid WhoIs range: [%d, %d]: max value is %d", *w.Low, *w.High, bacnet.MaxInstance)
		}
		if *w.Low > *w.High {
			return nil, fmt.Errorf("invalid WhoIs range: [%d, %d]: low limit is higher than high limit", *w.Low, *w.High)
		}
		encoder.ContextUnsigned(0, *w.Low)
		encoder.ContextUnsigned(1, *w.High)
	}
	return encoder.Bytes(), encoder.Error()
}

func (w *WhoIs) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		// If data is empty, the whoIs request is a full range
		// check. So keep the low and high pointer nil
		return nil
	}
	w.Low = new(uint32)
	w.High = new(uint32)
	decoder := encoding.NewDecoder(data)
	decoder.ContextValue(byte(0), w.Low)
	decoder.ContextValue(byte(1), w.High)
	return decoder.Error()
}

type Iam struct {
	ObjectID            bacnet.ObjectID
	MaxApduLength       uint32
	SegmentationSupport bacnet.SegmentationSupport
	VendorID            uint32
}

func (iam Iam) MarshalBinary() ([]byte, error) {
	encoder := encoding.NewEncoder()
	encoder.AppData(iam.ObjectID)
	encoder.AppData(iam.MaxApduLength)
	encoder.AppData(iam.SegmentationSupport)
	encoder.AppData(iam.VendorID)
	return encoder.Bytes(), encoder.Error()
}

func (iam *Iam) UnmarshalBinary(data []byte) error {
	decoder := encoding.NewDecoder(data)
	decoder.AppData(&iam.ObjectID)
	decoder.AppData(&iam.MaxApduLength)
	decoder.AppData(&iam.SegmentationSupport)
	decoder.AppData(&iam.VendorID)
	return decoder.Error()
}

type ReadProperty struct {
	ObjectID bacnet.ObjectID
	Property bacnet.PropertyIdentifier
	//Data is here to contains the response
	Data interface{}
}

func (rp ReadProperty) MarshalBinary() ([]byte, error) {

	encoder := encoding.NewEncoder()
	encoder.ContextObjectID(0, rp.ObjectID)
	encoder.ContextUnsigned(1, uint32(rp.Property.Type))
	if rp.Property.ArrayIndex != nil {
		encoder.ContextUnsigned(2, *rp.Property.ArrayIndex)
	}
	return encoder.Bytes(), encoder.Error()
}

func (rp *ReadProperty) UnmarshalBinary(data []byte) error {
	decoder := encoding.NewDecoder(data)
	decoder.ContextObjectID(0, &rp.ObjectID)
	var val uint32
	decoder.ContextValue(1, &val)
	rp.Property.Type = bacnet.PropertyType(val)
	rp.Property.ArrayIndex = new(uint32)
	decoder.ContextValue(2, rp.Property.ArrayIndex)
	err := decoder.Error()
	var e encoding.ErrorIncorrectTagID
	//This tag is optional, maybe it doesn't exist
	if err != nil && errors.As(err, &e) {
		rp.Property.ArrayIndex = nil
		decoder.ResetError()
	}
	decoder.ContextAbstractType(3, &rp.Data)
	return decoder.Error()
}

type WriteProperty struct {
	ObjectID      bacnet.ObjectID
	Property      bacnet.PropertyIdentifier
	PropertyValue bacnet.PropertyValue
	Priority      bacnet.PriorityList
}

func (wp WriteProperty) MarshalBinary() ([]byte, error) {
	encoder := encoding.NewEncoder()
	encoder.ContextObjectID(0, wp.ObjectID)
	encoder.ContextUnsigned(1, uint32(wp.Property.Type))
	if wp.Property.ArrayIndex != nil {
		encoder.ContextUnsigned(2, uint32(*wp.Property.ArrayIndex))
	}
	err := encoder.ContextAsbtractType(3, wp.PropertyValue)
	if err != nil {
		return nil, err
	}
	if wp.Priority != 0 {
		encoder.ContextUnsigned(4, uint32(wp.Priority))
	}
	return encoder.Bytes(), encoder.Error()
}

func (wp *WriteProperty) UnmarshalBinary(data []byte) error {
	decoder := encoding.NewDecoder(data)
	return decoder.Error()
}

type ApduError struct {
	Class bacnet.ErrorClass
	Code  bacnet.ErrorCode
}

func (e ApduError) Error() string {
	return fmt.Sprintf("apdu error class %v code %v", e.Class, e.Code)
}
func (e ApduError) MarshalBinary() ([]byte, error) {
	panic("not implemented")
}

func (e *ApduError) UnmarshalBinary(data []byte) error {
	decoder := encoding.NewDecoder(data)
	decoder.AppData(&e.Class)
	decoder.AppData(&e.Code)
	return decoder.Error()
}

// Todo http://kargs.net/BACnet/BACnet_Essential_Objects_Services.pdf -> Time synchro, Reinitialize device, DeviceCommunicationControl
