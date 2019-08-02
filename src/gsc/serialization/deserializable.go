package serialization

import (
	"github.com/sf-go/gsf/src/gsc/bytestream"
	"reflect"
)

type IDeserializable interface {
	Deserialize(args ...interface{}) []reflect.Value
}

type Deserializable struct {
	byteReader *bytestream.ByteReader
}

func NewDeserializable(byteReader *bytestream.ByteReader) *Deserializable {
	return &Deserializable{
		byteReader: byteReader,
	}
}

func (deserializable *Deserializable) Deserialize(args ...interface{}) []reflect.Value {

	length := deserializable.deserializeValue(deserializable.byteReader).Interface().(uint8)
	objects := make([]reflect.Value, length)

	for i := uint8(0); i < length; i++ {
		value := deserializable.DeserializeSingle(deserializable.byteReader, args...)
		objects[i] = value
	}

	return objects
}

func (deserializable *Deserializable) DeserializeSingle(byteReader *bytestream.ByteReader,
	args ...interface{}) reflect.Value {

	// 反序列值类型
	if data := deserializable.deserializeValue(byteReader); data != reflect.ValueOf(nil) {
		return data
	}

	// 反序列化引用类型
	if data := deserializable.deserializeRef(byteReader); data != reflect.ValueOf(nil) {
		return data
	}

	// 反序列化切片类型
	if data := deserializable.deserializeSlice(byteReader); data != reflect.ValueOf(nil) {
		return data
	}

	// 反序列化映射类型
	if data := deserializable.deserializeMap(byteReader); data != reflect.ValueOf(nil) {
		return data
	}

	if data := deserializable.deserializeStruct(byteReader, args...); data != reflect.ValueOf(nil) {
		return data
	}
	return reflect.ValueOf(nil)
}

func (deserializable *Deserializable) deserializeValue(byteReader *bytestream.ByteReader) reflect.Value {

	typeValue := uint8(0)
	byteReader.Read(&typeValue)
	kind := reflect.Kind(typeValue)

	if kind == reflect.Struct || kind == reflect.Slice ||
		kind == reflect.Map || kind == reflect.Ptr {
		byteReader.Shift(-1)
		return reflect.ValueOf(nil)
	}

	generate, ok := GenerateVar[kind]
	if !ok {
		return reflect.ValueOf(nil)
	}

	obj := generate(nil)
	byteReader.Read(obj)

	return reflect.ValueOf(obj).Elem()
}

func (deserializable *Deserializable) deserializeRef(byteReader *bytestream.ByteReader) reflect.Value {

	typeValue := uint8(0)
	byteReader.Read(&typeValue)
	kind := reflect.Kind(typeValue)

	if kind != reflect.Ptr {
		byteReader.Shift(-1)
		return reflect.ValueOf(nil)
	}

	byteReader.Read(&typeValue)
	kind = reflect.Kind(typeValue)

	generate, ok := GenerateVar[kind]
	if !ok {
		return reflect.ValueOf(nil)
	}

	obj := generate(nil)
	byteReader.Read(obj)

	return reflect.ValueOf(obj)
}

func (deserializable *Deserializable) deserializeSlice(byteReader *bytestream.ByteReader,
	args ...interface{}) reflect.Value {

	typeValue := uint8(0)
	byteReader.Read(&typeValue)
	kind := reflect.Kind(typeValue)

	if kind != reflect.Slice {
		byteReader.Shift(-1)
		return reflect.ValueOf(nil)
	}

	byteReader.Read(&typeValue)
	kind = reflect.Kind(typeValue)

	if kind == reflect.Ptr {
		return deserializable.deserializeSlicePtr(byteReader, args...)
	}

	byteReader.Shift(-1)
	return deserializable.deserializeSliceValue(byteReader, args...)
}

