package grid

import (
	"errors"
	"fmt"
	"html"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

const CallbackPrefix = "GridCallback"

// cbk is holding the name of the callback, the callback function and arguments.
type cbk struct {
	name string
	fn   reflect.Value
	args []interface{}
}

// dots return the number of dots in the name.
func (c *cbk) dots() int {
	return strings.Count(c.name, ".")
}

// checkDotNotation checks if in the given string there is a {{field}}} with a dot-notation.
// it will return an error if the dotNotations mismatch within the format ex {{ID}} {{Parts.Name}} is not allowed only {{Parts.ID}} {{Parts.Name}} for example or {{ID}} {{Name}}.
func checkDotNotation(format string) (bool, []string, error) {

	normalField := false
	hasDotNotation := false
	dotNotationLen := 0
	var dotNotation []string

	re, _ := regexp.Compile("{{.*?}}")
	for _, match := range re.FindAllString(format, -1) {
		match = strings.Replace(match, "{{", "", -1) // TODO better solution :D
		match = strings.Replace(match, "}}", "", -1) // TODO better solution :D

		relField := strings.Split(match, ".")
		if (normalField && hasDotNotation) || (hasDotNotation && len(relField) != dotNotationLen) {
			return true, nil, errors.New("callback: dotNotation mismatch")
		}

		if len(relField) > 1 {
			if !hasDotNotation {
				hasDotNotation = true
				dotNotationLen = len(relField)
			}
			dotNotation = append(dotNotation, match)
		} else {
			normalField = true
			dotNotation = append(dotNotation, match)
		}
	}

	return hasDotNotation, dotNotation, nil
}

func helperDecorator(format string, matchedFields []string, data map[string]interface{}) string {

	rv := reflect.ValueOf(data)
	relations := strings.Split(matchedFields[0], ".")
	result := format

	for i := 0; i < len(relations)-1; i++ {
		if reflect.TypeOf(rv.Interface()).Kind() == reflect.Slice {
			rv = reflect.ValueOf(rv.Interface())
			if rv.Len() > 0 {
				for k := range matchedFields {
					rel := strings.Split(matchedFields[k], ".")
					format = strings.Replace(format, "{{"+strings.Join(rel, ".")+"}}", "{{"+strings.Join(rel[i+1:], ".")+"}}", 1)
					matchedFields[k] = strings.Join(rel[i+1:], ".")
				}

				var res []string
				for n := 0; n < rv.Len(); n++ {
					res = append(res, helperDecorator(format, matchedFields, reflect.ValueOf(rv.Index(n).Interface()).MapIndex(reflect.ValueOf(relations[i])).Interface().(map[string]interface{})))
				}
				return strings.Join(res, "")
			}
			return ""
		} else {
			rv = rv.MapIndex(reflect.ValueOf(relations[i]))
		}
	}

	// zero value
	if !rv.IsValid() {
		return ""
	}

	var res []string
	rv = reflect.ValueOf(rv.Interface())
	switch rv.Type().Kind() {
	case reflect.Slice:
		// loop through the result
		for i := 0; i < rv.Len(); i++ {
			result = format
			item := reflect.ValueOf(rv.Index(i).Interface())

			// loop through the placeholder fields
			for _, field := range matchedFields {
				tmpField := strings.Split(field, ".")
				val := item.MapIndex(reflect.ValueOf(tmpField[len(tmpField)-1]))
				if val.IsValid() {
					result = strings.Replace(result, "{{"+field+"}}", fmt.Sprintf("%v", val), -1)
				}
			}
			res = append(res, result)
		}
	case reflect.Map:
		for _, field := range matchedFields {
			tmpField := strings.Split(field, ".")
			val := rv.MapIndex(reflect.ValueOf(tmpField[len(tmpField)-1]))
			if val.IsValid() {
				result = strings.Replace(result, "{{"+field+"}}", fmt.Sprintf("%v", val), -1)
			}
		}
		res = append(res, result)

	default:
		res = append(res, "callback: Decorator implement unknown type ", rv.Type().Kind().String())
	}
	result = strings.Join(res, "")

	return result
}

// Decorator is a predefined helper function for the grid callbacks.
// the first param is the decorator format, the second a boolean if the text should get escaped, the third is the result row.
// Usage:	field.SetCallback(grid.Decorator,"{{Parts.ID}} - {{Parts.Name}}",false).
func Decorator(format string, htmlEscape bool, data map[string]interface{}) string {

	dn, matchedFields, err := checkDotNotation(format)
	if err != nil {
		return "Error: " + err.Error()
	}

	// result
	result := format

	// if there are single fields
	if !dn {
		for _, field := range matchedFields {
			if val, ok := data[field]; ok {
				result = strings.Replace(result, "{{"+field+"}}", fmt.Sprintf("%v", val), -1)
			}
		}
	}

	// if there are dotNotations
	if dn {
		result = helperDecorator(format, matchedFields, data)
	}

	if htmlEscape {
		return html.EscapeString(result)
	}

	return result
}

// recursiveCallbackList returns recursive all callbacks in the struct.
// it gets sorted by dot notation, the deepest child will get at the beginning of the slice.
func recursiveCallbackList(fields map[string]Interface, parent string) []cbk {
	var callbacks []cbk

	prefix := ""
	if parent != "" {
		prefix = parent + "."
	}

	for k, f := range fields {

		// skip removed fields
		if f.getRemove() {
			continue
		}

		// check if callback exists
		if f.getCallback().IsValid() {
			callbacks = append(callbacks, cbk{name: prefix + k, fn: f.getCallback(), args: f.getCallbackArgs()})
		}

		if f.getFields() != nil {
			// recursive callbacks
			child := recursiveCallbackList(f.getFields(), prefix+f.getFieldName())
			callbacks = append(callbacks, child...)
		}
	}

	if parent == "" {
		sort.Slice(callbacks, func(i, j int) bool {
			return callbacks[i].dots() > callbacks[j].dots()
		})
	}

	return callbacks
}

// callbackFromString returns the callback by a string.
// if its a slice, multiple callbacks will return.
func callbackFromString(fn string, row reflect.Value) []reflect.Value {
	var callbacks []reflect.Value
	relField := strings.Split(fn, ".")

	// no dot notation
	if len(relField) == 1 {
		return append(callbacks, row.Addr().MethodByName(CallbackPrefix+fn))
	}

	// dot notation
	lastField := row
	for i, f := range relField {
		if i < len(relField)-1 {
			lastField = reflect.Indirect(lastField).FieldByName(f)
		} else {
			if lastField.Kind() == reflect.Slice {
				for n := 0; n < lastField.Len(); n++ {
					callbacks = append(callbacks, lastField.Index(n).Addr().MethodByName(CallbackPrefix+f))
				}
			} else {
				callbacks = append(callbacks, lastField.Addr().MethodByName(CallbackPrefix+f))
			}
		}
	}

	return callbacks
}

// setCallbackValueByString will set the value to the correct position in the map.
// if the return value is compatible with the struct, the value will also get set to the struct. This is needed that for example the Parts callback will get the correct callback value of Parts.Name.
func setCallbackValueByString(field string, callbacks []reflect.Value, cindex int, arguments []reflect.Value, row map[string]interface{}, structValue reflect.Value) error {
	relField := strings.Split(field, ".")

	// no dot notation
	if len(relField) == 1 {
		// arguments mismatch
		if callbacks[cindex].Type().NumIn() != len(arguments) {

			fmt.Println(cindex, callbacks[cindex].Type().NumIn(), len(arguments), arguments)
			return errors.New("callback: arguments does not fit")
		}

		// set map value
		value := callbacks[cindex].Call(arguments)[0]

		// set the value to the row, if there is a json name, the key will be the json name instead of the struct name.
		if _, ok := row[field]; ok {
			row[field] = value.Interface()
		} else {
			if sf, ok := reflect.TypeOf(structValue.Interface()).FieldByName(field); ok {
				jsonTagName := jsonTagName(sf.Tag.Get("json"))
				if jsonTagName != "" {
					row[jsonTagName] = value.Interface()
				}
			}
		}

		// set struct value for chained callbacks
		if structValue.FieldByName(field).IsValid() &&
			structValue.FieldByName(field).CanSet() &&
			callbacks[cindex].Type().Out(0).Kind() == structValue.FieldByName(field).Kind() {
			structValue.FieldByName(field).Set(value)
		}
		return nil
	}

	// dot notation
	last := row
	for i := 0; i < len(relField)-1; i++ {
		tmp := reflect.ValueOf(last[relField[i]])
		switch tmp.Type().Kind() {
		case reflect.Slice:
			for n := 0; n < tmp.Len(); n++ {
				err := setCallbackValueByString(strings.Join(relField[i+1:], "."), callbacks, n, arguments, tmp.Index(n).Interface().(map[string]interface{}), structValue.FieldByName(relField[i]).Index(n))
				if err != nil {
					return err
				}
			}
		case reflect.Map:
			err := setCallbackValueByString(strings.Join(relField[i+1:], "."), callbacks, 0, arguments, reflect.ValueOf(last[relField[i]]).Interface().(map[string]interface{}), structValue.FieldByName(relField[i]))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// callback will convert the struct to a map and manipulates the result by the given callbacks.
func (g *Grid) callback(res interface{}) ([]map[string]interface{}, error) {

	var newResult []map[string]interface{}
	headFields := headerFieldsLoop(g.fields, false)
	callbacks := recursiveCallbackList(g.fields, "")

	if len(callbacks) > 0 {
		s := reflect.Indirect(reflect.ValueOf(res))
		for i := 0; i < s.Len(); i++ { // loop through rows

			row, err := MapFromModel(s.Index(i).Interface(), headFields)
			if err != nil {
				return nil, err
			}

			for _, c := range callbacks {

				var in []reflect.Value
				var cfn []reflect.Value
				switch c.fn.Kind() {
				case reflect.String:
					in = make([]reflect.Value, len(c.args))
					cfn = callbackFromString(c.name, s.Index(i))
				case reflect.Func:
					in = make([]reflect.Value, len(c.args)+1)
					in[len(c.args)] = reflect.ValueOf(row) // todo set an option if callbacks should be passed through or if every callback should get the raw data. -> MapFromModel(s.Index(i).Interface(), headFields)
					cfn = append(cfn, c.fn)
				}

				// calling function
				if c.fn.IsValid() {
					// adding arguments
					if len(c.args) > 0 {
						for k, arg := range c.args {
							if in != nil {
								in[k] = reflect.ValueOf(arg)
							}
						}
					}

					err := setCallbackValueByString(c.name, cfn, 0, in, row, s.Index(i))
					if err != nil {
						return nil, err
					}
				} else {
					if c.fn.Kind() == reflect.String {
						return nil, errors.New("callback: method does not exist")
					}
					return nil, errors.New("callback: method does not exist")
				}
			}

			newResult = append(newResult, row)
		}
	}

	return newResult, nil
}
