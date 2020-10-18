package orm

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"
)

type NullBool struct {
	sql.NullBool
}

func NewNullBool(b bool) NullBool {
	return NullBool{NullBool: sql.NullBool{b, true}}
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports number and null input.
// 0 will not be considered a null Bool.
func (b *NullBool) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		b.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &b.Bool); err != nil {
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	b.Valid = true
	return nil
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this Bool is null.
func (b NullBool) MarshalJSON() ([]byte, error) {
	if !b.Valid {
		return []byte("null"), nil
	}
	if !b.Bool {
		return []byte("false"), nil
	}
	return []byte("true"), nil
}

type NullInt struct {
	sql.NullInt64
}

func NewNullInt(i int) NullInt {
	return NullInt{NullInt64: sql.NullInt64{int64(i), true}}
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports number and null input.
// 0 will not be considered a null Int.
// It also supports unmarshalling a sql.NullInt64.
func (i *NullInt) UnmarshalJSON(data []byte) error {

	var err error
	var v interface{}
	if err = json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch x := v.(type) {
	case float64:
		// Unmarshal again, directly to int64, to avoid intermediate float64
		err = json.Unmarshal(data, &i.Int64)
	case string:
		str := string(x)
		if len(str) == 0 {
			i.Valid = false
			return nil
		}
		i.Int64, err = strconv.ParseInt(str, 10, 64)
	case map[string]interface{}:
		err = json.Unmarshal(data, &i.NullInt64)
	case nil:
		i.Valid = false
		return nil
	default:
		err = fmt.Errorf("json: cannot unmarshal %v into Go value of type null.Int", reflect.TypeOf(v).Name())
	}
	i.Valid = err == nil
	return err
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this Int is null.
func (i NullInt) MarshalJSON() ([]byte, error) {
	if !i.Valid {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatInt(i.Int64, 10)), nil
}

var nullBytes = []byte("null")

type NullFloat struct {
	sql.NullFloat64
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports number and null input.
// 0 will not be considered a null Float.
func (f *NullFloat) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		f.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &f.Float64); err != nil {
		var typeError *json.UnmarshalTypeError
		if errors.As(err, &typeError) {
			// special case: accept string input
			if typeError.Value != "string" {
				return fmt.Errorf("null: JSON input is invalid type (need float or string): %w", err)
			}
			var str string
			if err := json.Unmarshal(data, &str); err != nil {
				return fmt.Errorf("null: couldn't unmarshal number string: %w", err)
			}
			n, err := strconv.ParseFloat(str, 64)
			if err != nil {
				return fmt.Errorf("null: couldn't convert string to float: %w", err)
			}
			f.Float64 = n
			f.Valid = true
			return nil
		}
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	f.Valid = true
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null Float if the input is blank.
// It will return an error if the input is not an integer, blank, or "null".
func (f *NullFloat) UnmarshalText(text []byte) error {
	str := string(text)
	if str == "" || str == "null" {
		f.Valid = false
		return nil
	}
	var err error
	f.Float64, err = strconv.ParseFloat(string(text), 64)
	if err != nil {
		return fmt.Errorf("null: couldn't unmarshal text: %w", err)
	}
	f.Valid = true
	return err
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this Float is null.
func (f NullFloat) MarshalJSON() ([]byte, error) {
	if !f.Valid {
		return []byte("null"), nil
	}
	if math.IsInf(f.Float64, 0) || math.IsNaN(f.Float64) {
		return nil, &json.UnsupportedValueError{
			Value: reflect.ValueOf(f.Float64),
			Str:   strconv.FormatFloat(f.Float64, 'g', -1, 64),
		}
	}
	return []byte(strconv.FormatFloat(f.Float64, 'f', -1, 64)), nil
}

// MarshalText implements encoding.TextMarshaler.
// It will encode a blank string if this Float is null.
func (f NullFloat) MarshalText() ([]byte, error) {
	if !f.Valid {
		return []byte{}, nil
	}
	return []byte(strconv.FormatFloat(f.Float64, 'f', -1, 64)), nil
}

type NullString struct {
	sql.NullString
}

func NewNullString(s string) NullString {
	return NullString{NullString: sql.NullString{Valid: true, String: s}}
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports string and null input. Blank string input does not produce a null String.
// It also supports unmarshalling a sql.NullString.
func (s *NullString) UnmarshalJSON(data []byte) error {
	var err error
	var v interface{}
	if err = json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch x := v.(type) {
	case string:
		s.String = x
	case map[string]interface{}:
		err = json.Unmarshal(data, &s.NullString)
	case nil:
		s.Valid = false
		return nil
	default:
		err = fmt.Errorf("json: cannot unmarshal %v into Go value of type null.String", reflect.TypeOf(v).Name())
	}
	s.Valid = err == nil
	return err
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this String is null.
func (s NullString) MarshalJSON() ([]byte, error) {
	if !s.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(s.String)
}

// MarshalText implements encoding.TextMarshaler.
// It will encode a blank string when this String is null.
func (s NullString) MarshalText() ([]byte, error) {
	if !s.Valid {
		return []byte{}, nil
	}
	return []byte(s.String), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null String if the input is a blank string.
func (s *NullString) UnmarshalText(text []byte) error {
	s.String = string(text)
	s.Valid = s.String != ""
	return nil
}

type NullTime struct {
	sql.NullTime
}

func NewNullTime(t time.Time) NullTime {
	return NullTime{NullTime: sql.NullTime{t, true}}
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this time is null.
func (t NullTime) MarshalJSON() ([]byte, error) {
	if !t.Valid {
		return []byte("null"), nil
	}
	return t.Time.MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports string, object (e.g. pq.NullTime and friends)
// and null input.
func (t *NullTime) UnmarshalJSON(data []byte) error {
	var err error
	var v interface{}
	if err = json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch x := v.(type) {
	case string:
		err = t.Time.UnmarshalJSON(data)
	case map[string]interface{}:
		ti, tiOK := x["Time"].(string)
		valid, validOK := x["Valid"].(bool)
		if !tiOK || !validOK {
			return fmt.Errorf(`json: unmarshalling object into Go value of type null.Time requires key "Time" to be of type string and key "Valid" to be of type bool; found %T and %T, respectively`, x["Time"], x["Valid"])
		}
		err = t.Time.UnmarshalText([]byte(ti))
		t.Valid = valid
		return err
	case nil:
		t.Valid = false
		return nil
	default:
		err = fmt.Errorf("json: cannot unmarshal %v into Go value of type null.Time", reflect.TypeOf(v).Name())
	}
	t.Valid = err == nil
	return err
}

func SanitizeToString(i interface{}) (string, error) {
	v, err := SanitizeValue(i)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", v), nil
}

//SanitizeValue is converting any int and nullint to int64 and nullString to string.
func SanitizeValue(i interface{}) (interface{}, error) {

	// ints or null int will be converted to int64
	switch i.(type) {
	case int, int8, int16, int32, int64:
		if int, ok := i.(int); ok {
			return int64(int), nil
		}
		return nil, fmt.Errorf("can not sanitize to int64 %v", i)
	case string:
		return i, nil
	case NullInt:
		if i.(NullInt).Valid {
			return i.(NullInt).Int64, nil
		}
		return 0, nil
	case NullString:
		if i.(NullString).Valid {
			return i.(NullString).String, nil
		}
		return nil, fmt.Errorf("can not sanitize nullString to string %v", i)
	}

	return nil, fmt.Errorf("can not sanitize because of unimplemented type %v", reflect.TypeOf(i))
}
