package serialization

import (
	"gsc/bytestream"
	"reflect"
)

type IDeserializable interface {
	Deserialize(data []byte) []reflect.Value
}

type Deserializable struct {
}

func (deserializable *Deserializable) Deserialize(bytes []byte) []reflect.Value {

	byteReader := bytestream.NewByteReader2(bytes)
	objects := make([]reflect.Value, 0)

	var typeValue uint8
	byteReader.Read(&typeValue)
	kind := reflect.Kind(typeValue)
	length := deserializeValue(&kind, byteReader).Interface().(uint16)

	for i := uint16(0); i < length; i++ {

		byteReader.Read(&typeValue)
		kind = reflect.Kind(typeValue)

		// 反序列值类型
		if data := deserializeValue(&kind, byteReader); data != reflect.ValueOf(nil) {
			objects = append(objects, data)
			continue
		}

		// 反序列化引用类型
		if data := deserializeRef(&kind, byteReader); data != reflect.ValueOf(nil) {
			objects = append(objects, data)
			continue
		}

		// 反序列化切片类型
		if data := deserializeSlice(&kind, byteReader); data != reflect.ValueOf(nil) {
			objects = append(objects, data)
			continue
		}

		// 反序列化映射类型
		if data := deserializeMap(&kind, byteReader); data != reflect.ValueOf(nil) {
			objects = append(objects, data)
			continue
		}

		objects = append(objects, reflect.ValueOf(nil))
	}

	return objects
}

func deserializeValue(kind *reflect.Kind, byteReader *bytestream.ByteReader) reflect.Value {

	generate, ok := GenerateVar[*kind]
	if !ok {
		return reflect.ValueOf(nil)
	}

	obj := generate()
	byteReader.Read(obj)

	return reflect.ValueOf(obj).Elem()
}

func deserializeRef(kind *reflect.Kind, byteReader *bytestream.ByteReader) reflect.Value {

	if *kind != reflect.Ptr {
		return reflect.ValueOf(nil)
	}

	var typeValue uint8
	byteReader.Read(&typeValue)
	*kind = reflect.Kind(typeValue)

	generate, ok := GenerateVar[*kind]
	if !ok {
		return reflect.ValueOf(nil)
	}

	obj := generate()
	byteReader.Read(obj)

	return reflect.ValueOf(obj)
}

func deserializeSlice(kind *reflect.Kind, byteReader *bytestream.ByteReader) reflect.Value {

	if *kind != reflect.Slice {
		return reflect.ValueOf(nil)
	}

	var typeValue uint8
	byteReader.Read(&typeValue)
	*kind = reflect.Kind(typeValue)

	if *kind == reflect.Ptr {

		byteReader.Read(&typeValue)
		*kind = reflect.Kind(typeValue)

		generate, ok := GenerateVar[*kind]
		if !ok {
			return reflect.ValueOf(nil)
		}

		var valueLength uint16
		byteReader.Read(&valueLength)

		varType := reflect.SliceOf(KindPtrType[*kind])
		slice := reflect.MakeSlice(varType, int(valueLength), int(valueLength))

		for i := uint16(0); i < valueLength; i++ {
			obj := generate()
			byteReader.Read(obj)
			slice.Index(int(i)).Set(reflect.ValueOf(obj))
		}
		return slice

	}

	generate, ok := GenerateVar[*kind]
	if !ok {
		return reflect.ValueOf(nil)
	}

	var valueLength uint16
	byteReader.Read(&valueLength)

	varType := reflect.SliceOf(KindType[*kind])
	slice := reflect.MakeSlice(varType, int(valueLength), int(valueLength))

	for i := uint16(0); i < valueLength; i++ {
		obj := generate()
		byteReader.Read(obj)
		slice.Index(int(i)).Set(reflect.ValueOf(obj).Elem())
	}
	return slice
}

func deserializeMap(kind *reflect.Kind, byteReader *bytestream.ByteReader) reflect.Value {
	if *kind != reflect.Map {
		return reflect.ValueOf(nil)
	}

	var typeValue uint8
	byteReader.Read(&typeValue)

	keyKind := reflect.Kind(typeValue)
	keyGenerate, ok := GenerateVar[keyKind]
	if !ok {
		return reflect.ValueOf(nil)
	}

	byteReader.Read(&typeValue)
	valueKind := reflect.Kind(typeValue)

	if valueKind == reflect.Ptr {

		byteReader.Read(&typeValue)
		valueKind = reflect.Kind(typeValue)

		valueGenerate, ok := GenerateVar[valueKind]
		if !ok {
			return reflect.ValueOf(nil)
		}

		var valueLength uint16
		byteReader.Read(&valueLength)

		varType := reflect.MapOf(KindType[keyKind], KindPtrType[valueKind])
		maps := reflect.MakeMap(varType)

		for i := uint16(0); i < valueLength; i++ {
			keyObj := keyGenerate()
			valueObj := valueGenerate()
			byteReader.Read(keyObj)
			byteReader.Read(valueObj)
			maps.SetMapIndex(reflect.ValueOf(keyObj).Elem(), reflect.ValueOf(valueObj).Elem())
		}
		return maps

	}

	valueGenerate, ok := GenerateVar[valueKind]
	if !ok {
		return reflect.ValueOf(nil)
	}

	var valueLength uint16
	byteReader.Read(&valueLength)

	varType := reflect.MapOf(KindType[keyKind], KindType[valueKind])
	maps := reflect.MakeMap(varType)

	for i := uint16(0); i < valueLength; i++ {
		keyObj := keyGenerate()
		valueObj := valueGenerate()
		byteReader.Read(keyObj)
		byteReader.Read(valueObj)
		maps.SetMapIndex(reflect.ValueOf(keyObj).Elem(), reflect.ValueOf(valueObj).Elem())
	}
	return maps
}
