package encoding

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"

	"github.com/baetyl/baetyl-bacnet/bacnet"
)

//Decoder is the struct used to turn byte arrays to baetyl-bacnet types. All
//public methods of decoder can set the internal error value. If such
//error is set, all decoding methods will be no-ops. This allows to
//defer error checking after several decoding operations
type Decoder struct {
	buf *bytes.Buffer
	err error
	//tagCounter int
}

func NewDecoder(b []byte) *Decoder {
	return &Decoder{
		buf: bytes.NewBuffer(b),
		err: nil,
	}
}

func (d *Decoder) Error() error {
	return d.err
}
func (d *Decoder) ResetError() {
	d.err = nil
}

//unread unread the last n bytes read from the decoder. This allows to retry decoding of the same data
func (d *Decoder) unread(n int) error {
	for x := 0; x < n; x++ {
		err := d.buf.UnreadByte()
		if err != nil {
			return err
		}
	}
	return nil
}

//ContextValue reads the next context tag/value couple and set val accordingly.
//Sets the decoder error  if the tagID isn't the expected or if the tag isn't contextual.
//If ErrorIncorrectTag is set, the internal buffer cursor is ready to read again the same tag.
func (d *Decoder) ContextValue(expectedTagID byte, val *uint32) {
	if d.err != nil {
		return
	}
	length, t, err := decodeTag(d.buf)
	if err != nil {
		d.err = err
		return
	}
	if t.ID != expectedTagID {
		d.err = ErrorIncorrectTagID{Expected: expectedTagID, Got: t.ID}
		err := d.unread(length)
		if err != nil {
			d.err = err
		}
		return
	}
	if !t.Context {
		d.err = errors.New("tag isn't contextual")
	}
	v, err := decodeUnsignedWithLen(d.buf, int(t.Value))
	if err != nil {
		d.err = err
		return
	}
	*val = v
}

//ContextObjectID read a (context)tag / value pair where the value
//type is an unsigned int
//If ErrorIncorrectTag is set, the internal buffer cursor is ready to read again the same tag.
func (d *Decoder) ContextObjectID(expectedTagID byte, objectID *bacnet.ObjectID) {
	if d.err != nil {
		return
	}
	length, t, err := decodeTag(d.buf)
	if err != nil {
		d.err = err
		return
	}

	if t.ID != expectedTagID {
		d.err = ErrorIncorrectTagID{Expected: expectedTagID, Got: t.ID}
		err := d.unread(length)
		if err != nil {
			d.err = err
		}
		return
	}
	if !t.Context {
		d.err = errors.New("tag isn't contextual")
		return
	}
	//Todo: check is tag size is ok
	var val uint32
	_ = binary.Read(d.buf, binary.BigEndian, &val)
	*objectID = bacnet.ObjectIDFromUint32(val)
}

type AppDataTypeMismatch struct {
	wanted string
	got    reflect.Type
}

func (e AppDataTypeMismatch) Error() string {
	return fmt.Sprintf("decode AppData: mismatched type, cannot decode %s in type %s", e.wanted, e.got.String())
}

