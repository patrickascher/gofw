package grid

// Config - general grid config
type Config struct {
	Action Action `json:"action,omitempty"`
}

// Action
type Action struct {
	DisplayLeft bool `json:"l,omitempty"`

	DisableCreate  bool `json:"disableCreate,omitempty"`
	DisableDetails bool `json:"DisableDetails,omitempty"`
	DisableUpdate  bool `json:"DisableUpdate,omitempty"`
	DisableDelete  bool `json:"DisableDelete,omitempty"`
}

func (c *Config) DisableCreate(b bool) *Config {
	c.Action.DisableCreate = b
	return c
}
