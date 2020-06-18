package orm

import (
	driverI "database/sql/driver"
	"fmt"
	"reflect"
	"sort"
	"strings"

	valid "github.com/go-playground/validator"
	"github.com/patrickascher/gofw/sqlquery/types"
)

const (
	TagValidate        = "validate"
	validatorSeparator = ","
	validatorSkip      = "-"
	validatorOr        = "|"
)

var errValidation = "orm: validation failed '%s' for '%s' on the '%s' tag (allowed:%v given:%v)"

// validator struct
type validator struct {
	Config string
}

// appendConfig adds a config if the validation key does not exist yet.
// if one or more keys are duplicated, they will be skipped.
func (v *validator) appendConfig(config string) {

	actualKeys := v.validationKeys(v.Config)
	newKeys := v.validationKeys(config)

	// append config
	var c []string
loop:
	for nk, nv := range newKeys {
		// skip if the key already exists
		for k := range actualKeys {
			if k == nk {
				continue loop
			}
		}
		if nv == "" {
			c = append(c, nk)
		} else {
			c = append(c, nk+"="+nv)
		}
	}

	// set the new config string
	if len(c) > 0 {
		if v.Config != "" {
			v.Config += validatorSeparator
		}
		v.Config += strings.Join(c, validatorSeparator)
	}
}

// validationKeys return all defined config keys.
func (*validator) validationKeys(config string) map[string]string {
	keys := map[string]string{}

	a := strings.FieldsFunc(config, split)
	for _, k := range a {
		key := k
		value := ""
		if strings.Contains(k, "=") {
			key = strings.Split(k, "=")[0]
			value = strings.Split(k, "=")[1]
		}
		key = strings.Trim(key, " ")
		keys[key] = value
	}

	return keys
}

// split the validation by 'validatorSeparator' and 'validatorOr' tag.
func split(r rune) bool {
	return r == []rune(validatorSeparator)[0] || r == []rune(validatorOr)[0]
}

// skipByTag - skips the validation if that is defined by tag.
func (v *validator) skipByTag() bool {
	return strings.Contains(v.Config, validatorSkip)
}

// sort checks that omitempty is always on the first place.
func (v *validator) sort() {
	// sort the config list, that omitempty is always on the first place.
	list := strings.Split(v.Config, validatorSeparator)
	sort.Slice(list, func(i, j int) bool {
		x := 0
		y := 0
		if list[i] == "omitempty" {
			x = 1
		}
		if list[j] == "ommitempty" {
			y = 1
		}
		return x > y
	})
	v.Config = strings.Join(list, validatorSeparator)
}

// addDBValidation will add the following validations:
// - belongsTo relations will be set to omitempty (no required tag will be added) because they are added dynamically after isValid is called on the main orm.
// - columns which does not allow NULL will be required. Except if the column is an autoincrement field, then omitempty is added.
// - Integer: numeric (min,max)
// - Float: numeric
// - Text: size (max)
// - TextArea: size (max)
// - Select: oneof (list items)
// TODO Date,Timestamp,DateTime
func (m *Model) addDBValidation() error {

	writePerm := Permission{Write: true}

	for _, f := range m.scope.Fields(writePerm) {

		// check if skip tag exists
		if f.Validator.skipByTag() {
			continue
		}

		// the belongsTo fk is set dynamically, but the validation function is called before.
		// the belongsTo fk is allowed to be empty in that case.
		// no Permission is set, because the user could have disabled the write permission of the belongsTo relation and so the fk field would not be set to omitempty.
		// TODO the real value of the foreign key is never checked in that case. This must happen in the strategy for it?
		isBelongsTo := false
		for _, relation := range m.scope.Relations(Permission{}) {
			if relation.Kind == BelongsTo && relation.ForeignKey.Name == f.Name {
				isBelongsTo = true
				f.Validator.appendConfig("omitempty") // needed that an empty string "" or 0 will not throw an error.
			}
		}

		switch f.Information.Type.Kind() {
		case "Bool":
			isBelongsTo = true
			f.Validator.appendConfig("eq=false|eq=true")
		case "Integer":
			f.Validator.appendConfig("numeric")
			opt := f.Information.Type.(*types.Int)
			f.Validator.appendConfig(fmt.Sprintf("min=%d", opt.Min))
			f.Validator.appendConfig(fmt.Sprintf("max=%d", opt.Max))
		case "Float":
			f.Validator.appendConfig("numeric")
		case "Text":
			opt := f.Information.Type.(*types.Text)
			f.Validator.appendConfig(fmt.Sprintf("max=%d", opt.Size)) // TODO FIX it must be the byte size
		case "TextArea":
			opt := f.Information.Type.(*types.TextArea)
			f.Validator.appendConfig(fmt.Sprintf("max=%d", opt.Size)) // TODO FIX it must be the byte size
		case "Time":
			//TODO write own
		case "Date":
			//TODO write own
		case "DateTime":
			//TODO write own
		case "Select":
			opt := f.Information.Type.(types.Select)
			f.Validator.appendConfig(fmt.Sprintf("oneof=%s", strings.Join(opt.Items(), " ")))
		}

		// required if null is not allowed and its not an autoincrement column
		if !f.Information.NullAble && !f.Information.Autoincrement && !isBelongsTo {
			f.Validator.appendConfig("required")
		} else {
			// omitempty just makes sense if there is a config defined
			if f.Validator.Config != "" {
				f.Validator.appendConfig("omitempty")
			}
		}
		f.Validator.sort()
	}

	return nil
}

// isValid checks if all system added database fields are valid.
// after that the whole struct gets checked, so that dive and other validations are properly checked.
func (m *Model) isValid() error {
	writePerm := Permission{Write: true}

	// checking all fields (exclusive relations) for system added validation (not per tag)
	for _, f := range m.scope.Fields(writePerm) {
		err := validate.Var(m.scope.CallerField(f.Name).Interface(), f.Validator.Config)
		if err != nil {
			for _, vErr := range err.(valid.ValidationErrors) {
				return fmt.Errorf(errValidation, m.scope.Name(true), f.Name, vErr.ActualTag(), vErr.Param(), vErr.Value())
			}
		}
	}

	// check all the whole struct (incl relations). in that case relations are included and dive validation is working.
	err := validate.Struct(m.caller)
	if err != nil {
		for _, vErr := range err.(valid.ValidationErrors) {
			for _, f := range m.scope.Fields(writePerm) {
				if f.Name == vErr.StructField() {
					return fmt.Errorf(errValidation, m.scope.Name(true), f.Name, vErr.ActualTag(), vErr.Param(), vErr.Value())
				}
			}
		}
	}
	return nil
}

// ValidateValuer is needed for the NullInt,NullString,.. struct.
// It implements the valid.CustomTypeFunc interface.
func ValidateValuer(field reflect.Value) interface{} {
	if valuer, ok := field.Interface().(driverI.Valuer); ok {
		val, err := valuer.Value()
		if err == nil {
			return val
		}
	}
	return nil
}
