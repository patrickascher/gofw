package grid

import (
	"errors"
	"reflect"
	"strings"
)

type Struct struct {
	model      reflect.Value
	headFields []*head
}

func MapFromModel(res interface{}, hFields []*head) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	strctVal := reflect.Indirect(reflect.ValueOf(res))
	if strctVal.Kind() != reflect.Struct {
		return nil, errors.New("map: not a struct")
	}

	strct := Struct{model: strctVal, headFields: hFields}
	fields := strct.structFields()
	for _, field := range fields {
		var finalVal interface{}
		fieldName := field.Name
		val := strct.model.FieldByName(field.Name)

		finalVal = strct.nested(val, strct.headFieldsByName(field.Name))

		// check json tag naming
		if f, ok := strctVal.Type().FieldByName(field.Name); ok {
			jsonTagName := jsonTagName(f.Tag.Get("json"))
			if jsonTagName != "" {
				fieldName = jsonTagName
			}
		}

		out[fieldName] = finalVal
	}

	return out, nil
}

func jsonTagName(jsonTag string) string {
	fieldName := ""
	if jsonTag != "" && jsonTag != "-" {
		if commaIdx := strings.Index(jsonTag, ","); commaIdx != 0 {
			if commaIdx > 0 {
				fieldName = jsonTag[:commaIdx]
			} else {
				fieldName = jsonTag
			}
		}
	}
	return fieldName
}

func (s *Struct) headFieldsByName(field string) []*head {
	for _, hf := range s.headFields {
		if hf.FieldName == field {
			return hf.Fields
		}
	}
	return nil
}

func (s *Struct) nested(val reflect.Value, fields []*head) interface{} {
	var finalVal interface{}
	if !val.IsValid() {
		return finalVal
	}

	v := reflect.Indirect(reflect.ValueOf(val.Interface()))
	switch v.Kind() {
	case reflect.Slice:

		slices := make([]interface{}, val.Len())
		for x := 0; x < val.Len(); x++ {
			slices[x] = s.nested(val.Index(x), fields)
		}
		finalVal = slices
	case reflect.Struct:
		// handling null. types
		m := val.MethodByName("Value")
		if m.IsValid() {
			in := make([]reflect.Value, 0)
			finalVal = m.Call(in)[0].Interface()
		} else {
			n, _ := MapFromModel(val.Interface(), fields)
			finalVal = n
		}
	default:
		finalVal = val.Interface()
	}
	return finalVal
}

func (s *Struct) structFields() []reflect.StructField {
	t := s.model.Type()
	var f []reflect.StructField

	for _, field := range s.headFields {
		field, _ := t.FieldByName(field.FieldName)
		f = append(f, field)
	}

	return f
}
