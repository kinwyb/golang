package phpserialize

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
)

const UNSERIAZABLE_OBJECT_MAX_LEN = 10 * 1024 * 1024 * 1024

//UnSerialize 使用默认函数解析序列化
func UnSerialize(s string) (PhpValue, error) {
	decoder := NewUnSerializer(s)
	decoder.SetSerializedDecodeFunc(SerializedDecodeFunc(UnSerialize))
	return decoder.Decode()
}

//UnSerializer 解析序列化类型
type UnSerializer struct {
	source     string
	r          *strings.Reader
	lastErr    error
	decodeFunc SerializedDecodeFunc
}

//NewUnSerializer 新的解析序列化类型
func NewUnSerializer(data string) *UnSerializer {
	return &UnSerializer{
		source: data,
	}
}

//SetReader 设置解析结果流
func (s *UnSerializer) SetReader(r *strings.Reader) {
	s.r = r
}

//SetSerializedDecodeFunc 设置解析结果方法
func (s *UnSerializer) SetSerializedDecodeFunc(f SerializedDecodeFunc) {
	s.decodeFunc = f
}

//Decode 解析数据
func (s *UnSerializer) Decode() (PhpValue, error) {
	if s.r == nil {
		s.r = strings.NewReader(s.source)
	}
	var value PhpValue
	if token, _, err := s.r.ReadRune(); err == nil {
		switch token {
		default:
			s.saveError(fmt.Errorf("php_serialize: Unknown token %#U", token))
		case TOKEN_NULL:
			value = s.decodeNull()
		case TOKEN_BOOL:
			value = s.decodeBool()
		case TOKEN_INT:
			value = s.decodeNumber(false)
		case TOKEN_FLOAT:
			value = s.decodeNumber(true)
		case TOKEN_STRING:
			value = s.decodeString(DELIMITER_STRING_LEFT, DELIMITER_STRING_RIGHT, true)
		case TOKEN_ARRAY:
			value = s.decodeArray()
		case TOKEN_OBJECT:
			value = s.decodeObject()
		case TOKEN_OBJECT_SERIALIZED:
			value = s.decodeSerialized()
		case TOKEN_REFERENCE, TOKEN_REFERENCE_OBJECT:
			value = s.decodeReference()
		case TOKEN_SPL_ARRAY:
			value = s.decodeSplArray()

		}
	}

	return value, s.lastErr
}

func (s *UnSerializer) decodeNull() PhpValue {
	s.expect(SEPARATOR_VALUES)
	return nil
}

func (s *UnSerializer) decodeBool() PhpValue {
	var (
		raw rune
		err error
	)
	s.expect(SEPARATOR_VALUE_TYPE)

	if raw, _, err = s.r.ReadRune(); err != nil {
		s.saveError(fmt.Errorf("php_serialize: Error while reading bool value: %v", err))
	}

	s.expect(SEPARATOR_VALUES)
	return raw == '1'
}

func (s *UnSerializer) decodeNumber(isFloat bool) PhpValue {
	var (
		raw string
		err error
		val PhpValue
	)
	s.expect(SEPARATOR_VALUE_TYPE)

	if raw, err = s.readUntil(SEPARATOR_VALUES); err != nil {
		s.saveError(fmt.Errorf("php_serialize: Error while reading number value: %v", err))
	} else {
		if isFloat {
			if val, err = strconv.ParseFloat(raw, 64); err != nil {
				s.saveError(fmt.Errorf("php_serialize: Unable to convert %s to float: %v", raw, err))
			}
		} else {
			if val, err = strconv.Atoi(raw); err != nil {
				s.saveError(fmt.Errorf("php_serialize: Unable to convert %s to int: %v", raw, err))
			}
		}
	}

	return val
}

func (s *UnSerializer) decodeString(left, right rune, isFinal bool) PhpValue {
	var (
		err     error
		val     PhpValue
		strLen  int
		readLen int
	)

	strLen = s.readLen()
	s.expect(left)

	if strLen > 0 {
		buf := make([]byte, strLen, strLen)
		if readLen, err = s.r.Read(buf); err != nil {
			s.saveError(fmt.Errorf("php_serialize: Error while reading string value: %v", err))
		} else {
			if readLen != strLen {
				s.saveError(fmt.Errorf("php_serialize: Unable to read string. Expected %d but have got %d bytes", strLen, readLen))
			} else {
				val = string(buf)
			}
		}
	}

	s.expect(right)
	if isFinal {
		s.expect(SEPARATOR_VALUES)
	}
	return val
}

