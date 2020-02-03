package orm

import (
	"database/sql/driver"
	"fmt"
	"github.com/patrickascher/gofw/sqlquery"
	"gopkg.in/go-playground/validator.v9"
	"reflect"
	"strings"
)

const (
	ValidatorSeparator = ","
	ValidatorSkip      = "-"
	ValidatorOr        = "|"
)

type Validator struct {
	Config        string
	tagDefinition bool // indicator if the user added some validation
}

func NewValidator(config string) *Validator {

	v := Validator{}

	// adding config and set the indicator that it was added by the user
	if config != "" {
		v.Config = config
		v.tagDefinition = true
	}

	return &v
}

// appendConfig adds the new config if the validation key does not exist yet
func (v *Validator) appendConfig(config string) {

	actualKeys := v.validationKeys(v.Config)
	newKeys := v.validationKeys(config)

	for _, aK := range actualKeys {
		if aK == newKeys[0] {
			return
		}
	}

	if v.Config != "" {
		v.Config += ValidatorSeparator
	}
	v.Config += config
}

// get all defined validation keys
func (*Validator) validationKeys(config string) []string {
	var keys []string

	a := strings.FieldsFunc(config, split)
	for _, k := range a {
		k2 := strings.Split(k, "=")
		keys = append(keys, k2[0])
	}

	return keys
}

// checks if the field validation should get skipped
func (v *Validator) skip() bool {
	return strings.Contains(v.Config, ValidatorSkip)
}

// split the validation by separator and or tag
func split(r rune) bool {
	return r == []rune(ValidatorSeparator)[0] || r == []rune(ValidatorOr)[0]
}

// add the database validation config
func (m *Model) addDBValidation() error {

	for _, col := range m.Table().Cols {

		// check if skip tag is set, the field has no write permission anyway or if its a custom type and does not exist in db.
		if !col.ExistsInDB() || col.Validator.skip() || col.Permission.Write == false {
			continue
		}

		// needed for BelongTo relations, because the field gets updated dynamic in save... valid is called before.
		skipBelongsToRelation := false
		for _, rel := range m.Table().Associations {
			if rel.Type == BelongsTo && rel.StructTable.StructField == col.StructField {
				col.Validator.appendConfig("omitempty") // needed that an empty string "" will not throw an error.
				skipBelongsToRelation = true
			}
		}

		// required if null is not allowed and its not an autoincrement
		req := false
		if !col.Information.NullAble && !col.Information.Autoincrement && !skipBelongsToRelation {
			col.Validator.appendConfig("required")
			req = true
		}

		switch col.Information.Type.Kind() {
		case "Integer":
			if !req {
				col.Validator.appendConfig("omitempty") // needed that an empty string "" will not throw an error.
			}
			col.Validator.appendConfig("numeric")
			opt := col.Information.Type.(*sqlquery.Int)
			col.Validator.appendConfig(fmt.Sprintf("min=%d", opt.Min))
			col.Validator.appendConfig(fmt.Sprintf("max=%d", opt.Max))
		case "Float":
			col.Validator.appendConfig("numeric")
		case "Text":
			opt := col.Information.Type.(*sqlquery.Text)
			col.Validator.appendConfig(fmt.Sprintf("max=%d", opt.Size)) // TODO FIX it must be the byte size
		case "TextArea":
			opt := col.Information.Type.(*sqlquery.TextArea)
			col.Validator.appendConfig(fmt.Sprintf("max=%d", opt.Size)) // TODO FIX it must be the byte size
		case "Time":
			//TODO write my own
		case "Date":
			//TODO write my own
		case "DateTime":
			//TODO write my own
		}
	}
	return nil
}

// TODO Rollback on error
// TODO add db validation if no custom is given to this field
func (m *Model) isValid() error {

	tagDefinition := false

	// check db valid added
	for _, col := range m.Table().Cols {
		// whitelist/blacklist
		if !col.Permission.Write {
			continue
		}

		if col.Validator.tagDefinition {
			tagDefinition = true
		}

		err := validate.Var(reflectField(m.caller, col.StructField).Interface(), col.Validator.Config)
		if err != nil {
			return fmt.Errorf(strings.Replace(err.Error(), "''", "'%s'", -1), structName(m.caller, false)+"."+col.StructField, col.StructField)
		}
	}

	// check user added config
	// all field tags are already checked before, here are the check for dive & co.
	if tagDefinition {
		err := validate.Struct(m.caller)
		if err != nil {

			validationErrors := err.(validator.ValidationErrors)
			for vK, vErr := range validationErrors {
				for _, col := range m.Table().Cols {
					// White/Blacklist
					if col.StructField == vErr.StructField() && !col.Permission.Write {
						validationErrors = append(validationErrors[:vK], validationErrors[vK+1:]...)
					}
				}
			}

			if len(validationErrors) > 0 {
				return err
			}

		}
	}

	return nil
}

// ValidateValuer implements validator.CustomTypeFunc
func ValidateValuer(field reflect.Value) interface{} {

	if valuer, ok := field.Interface().(driver.Valuer); ok {
		val, err := valuer.Value()
		if err == nil {
			return val
		}
	}

	return nil
}
