package phpserialize

const (
	TOKEN_NULL              rune = 'N'
	TOKEN_BOOL              rune = 'b'
	TOKEN_INT               rune = 'i'
	TOKEN_FLOAT             rune = 'd'
	TOKEN_STRING            rune = 's'
	TOKEN_ARRAY             rune = 'a'
	TOKEN_OBJECT            rune = 'O'
	TOKEN_OBJECT_SERIALIZED rune = 'C'
	TOKEN_REFERENCE         rune = 'R'
	TOKEN_REFERENCE_OBJECT  rune = 'r'
	TOKEN_SPL_ARRAY         rune = 'x'
	TOKEN_SPL_ARRAY_MEMBERS rune = 'm'

	SEPARATOR_VALUE_TYPE rune = ':'
	SEPARATOR_VALUES     rune = ';'

	DELIMITER_STRING_LEFT  rune = '"'
	DELIMITER_STRING_RIGHT rune = '"'
	DELIMITER_OBJECT_LEFT  rune = '{'
	DELIMITER_OBJECT_RIGHT rune = '}'

	FORMATTER_FLOAT     byte = 'g'
	FORMATTER_PRECISION int  = 17
)

var (
	debugMode = false
)

//Debug 是否开启debug模式，默认:false
func Debug(value bool) {
	debugMode = value
}

//NewPhpObject 一个新的php对象
func NewPhpObject(className string) *PhpObject {
	return &PhpObject{
		className: className,
		members:   PhpArray{},
	}
}

//SerializedDecodeFunc 解序列化函数结构
type SerializedDecodeFunc func(string) (PhpValue, error)

//SerializedEncodeFunc 序列化函数结构
type SerializedEncodeFunc func(PhpValue) (string, error)

//PhpValue php值(interface{})
type PhpValue interface{}

//PhpArray php数组(map[interface{}]interface{})
type PhpArray map[PhpValue]PhpValue

//PhpSlice php集合([]interface{})
type PhpSlice []PhpValue

//PhpObject php对象
type PhpObject struct {
	className string
	members   PhpArray
}

//GetClassName php对象类名
func (s *PhpObject) GetClassName() string {
	return s.className
}

//SetClassName 设置php对象类名
func (s *PhpObject) SetClassName(name string) *PhpObject {
	s.className = name
	return s
}

//GetMembers 获取php对象属性
func (s *PhpObject) GetMembers() PhpArray {
	return s.members
}

//SetMembers 设置php对象属性
func (s *PhpObject) SetMembers(members PhpArray) *PhpObject {
	s.members = members
	return s
}

//GetPrivate 获取Private属性
func (s *PhpObject) GetPrivate(name string) (v PhpValue, ok bool) {
	v, ok = s.members["\x00"+s.className+"\x00"+name]
	return
}

//SetPrivate 设置Private属性
func (s *PhpObject) SetPrivate(name string, value PhpValue) *PhpObject {
	s.members["\x00"+s.className+"\x00"+name] = value
	return s
}

//GetProtected 获取Protected属性
func (s *PhpObject) GetProtected(name string) (v PhpValue, ok bool) {
	v, ok = s.members["\x00*\x00"+name]
	return
}

//SetProtected 设置Protected属性
func (s *PhpObject) SetProtected(name string, value PhpValue) *PhpObject {
	s.members["\x00*\x00"+name] = value
	return s
}

//GetPublic 获取Public属性
func (s *PhpObject) GetPublic(name string) (v PhpValue, ok bool) {
	v, ok = s.members[name]
	return
}

//SetPublic 设置Public属性
func (s *PhpObject) SetPublic(name string, value PhpValue) *PhpObject {
	s.members[name] = value
	return s
}

//NewPhpObjectSerialized php对象序列号
func NewPhpObjectSerialized(className string) *PhpObjectSerialized {
	return &PhpObjectSerialized{
		className: className,
	}
}

//PhpObjectSerialized php对象序列号结构
type PhpObjectSerialized struct {
	className string
	data      string
	value     PhpValue
}

//GetClassName 对象名称
func (s *PhpObjectSerialized) GetClassName() string {
	return s.className
}

//SetClassName 对象名称
func (s *PhpObjectSerialized) SetClassName(name string) *PhpObjectSerialized {
	s.className = name
	return s
}

//GetData 数据
func (s *PhpObjectSerialized) GetData() string {
	return s.data
}

//SetData 数据
func (s *PhpObjectSerialized) SetData(data string) *PhpObjectSerialized {
	s.data = data
	return s
}

//GetValue 值
func (s *PhpObjectSerialized) GetValue() PhpValue {
	return s.value
}

//SetValue 值
func (s *PhpObjectSerialized) SetValue(value PhpValue) *PhpObjectSerialized {
	s.value = value
	return s
}

//NewPhpSplArray SPLArrayObject
func NewPhpSplArray(array, properties PhpValue) *PhpSplArray {
	if array == nil {
		array = make(PhpArray)
	}

	if properties == nil {
		properties = make(PhpArray)
	}

	return &PhpSplArray{
		array:      array,
		properties: properties,
	}
}

//PhpSplArray SPLArrayObject
type PhpSplArray struct {
	flags      int
	array      PhpValue
	properties PhpValue
}

//GetFlags Flag
func (s *PhpSplArray) GetFlags() int {
	return s.flags
}

//SetFlags Flag
func (s *PhpSplArray) SetFlags(value int) {
	s.flags = value
}

//GetArray Array
func (s *PhpSplArray) GetArray() PhpValue {
	return s.array
}

//SetArray Array
func (s *PhpSplArray) SetArray(value PhpValue) {
	s.array = value
}

//GetProperties properties
func (s *PhpSplArray) GetProperties() PhpValue {
	return s.properties
}

//SetProperties properties
func (s *PhpSplArray) SetProperties(value PhpValue) {
	s.properties = value
}