func (deserializable *Deserializable) deserializeSliceValue(byteReader *bytestream.ByteReader,
	args ...interface{}) reflect.Value {

	typeValue := uint8(0)
	byteReader.Read(&typeValue)
	kind := reflect.Kind(typeValue)

	keyType := KindType[kind]()
	params := make([]interface{}, 0)
	if kind == reflect.Struct {
		name := ""
		byteReader.Read(&name)
		params = append(append(params, name), args...)
		keyType = KindType[kind](params...)
	}

	generate, ok := GenerateVar[kind]

	if !ok {
		return reflect.ValueOf(nil)
	}

	valueLength := uint16(0)
	byteReader.Read(&valueLength)

	varType := reflect.SliceOf(keyType)
	slice := reflect.MakeSlice(varType, int(valueLength), int(valueLength))

	for i := uint16(0); i < valueLength; i++ {
		if kind == reflect.Struct {
			byteReader.Shift(2)
			slice.Index(int(i)).Set(deserializable.deserializeStruct(byteReader, args...).Elem())
		} else {
			obj := generate(params...)
			byteReader.Read(obj)
			slice.Index(int(i)).Set(reflect.ValueOf(obj).Elem())
		}
	}
	return slice
}

func (deserializable *Deserializable) deserializeSlicePtr(byteReader *bytestream.ByteReader,
	args ...interface{}) reflect.Value {

	typeValue := uint8(0)
	byteReader.Read(&typeValue)
	kind := reflect.Kind(typeValue)

	keyType := reflect.TypeOf(nil)
	params := make([]interface{}, 0)
	if kind == reflect.Struct {
		name := ""
		byteReader.Read(&name)
		params = append(append(params, name), args...)
		keyType = KindPtrType[kind](params...)
	} else {
		keyType = KindPtrType[kind]()
	}

	generate, ok := GenerateVar[kind]

	if !ok {
		return reflect.ValueOf(nil)
	}

	valueLength := uint16(0)
	byteReader.Read(&valueLength)

	varType := reflect.SliceOf(keyType)
	slice := reflect.MakeSlice(varType, int(valueLength), int(valueLength))

	for i := uint16(0); i < valueLength; i++ {
		if kind == reflect.Struct {
			byteReader.Shift(2)
			slice.Index(int(i)).Set(deserializable.deserializeStruct(byteReader, args...))
		} else {
			obj := generate(params...)
			byteReader.Read(obj)
			slice.Index(int(i)).Set(reflect.ValueOf(obj))
		}
	}
	return slice
}

func (deserializable *Deserializable) deserializeMap(byteReader *bytestream.ByteReader,
	args ...interface{}) reflect.Value {

	typeValue := uint8(0)
	byteReader.Read(&typeValue)
	kind := reflect.Kind(typeValue)

	if kind != reflect.Map {
		byteReader.Shift(-1)
		return reflect.ValueOf(nil)
	}

	byteReader.Read(&typeValue)
	valueKind := reflect.Kind(typeValue)

	if valueKind == reflect.Ptr {
		return deserializable.deserializeMapPtrStruct(byteReader, args...)
	}

	byteReader.Shift(-1)
	return deserializable.deserializeMapValueStruct(byteReader, args...)
}

func (deserializable *Deserializable) deserializeMap2(byteReader *bytestream.ByteReader) reflect.Value {

	typeValue := uint8(0)
	byteReader.Read(&typeValue)
	kind := reflect.Kind(typeValue)

	if kind != reflect.Map {
		byteReader.Shift(-1)
		return reflect.ValueOf(nil)
	}

	byteReader.Read(&typeValue)

	keyKind := reflect.Kind(typeValue)
	keyGenerate, ok := GenerateVar[keyKind]
	if !ok {
		return reflect.ValueOf(nil)
	}

	byteReader.Read(&typeValue)
	valueKind := reflect.Kind(typeValue)

	keyType := KindType[keyKind]()
	valueType := KindType[valueKind]()

	if valueKind == reflect.Ptr {
		byteReader.Read(&typeValue)
		valueKind = reflect.Kind(typeValue)

		keyType = KindType[keyKind]()
		valueType = KindPtrType[valueKind]()
	}

	valueGenerate, ok := GenerateVar[valueKind]
	if !ok {
		return reflect.ValueOf(nil)
	}

	valueLength := uint16(0)
	byteReader.Read(&valueLength)

	varType := reflect.MapOf(keyType, valueType)
	maps := reflect.MakeMap(varType)

	for i := uint16(0); i < valueLength; i++ {
		keyObj := keyGenerate(nil)
		valueObj := valueGenerate(nil)
		byteReader.Read(keyObj)
		byteReader.Read(valueObj)
		maps.SetMapIndex(reflect.ValueOf(keyObj).Elem(), reflect.ValueOf(valueObj).Elem())
	}
	return maps
}

