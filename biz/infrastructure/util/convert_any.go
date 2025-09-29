package util

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func Anypb2Any(req map[string]*anypb.Any) (map[string]any, error) {
	res := make(map[string]any)
	for k, v := range req {
		msg, err := v.UnmarshalNew()
		if err != nil {
			return nil, err
		}

		switch m := msg.(type) {
		case *wrapperspb.StringValue:
			res[k] = m.Value
		case *wrapperspb.Int32Value:
			res[k] = m.Value
		case *wrapperspb.Int64Value:
			res[k] = m.Value
		case *wrapperspb.FloatValue:
			res[k] = m.Value
		case *wrapperspb.DoubleValue:
			res[k] = m.Value
		case *wrapperspb.BoolValue:
			res[k] = m.Value
		}
	}
	return res, nil
}

func Any2Anypb(req map[string]any) (map[string]*anypb.Any, error) {
	res := make(map[string]*anypb.Any)

	for k, v := range req {
		anyVal, err := Wrap(v)
		if err != nil {
			return nil, err
		}
		res[k] = anyVal
	}

	return res, nil
}

func Wrap(v any) (*anypb.Any, error) {
	var msg proto.Message
	switch val := v.(type) {
	case string:
		msg = wrapperspb.String(val)
	case int:
		msg = wrapperspb.Int64(int64(val))
	case int32:
		msg = wrapperspb.Int32(val)
	case int64:
		msg = wrapperspb.Int64(val)
	case float32:
		msg = wrapperspb.Float(val)
	case float64:
		msg = wrapperspb.Double(val)
	case bool:
		msg = wrapperspb.Bool(val)
	}
	return anypb.New(msg)
}