//AppData read the next tag and value. The value type advertised
//in tag must be a standard baetyl-bacnet application data type and must
//match the type passed in the v parameter. If no error is
//returned, v will contain the data read
func (d *Decoder) AppData(v interface{}) {
	if d.err != nil {
		return
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		d.err = errors.New("decodeAppData: interface parameter isn't a pointer")
		return
	}
	_, tag, err := decodeTag(d.buf)
	if err != nil {
		d.err = fmt.Errorf("decodeAppData: read tag: %w", err)
		return
	}
	if tag.Context {
		d.err = errors.New("decode AppData: unexpected context tag ")
		return
	}
	//Take the pointer value
	rv = rv.Elem()
	switch tag.ID {
	case applicationTagNull:
		//nothing to do
	case applicationTagUnsignedInt:
		val, err := decodeUnsignedWithLen(d.buf, int(tag.Value))
		if err != nil {
			d.err = fmt.Errorf("decodeAppData: read ObjectID: %w", err)
			return
		}
		if rv.Kind() != reflect.Uint8 && rv.Kind() != reflect.Uint16 && rv.Kind() != reflect.Uint32 && !isEmptyInterface(rv) {
			d.err = AppDataTypeMismatch{wanted: "UnsignedInt", got: rv.Type()}
			return
		}
		rv.Set(reflect.ValueOf(val))
	case applicationTagReal:
		var f float32
		err := binary.Read(d.buf, binary.BigEndian, &f)
		if err != nil {
			d.err = fmt.Errorf("decode AppData: read float32: %w", err)
			return
		}
		if rv.Kind() != reflect.Float32 && !isEmptyInterface(rv) {
			d.err = AppDataTypeMismatch{wanted: "Real", got: rv.Type()}
			return
		}
		rv.Set(reflect.ValueOf(f))
	case applicationTagCharacterString:
		sEncoding, err := d.buf.ReadByte()
		if err != nil {
			d.err = fmt.Errorf("decode appData: read string encoding: %w", err)
			return
		}
		if sEncoding != utf8Encoding {
			d.err = fmt.Errorf("unsuported strign encoding: 0x%x", sEncoding)
			return
		}
		b := make([]byte, int(tag.Value)-1) //Minus one because encoding is already consumed
		n, err := d.buf.Read(b)
		if err != nil {
			d.err = fmt.Errorf("decode appdata: read string: %w", err)
			return
		}
		if n != len(b) {
			d.err = fmt.Errorf("decode string: stort read %d instead of %d", n, len(b))
			return
		}
		s := string(b) //Conversion allowed because string are utf8 only in go
		if rv.Type() != reflect.TypeOf(s) && !isEmptyInterface(rv) {
			d.err = AppDataTypeMismatch{wanted: "CharacterString", got: rv.Type()}
			return
		}
		rv.Set(reflect.ValueOf(s))
	case applicationTagEnumerated:
		val, err := decodeUnsignedWithLen(d.buf, int(tag.Value))
		if err != nil {
			d.err = fmt.Errorf("decodeAppData: read ObjectID: %w", err)
			return
		}
		switch rv.Type() {
		case reflect.TypeOf(bacnet.SegmentationSupport(0)):
			rv.Set(reflect.ValueOf(bacnet.SegmentationSupport(val)))
		case reflect.TypeOf(bacnet.ErrorClass(0)):
			rv.Set(reflect.ValueOf(bacnet.ErrorClass(val)))
		case reflect.TypeOf(bacnet.ErrorCode(0)):
			rv.Set(reflect.ValueOf(bacnet.ErrorCode(val)))
		default:
			if isEmptyInterface(rv) {
				rv.Set(reflect.ValueOf(val))
			} else {
				d.err = AppDataTypeMismatch{wanted: "Enumerated", got: rv.Type()}
				return
			}
		}
	case applicationTagObjectID:
		var obj bacnet.ObjectID
		var val uint32
		err := binary.Read(d.buf, binary.BigEndian, &val)
		if err != nil {
			d.err = fmt.Errorf("decodeAppData: read ObjectID: %w", err)
			return
		}
		obj = bacnet.ObjectIDFromUint32(val)
		if rv.Type() != reflect.TypeOf(obj) && !isEmptyInterface(rv) {
			d.err = AppDataTypeMismatch{wanted: "ObjectID", got: rv.Type()}
			return
		}
		rv.Set(reflect.ValueOf(obj))
	case applicationTagBoolean:
		rv.Set(reflect.ValueOf(tag.Value != 0))
	default:
		//TODO: support all app data types
		d.err = fmt.Errorf("decodeAppData: unsupported type 0x%x", tag.ID)
		return
	}
}

func isEmptyInterface(rv reflect.Value) bool {
	return rv.Kind() == reflect.Interface && rv.Type().NumMethod() == 0
}

const utf8Encoding = byte(0)

func (d *Decoder) ContextAbstractType(expectedTagNumber byte, v interface{}) {
	if d.err != nil {
		return
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		d.err = errors.New("decodeAppData: interface parameter isn't a pointer")
		return
	}
	_, tag, err := decodeTag(d.buf)
	if err != nil {
		d.err = fmt.Errorf("decoder abstractType: read opening tag: %w", err)
		return
	}
	if !tag.Opening {
		d.err = fmt.Errorf("decoder abstractType: expected opening tag")
		return
	}
	if tag.ID != expectedTagNumber {
		d.err = ErrorIncorrectTagID{Expected: expectedTagNumber, Got: tag.ID}
	}
	//Todo: check if we can have several tag inside the Opening/closing pair
	d.AppData(v)
	if d.err != nil {
		return
	}
	_, tag, err = decodeTag(d.buf)
	if err != nil {
		d.err = fmt.Errorf("decoder abstractType: read closing tag: %w", err)
		return
	}
	if !tag.Closing {
		d.err = fmt.Errorf("decoder abstractType: expected closing tag")
		return
	}
	if tag.ID != expectedTagNumber {
		d.err = ErrorIncorrectTagID{Expected: expectedTagNumber, Got: tag.ID}
	}

}

const (
	size8  = 1
	size16 = 2
	size24 = 3
	size32 = 4
)

func decodeUnsignedWithLen(buf *bytes.Buffer, length int) (uint32, error) {
	switch length {
	case size8:
		val, err := buf.ReadByte()
		if err != nil {
			return 0, fmt.Errorf("read unsigned with length 1 : %w", err)
		}
		return uint32(val), nil
	case size16:
		var val uint16
		err := binary.Read(buf, binary.BigEndian, &val)
		if err != nil {
			return 0, fmt.Errorf("read unsigned with length 2 : %w", err)
		}
		return uint32(val), nil
	case size24:
		// There is no default 24 bit integer in go, so we have tXo
		// write it manually (in big endian)
		var val uint16
		msb, err := buf.ReadByte()
		if err != nil {
			return 0, fmt.Errorf("read unsigned with length 3 : %w", err)
		}
		err = binary.Read(buf, binary.BigEndian, &val)
		if err != nil {
			return 0, fmt.Errorf("read unsigned with length 3 : %w", err)
		}
		return uint32(msb)<<16 + uint32(val), nil
	case size32:
		var val uint32
		err := binary.Read(buf, binary.BigEndian, &val)
		if err != nil {
			return 0, fmt.Errorf("read unsigned with length 4 : %w", err)
		}
		return val, nil
	default:
		//TODO: check If allowed by specification, other
		//implementation allow it but i'm not sure
		return 0, nil
	}
}
