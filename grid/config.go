package grid

// Config - general grid config
type Config struct {
	// ID is used to cache the grid fields. This must be unique.
	ID          string `json:"-"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Policy      int    `json:"-"`
	Action      Action `json:"action,omitempty"`
}

// Action
type Action struct {
	DisplayLeft   bool `json:"l,omitempty"`
	DisableCreate bool `json:"disableCreate,omitempty"`
	DisableDetail bool `json:"disableDetail,omitempty"`
	DisableUpdate bool `json:"disableUpdate,omitempty"`
	DisableDelete bool `json:"disableDelete,omitempty"`
}

func (c *Config) DisableCreate(b bool) *Config {
	c.Action.DisableCreate = b
	return c
}
func (c *Config) DisableDetail(b bool) *Config {
	c.Action.DisableDetail = b
	return c
}
func (c *Config) DisableUpdate(b bool) *Config {
	c.Action.DisableUpdate = b
	return c
}
func (c *Config) DisableDelete(b bool) *Config {
	c.Action.DisableDelete = b
	return c
}
