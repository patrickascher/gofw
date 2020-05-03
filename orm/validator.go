package orm

import (
	driverI "database/sql/driver"
	"fmt"
	"reflect"
	"strings"

	valid "github.com/go-playground/validator"
	"github.com/patrickascher/gofw/sqlquery/types"
)

const (
	tagValidate        = "validate"
	validatorSeparator = ","
	validatorSkip      = "-"
	validatorOr        = "|"
)

var errValidation = "orm: validation failed '%s' for '%s' on the '%s' tag (allowed:%v given:%v)"

// validator struct
type validator struct {
	Config string
}

// appendConfig adds the new config if the validation key does not exist yet.
// if one or more keys are duplicated, they will be skipped.
func (v *validator) appendConfig(config string) {

	actualKeys := v.validationKeys(v.Config)
	newKeys := v.validationKeys(config)

	// list already exists
	var c []string
loop:
	for nk, nv := range newKeys {
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

	// add an trailing separator
	if v.Config != "" && len(c) > 0 {
		v.Config += validatorSeparator
	}

	// set the new config string
	v.Config += strings.Join(c, validatorSeparator)
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

// addDBValidation will add the following validations:
// - belongsTo relations will be set to omitempty (no required tag will be added) (because they are added dynamically after isValid is called on the main orm)
// - columns which does not allow NULL will be required. Except if the column is an autoincrement field, then omitempty is added.
// - Integer: numeric (min,max)
// - Float: numeric
// - Text: size (max)
// - TextArea: size (max)
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
		isBelongsTo := false
		for _, relation := range m.scope.Relations(writePerm) {
			if relation.Kind == BelongsTo && relation.ForeignKey.Name == f.Name {
				f.Validator.appendConfig("omitempty") // needed that an empty string "" will not throw an error.
				isBelongsTo = true
			}
		}

		// required if null is not allowed and its not an autoincrement column
		// skip belongTo fk fields
		if !f.Information.NullAble && !f.Information.Autoincrement && !isBelongsTo {
			if f.Information.Type.Kind() == "Integer" || f.Information.Type.Kind() == "Float" {
				f.Validator.appendConfig("numeric")

			} else {
				f.Validator.appendConfig("required")
			}
		} else {
			f.Validator.appendConfig("omitempty")
		}

		switch f.Information.Type.Kind() {
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
		}

		// TODO create a better way. but omitempty just makes sense if there are other validations added.
		if f.Validator.Config == "omitempty" {
			f.Validator.Config = ""
		}
	}
	return nil
}

// isValid checks if all system added database fields are valid.
// after that the whole struct gets checked against the field tags.
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
