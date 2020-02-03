package grid

type FieldType interface {
	Name() string
	SetName(name string)

	Options() map[string]interface{}
	SetOption(key string, value interface{})

	Validator() string
	SetValidator(validator string)

	ViewScript() string
	SetViewScript(view interface{})
}

type defaultFieldType struct {
	name      string
	validator string
	view      *value
	options   map[string]interface{}
}

func DefaultFieldType(g *Grid) FieldType {
	return &defaultFieldType{view: valueWithGrid("", g)}
}

func fieldTypeToMap(f FieldType) map[string]interface{} {
	rv := map[string]interface{}{}

	if f.Name() != "" {
		rv["name"] = f.Name()
	}
	if f.Validator() != "" {
		rv["validation"] = f.Validator()
	}
	if len(f.Options()) > 0 {
		rv["options"] = f.Options()
	}
	if f.ViewScript() != "" {
		rv["view"] = f.ViewScript()
	}

	return rv
}

func (f *defaultFieldType) Name() string {
	return f.name
}

func (f *defaultFieldType) SetName(name string) {
	f.name = name
}

func (f *defaultFieldType) Options() map[string]interface{} {
	return f.options
}

func (f *defaultFieldType) SetOption(key string, value interface{}) {
	if f.options == nil {
		f.options = map[string]interface{}{}
	}
	f.options[key] = value
}

func (f *defaultFieldType) Validator() string {
	return f.validator
}

func (f *defaultFieldType) SetValidator(validator string) {
	f.validator = validator
}

func (f *defaultFieldType) ViewScript() string {
	return f.view.getString()
}

func (f *defaultFieldType) SetViewScript(view interface{}) {
	setValueHelper(f.view, view)
}
