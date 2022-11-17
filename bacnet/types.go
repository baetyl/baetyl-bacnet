//Package baetyl-bacnet provides various types to represent Bacnet related concepts
package bacnet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
)

const (
	MaxInstance   = 0x3FFFFF
	instanceBits  = 22
	maxObjectType = 0x400
)

//ObjectType is the category of an object
type ObjectType uint16

//ObjectInstance is a unique identifier of an baetyl-bacnet object
type ObjectInstance uint32

//go:generate stringer -type=ObjectType
const (
	AnalogInput           ObjectType = 0x00
	AnalogOutput          ObjectType = 0x01
	AnalogValue           ObjectType = 0x02
	BinaryInput           ObjectType = 0x03
	BinaryOutput          ObjectType = 0x04
	BinaryValue           ObjectType = 0x05
	Calendar              ObjectType = 0x06
	Command               ObjectType = 0x07
	BacnetDevice          ObjectType = 0x08
	EventEnrollment       ObjectType = 0x09
	File                  ObjectType = 0x0A
	Group                 ObjectType = 0x0B
	Loop                  ObjectType = 0x0C
	MultiStateInput       ObjectType = 0x0D
	MultiStateOutput      ObjectType = 0x0E
	NotificationClass     ObjectType = 0x0F
	Program               ObjectType = 0x10
	Schedule              ObjectType = 0x11
	Averaging             ObjectType = 0x12
	MultiStateValue       ObjectType = 0x13
	Trendlog              ObjectType = 0x14
	LifeSafetyPoint       ObjectType = 0x15
	LifeSafetyZone        ObjectType = 0x16
	Accumulator           ObjectType = 0x17
	PulseConverter        ObjectType = 0x18
	EventLog              ObjectType = 0x19
	GlobalGroup           ObjectType = 0x1A
	TrendLogMultiple      ObjectType = 0x1B
	LoadControl           ObjectType = 0x1C
	StructuredView        ObjectType = 0x1D
	AccessDoor            ObjectType = 0x1E
	Timer                 ObjectType = 0x1F
	AccessCredential      ObjectType = 0x20 // Addendum 2008-j
	AccessPoint           ObjectType = 0x21
	AccessRights          ObjectType = 0x22
	AccessUser            ObjectType = 0x23
	AccessZone            ObjectType = 0x24
	CredentialDataInput   ObjectType = 0x25 // Authentication-factor-input
	NetworkSecurity       ObjectType = 0x26 // Addendum 2008-g
	BitstringValue        ObjectType = 0x27 // Addendum 2008-w
	CharacterstringValue  ObjectType = 0x28 // Addendum 2008-w
	DatePatternValue      ObjectType = 0x29 // Addendum 2008-w
	DateValue             ObjectType = 0x2a // Addendum 2008-w
	DatetimePatternValue  ObjectType = 0x2b // Addendum 2008-w
	DatetimeValue         ObjectType = 0x2c // Addendum 2008-w
	IntegerValue          ObjectType = 0x2d // Addendum 2008-w
	LargeAnalogValue      ObjectType = 0x2e // Addendum 2008-w
	OctetstringValue      ObjectType = 0x2f // Addendum 2008-w
	PositiveIntegerValue  ObjectType = 0x30 // Addendum 2008-w
	TimePatternValue      ObjectType = 0x31 // Addendum 2008-w
	TimeValue             ObjectType = 0x32 // Addendum 2008-w
	NotificationForwarder ObjectType = 0x33 // Addendum 2010-af
	AlertEnrollment       ObjectType = 0x34 // Addendum 2010-af
	Channel               ObjectType = 0x35 // Addendum 2010-aa
	LightingOutput        ObjectType = 0x36 // Addendum 2010-i
	BinaryLightingOutput  ObjectType = 0x37 // Addendum 135-2012az
	NetworkPort           ObjectType = 0x38 // Addendum 135-2012az
	ProprietaryMin        ObjectType = 0x80
	Proprietarymax        ObjectType = 0x3ff
)

//ObjectID represent the type of a baetyl-bacnet object and it's instance number
type ObjectID struct {
	Type     ObjectType
	Instance ObjectInstance
}

