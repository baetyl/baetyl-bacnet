package encoding

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

//ErrorIncorrectTagID is the error return when the decoded tag doesn't
//have the expected ID
type ErrorIncorrectTagID struct {
	Expected byte
	Got      byte
}

func (e ErrorIncorrectTagID) Error() string {
	return fmt.Sprintf("incorrect tagID %d, expected %d.", e.Got, e.Expected)
}

//nolint: deadcode, varcheck
const (
	applicationTagNull            byte = 0x00
	applicationTagBoolean         byte = 0x01
	applicationTagUnsignedInt     byte = 0x02
	applicationTagSignedInt       byte = 0x03
	applicationTagReal            byte = 0x04
	applicationTagDouble          byte = 0x05
	applicationTagOctetString     byte = 0x06
	applicationTagCharacterString byte = 0x07
	applicationTagBitString       byte = 0x08
	applicationTagEnumerated      byte = 0x09
	applicationTagDate            byte = 0x0A
	applicationTagTime            byte = 0x0B
	applicationTagObjectID        byte = 0x0C
)

type tag struct {
	// Tag id. Typically sequential when tag is contextual. Or refer
	// to the standard AppData Types
	ID      byte
	Context bool
	// Either has a value or length of the next value
	Value   uint32
	Opening bool
	Closing bool
}

func isExtendedTagNumber(x byte) bool {
	return x&0xF0 == 0xF0
}

func isExtendedValue(x byte) bool {
	return x&7 == 5
}

func isOpeningTag(x byte) bool {
	return x&7 == 6

}
func isClosingTag(x byte) bool {
	return x&7 == 7
}

func isContextSpecific(x byte) bool {
	return x&8 > 0
}

func encodeTag(buf *bytes.Buffer, t tag) {
	var tagMeta byte
	if t.Context {
		tagMeta |= 0x8
	}
	if t.Opening {
		tagMeta |= 0x6
	}
	if t.Closing {
		tagMeta |= 0x7
	}
	if t.Value <= 4 {
		tagMeta |= byte(t.Value)
	} else {
		tagMeta |= 5
	}

	if t.ID <= 14 {
		tagMeta |= (t.ID << 4)
		buf.WriteByte(tagMeta)

		// We don't have enough space so make it in a new byte
	} else {
		tagMeta |= byte(0xF0)
		buf.WriteByte(tagMeta)
		buf.WriteByte(t.ID)
	}

	if t.Value > 4 {
		// Depending on the length, we will either write it as an 8 bit, 32 bit, or 64 bit integer
		if t.Value <= 253 {
			buf.WriteByte(byte(t.Value))
		} else if t.Value <= 65535 {
			buf.WriteByte(flag16bits)
			_ = binary.Write(buf, binary.BigEndian, uint16(t.Value))
		} else {
			buf.WriteByte(flag32bits)
			_ = binary.Write(buf, binary.BigEndian, t.Value)
		}
	}
}

func decodeTag(buf *bytes.Buffer) (length int, t tag, err error) {
	firstByte, err := buf.ReadByte()
	if err != nil {
		return length, t, fmt.Errorf("read tagID: %w", err)
	}
	length++
	if isExtendedTagNumber(firstByte) {
		tagNumber, err := buf.ReadByte()
		if err != nil {
			return length, t, fmt.Errorf("read extended tagId: %w", err)
		}
		length++
		t.ID = tagNumber
	} else {
		tagNumber := firstByte >> 4
		t.ID = tagNumber
	}
	if isContextSpecific(firstByte) {
		t.Context = true
	}

	if isOpeningTag(firstByte) {
		t.Opening = true
		return length, t, nil
	}
	if isClosingTag(firstByte) {
		t.Closing = true
		return length, t, nil
	}

	if isExtendedValue(firstByte) {
		firstValueByte, err := buf.ReadByte()
		if err != nil {
			return length, t, fmt.Errorf("read first byte of extended value tag: %w", err)
		}
		length++
		switch firstValueByte {
		case flag16bits:
			var val uint16
			err := binary.Read(buf, binary.BigEndian, &val)
			if err != nil {
				return length, t, fmt.Errorf("read extended 16bits tag value: %w ", err)
			}
			length += 2
			t.Value = uint32(val)
		case flag32bits:
			err := binary.Read(buf, binary.BigEndian, &t.Value)
			if err != nil {
				return length, t, fmt.Errorf("read extended 32bits tag value: %w", err)
			}
			length += 4
		default:
			t.Value = uint32(firstValueByte)

		}
	} else {
		t.Value = uint32(firstByte & 0x7)
	}
	return length, t, nil
}