func (deserializable *Deserializable) deserializeMapValueStruct(byteReader *bytestream.ByteReader,
	args ...interface{}) reflect.Value {
	typeValue := uint8(0)
	byteReader.Read(&typeValue)
	valueKind := reflect.Kind(typeValue)

	valueGenerate, ok := GenerateVar[valueKind]
	if !ok {
		return reflect.ValueOf(nil)
	}

	params := make([]interface{}, 0)
	valueType := reflect.TypeOf(nil)
	if valueKind == reflect.Struct {
		name := ""
		byteReader.Read(&name)
		params = append(append(params, name), args...)
		valueType = KindType[valueKind](params...)
	} else {
		valueType = KindType[valueKind]()
	}

	byteReader.Read(&typeValue)
	keyKind := reflect.Kind(typeValue)
	keyType := KindType[keyKind]()
	keyGenerate, ok := GenerateVar[keyKind]
	if !ok {
		return reflect.ValueOf(nil)
	}

	valueLength := uint16(0)
	byteReader.Read(&valueLength)

	varType := reflect.MapOf(keyType, valueType)
	maps := reflect.MakeMap(varType)

	for i := uint16(0); i < valueLength; i++ {
		keyObj := keyGenerate()
		byteReader.Read(keyObj)

		if valueKind == reflect.Struct {
			byteReader.Shift(2)
			maps.SetMapIndex(reflect.ValueOf(keyObj).Elem(), deserializable.deserializeStruct(byteReader, args...).Elem())
		} else {
			valueObj := valueGenerate(params...)
			byteReader.Read(valueObj)
			maps.SetMapIndex(reflect.ValueOf(keyObj).Elem(), reflect.ValueOf(valueObj).Elem())
		}
	}

	return maps
}

func (deserializable *Deserializable) deserializeMapPtrStruct(byteReader *bytestream.ByteReader,
	args ...interface{}) reflect.Value {
	typeValue := uint8(0)
	byteReader.Read(&typeValue)
	valueKind := reflect.Kind(typeValue)

	valueGenerate, ok := GenerateVar[valueKind]
	if !ok {
		return reflect.ValueOf(nil)
	}

	params := make([]interface{}, 0)
	valueType := reflect.TypeOf(nil)
	if valueKind == reflect.Struct {
		name := ""
		byteReader.Read(&name)
		params = append(append(params, name), args...)
		valueType = KindType[valueKind](params...)
	} else {
		valueType = KindType[valueKind]()
	}

	byteReader.Read(&typeValue)
	keyKind := reflect.Kind(typeValue)
	keyType := KindType[keyKind]()
	keyGenerate, ok := GenerateVar[keyKind]
	if !ok {
		return reflect.ValueOf(nil)
	}

	valueLength := uint16(0)
	byteReader.Read(&valueLength)

	varType := reflect.MapOf(keyType, valueType)
	maps := reflect.MakeMap(varType)

	for i := uint16(0); i < valueLength; i++ {
		keyObj := keyGenerate()
		byteReader.Read(keyObj)

		if valueKind == reflect.Struct {
			byteReader.Shift(2)
			maps.SetMapIndex(reflect.ValueOf(keyObj).Elem(), deserializable.deserializeStruct(byteReader, args...))
		} else {
			valueObj := valueGenerate(params...)
			byteReader.Read(valueObj)
			maps.SetMapIndex(reflect.ValueOf(keyObj).Elem(), reflect.ValueOf(valueObj))
		}
	}

	return maps
}

func (deserializable *Deserializable) deserializeStruct(byteReader *bytestream.ByteReader,
	args ...interface{}) reflect.Value {

	typeValue := uint8(0)
	byteReader.Read(&typeValue)
	kind := reflect.Kind(typeValue)

	if kind != reflect.Struct {
		return reflect.ValueOf(nil)
	}

	name := ""
	byteReader.Read(&name)

	valueGenerate, ok := GenerateVar[reflect.Struct]
	if !ok {
		return reflect.ValueOf(nil)
	}

	value := valueGenerate(append(append(make([]interface{}, 0), name), args...)...)
	packet := value.(ISerializablePacket)

	length := uint16(0)
	byteReader.Read(&length)

	deser := NewDeserializable(byteReader)
	packet.FromBinaryReader(deser)

	return reflect.ValueOf(packet)
}
