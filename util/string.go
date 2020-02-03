package util

import (
	"reflect"
	"strings"
	"unicode"
)

func IsEmpty(x interface{}) bool {

	if reflect.TypeOf(x) == nil {
		return true
	}

	return x == reflect.Zero(reflect.TypeOf(x)).Interface()

}

func SliceValueExist(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

func StructName(myvar interface{}, pointerInfo bool) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		rv := ""
		if pointerInfo == true {
			rv += "*"
		}
		return rv + t.Elem().Name()
	} else {
		return t.Name()
	}
}

func SnakeCaseToCamelCase(s string) string {
	return strings.Replace(strings.Title(strings.Replace(strings.ToLower(s), "_", " ", -1)), " ", "", -1)
}

//copyright https://github.com/asaskevich/govalidator/blob/master/utils.go#L107-L119
func CamelCaseToSnakeCase(str string) string {
	var output []rune
	var segment []rune
	for _, r := range str {
		if !unicode.IsLower(r) && string(r) != "_" {
			output = addSegment(output, segment)
			segment = nil
		}
		segment = append(segment, unicode.ToLower(r))
	}
	output = addSegment(output, segment)
	return string(output)
}

func addSegment(inrune, segment []rune) []rune {
	if len(segment) == 0 {
		return inrune
	}
	if len(inrune) != 0 {
		inrune = append(inrune, '_')
	}
	inrune = append(inrune, segment...)
	return inrune
}