func (s *UnSerializer) decodeArray() PhpValue {
	var arrLen int
	val := make(PhpArray)

	arrLen = s.readLen()
	s.expect(DELIMITER_OBJECT_LEFT)

	for i := 0; i < arrLen; i++ {
		k, errKey := s.Decode()
		v, errVal := s.Decode()

		if errKey == nil && errVal == nil {
			val[k] = v
			/*switch t := k.(type) {
			default:
				s.saveError(fmt.Errorf("php_serialize: Unexpected key type %T", t))
			case string:
				stringKey, _ := k.(string)
				val[stringKey] = v
			case int:
				intKey, _ := k.(int)
				val[strconv.Itoa(intKey)] = v
			}*/
		} else {
			s.saveError(fmt.Errorf("php_serialize: Error while reading key or(and) value of array"))
		}
	}

	s.expect(DELIMITER_OBJECT_RIGHT)
	return val
}

func (s *UnSerializer) decodeObject() PhpValue {
	val := &PhpObject{
		className: s.readClassName(),
	}

	rawMembers := s.decodeArray()
	val.members, _ = rawMembers.(PhpArray)

	return val
}

func (s *UnSerializer) decodeSerialized() PhpValue {
	val := &PhpObjectSerialized{
		className: s.readClassName(),
	}

	rawData := s.decodeString(DELIMITER_OBJECT_LEFT, DELIMITER_OBJECT_RIGHT, false)
	val.data, _ = rawData.(string)

	if s.decodeFunc != nil && val.data != "" {
		var err error
		if val.value, err = s.decodeFunc(val.data); err != nil {
			s.saveError(err)
		}
	}

	return val
}

func (s *UnSerializer) decodeReference() PhpValue {
	s.expect(SEPARATOR_VALUE_TYPE)
	if _, err := s.readUntil(SEPARATOR_VALUES); err != nil {
		s.saveError(fmt.Errorf("php_serialize: Error while reading reference value: %v", err))
	}
	return nil
}

func (s *UnSerializer) expect(expected rune) {
	if token, _, err := s.r.ReadRune(); err != nil {
		s.saveError(fmt.Errorf("php_serialize: Error while reading expected rune %#U: %v", expected, err))
	} else if token != expected {
		if debugMode {
			log.Printf("php_serialize: source\n%s\n", s.source)
			log.Printf("php_serialize: reader info\n%#v\n", s.r)
		}
		s.saveError(fmt.Errorf("php_serialize: Expected %#U but have got %#U", expected, token))
	}
}

func (s *UnSerializer) readUntil(stop rune) (string, error) {
	var (
		token rune
		err   error
	)
	buf := bytes.NewBuffer([]byte{})

	for {
		if token, _, err = s.r.ReadRune(); err != nil || token == stop {
			break
		} else {
			buf.WriteRune(token)
		}
	}

	return buf.String(), err
}

func (s *UnSerializer) readLen() int {
	var (
		raw string
		err error
		val int
	)
	s.expect(SEPARATOR_VALUE_TYPE)

	if raw, err = s.readUntil(SEPARATOR_VALUE_TYPE); err != nil {
		s.saveError(fmt.Errorf("php_serialize: Error while reading lenght of value: %v", err))
	} else {
		if val, err = strconv.Atoi(raw); err != nil {
			s.saveError(fmt.Errorf("php_serialize: Unable to convert %s to int: %v", raw, err))
		} else if val > UNSERIAZABLE_OBJECT_MAX_LEN {
			s.saveError(fmt.Errorf("php_serialize: Unserializable object length looks too big(%d). If you are sure you wanna unserialise it, please increase UNSERIAZABLE_OBJECT_MAX_LEN const", val, err))
			val = 0
		}
	}
	return val
}

func (s *UnSerializer) readClassName() (res string) {
	rawClass := s.decodeString(DELIMITER_STRING_LEFT, DELIMITER_STRING_RIGHT, false)
	res, _ = rawClass.(string)
	return
}

func (s *UnSerializer) saveError(err error) {
	if s.lastErr == nil {
		s.lastErr = err
	}
}

func (s *UnSerializer) decodeSplArray() PhpValue {
	var err error
	val := &PhpSplArray{}

	s.expect(SEPARATOR_VALUE_TYPE)
	s.expect(TOKEN_INT)

	flags := s.decodeNumber(false)
	if flags == nil {
		s.saveError(fmt.Errorf("php_serialize: Unable to read flags of SplArray"))
		return nil
	}
	val.flags = PhpValueInt(flags)

	if val.array, err = s.Decode(); err != nil {
		s.saveError(fmt.Errorf("php_serialize: Can't parse SplArray: %v", err))
		return nil
	}

	s.expect(SEPARATOR_VALUES)
	s.expect(TOKEN_SPL_ARRAY_MEMBERS)
	s.expect(SEPARATOR_VALUE_TYPE)

	if val.properties, err = s.Decode(); err != nil {
		s.saveError(fmt.Errorf("php_serialize: Can't parse properties of SplArray: %v", err))
		return nil
	}

	return val
}