//Encode turns the object ID into a uint32 for encoding.  Returns an
//error if the ObjectID is invalid
func (o ObjectID) Encode() (uint32, error) {
	if o.Instance > MaxInstance {
		return 0, errors.New("invalid ObjectID: instance too high")
	}
	if o.Type > maxObjectType {
		return 0, errors.New("invalid ObjectID: objectType too high too high")
	}
	v := uint32(o.Type)<<instanceBits | (uint32(o.Instance))
	return v, nil
}

func ObjectIDFromUint32(v uint32) ObjectID {
	return ObjectID{
		Type:     ObjectType(v >> instanceBits),
		Instance: ObjectInstance(v & MaxInstance),
	}
}

//Device represent a baetyl-bacnet device. Note: A baetyl-bacnet device is different
//from a baetyl-bacnet object. A device "contains" several object. Only the device has a baetyl-bacnet address
type Device struct {
	ID           ObjectID
	MaxApdu      uint32
	Segmentation SegmentationSupport
	Vendor       uint32
	Addr         Address
}

//Address is the baetyl-bacnet address of an device.
type Address struct {
	// mac_len = 0 is a broadcast address
	// note: MAC for IP addresses uses 4 bytes for addr, 2 bytes for port
	Mac []byte
	// the following are used if the device is behind a router
	// net = 0 indicates local
	Net uint16 // BACnet network number
	Adr []byte // hwaddr (MAC) address
}

func AddressFromUDP(udp net.UDPAddr) *Address {
	b := bytes.NewBuffer(udp.IP)
	port := uint16(udp.Port)
	_ = binary.Write(b, binary.BigEndian, port)
	return &Address{
		Mac: b.Bytes(),
	}
}

func UDPFromAddress(addr Address) net.UDPAddr {
	return net.UDPAddr{
		IP:   addr.Mac[:4],
		Port: int(binary.BigEndian.Uint16(addr.Mac[4:])),
		Zone: "",
	}
}

//go:generate stringer -type=SegmentationSupport
type SegmentationSupport byte

const (
	SegmentationSupportBoth     SegmentationSupport = 0x00
	SegmentationSupportTransmit SegmentationSupport = 0x01
	SegmentationSupportReceive  SegmentationSupport = 0x02
	SegmentationSupportNone     SegmentationSupport = 0x03
)

//PropertyIdentifier is used to control a ReadProperty request
type PropertyIdentifier struct {
	Type PropertyType
	//Not null if it's an array property and we want only one index of
	//this array
	ArrayIndex *uint32
}

//go:generate stringer -type=PriorityList
type PriorityList uint8

const (
	ManualLifeSafety1          PriorityList = 1
	ManualLifeSafety2          PriorityList = 2
	Available3                 PriorityList = 3
	Available4                 PriorityList = 4
	CriticalEquipementControl5 PriorityList = 5
	MinimumOnOff6              PriorityList = 6
	Available7                 PriorityList = 7
	ManualOperator8            PriorityList = 8
	Available9                 PriorityList = 9
	Available10                PriorityList = 10
	Available11                PriorityList = 11
	Available12                PriorityList = 12
	Available13                PriorityList = 13
	Available14                PriorityList = 14
	Available15                PriorityList = 15
	Available16                PriorityList = 16
)

type PropertyValue struct {
	Type  PropertyValueType
	Value interface{}
}

type PropertyValueType byte

const (
	TypeNull            PropertyValueType = 0
	TypeBoolean         PropertyValueType = 1
	TypeUnsignedInt     PropertyValueType = 2
	TypeSignedInt       PropertyValueType = 3
	TypeReal            PropertyValueType = 4
	TypeDouble          PropertyValueType = 5
	TypeOctetString     PropertyValueType = 6
	TypeCharacterString PropertyValueType = 7
	TypeBitString       PropertyValueType = 8
	TypeEnumerated      PropertyValueType = 9
	TypeDate            PropertyValueType = 10
	TypeTime            PropertyValueType = 11
	TypeObjectID        PropertyValueType = 12
)
