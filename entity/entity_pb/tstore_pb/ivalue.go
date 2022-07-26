package tstore_pb

import (
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func NewStrValue(str string) *IValue {
	return &IValue{
		StrValue: str,
		Itype:    IValue_str,
	}
}

func NewIntValue(i int32) *IValue {
	return &IValue{
		IntValue: i,
		Itype:    IValue_int,
	}
}

func NewMapValue(m map[string]interface{}) *IValue {
	v := map[string]*IValue{}

	for key, value := range m {
		v[key] = NewIValue(value)
	}

	return &IValue{
		MapValue: v,
		Itype:    IValue_map,
	}
}

func NewAnyValue(m proto.Message) *IValue {
	any, err := anypb.New(m)
	if err != nil {
		log.Error().Msgf("NewAnyValue error, %s", err)
	}

	return &IValue{
		AnyValue: any,
		Itype:    IValue_any,
	}
}

func NewIValue(value interface{}) *IValue {
	switch value.(type) {
	case int32:
		return NewIntValue(value.(int32))
	case string:
		return NewStrValue(value.(string))
	case map[string]interface{}:
		return NewMapValue(value.(map[string]interface{}))
	case proto.Message:
		return NewAnyValue(value.(proto.Message))
	default:
		log.Error().Msgf("new ivalue, unsupported type, value = %v", value)
	}

	return &IValue{
		Itype: IValue_nil,
	}
}

func (x *IValue) MapSet(key string, v *IValue) {
	if x.Itype != IValue_map {
		log.Error().Msgf("MapSet error, self is not a map value, key = %s, value = %v", key, v)
		return
	}

	x.MapValue[key] = v
}

func (x *IValue) MapGet(key string) (*IValue, bool) {
	if x.Itype != IValue_map {
		log.Error().Msgf("MapSet error, self is not a map value")
		return &IValue{Itype: IValue_nil}, false
	}

	v, ok := x.MapValue[key]
	return v, ok
}
