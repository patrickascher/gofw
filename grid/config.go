package grid

// Config - general grid config
type Config struct {
	// ID is used to cache the grid fields. This must be unique.
	ID          string `json:"id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Policy      int    `json:"-"`
	Action      Action `json:"action,omitempty"`

	// should not be able for the user to config.
	Filter  Filter        `json:"filter,omitempty"`
	Exports []ExportTypes `json:"exports,omitempty"`
}

type Filter struct {
	List   []FeGridFilter `json:"list,omitempty"`
	Active FeGridActive   `json:"active,omitempty"`
}

// Action
type Action struct {
	DisplayLeft   bool `json:"l,omitempty"`
	DisableFilter bool `json:"disableFilter,omitempty"`
	DisableCreate bool `json:"disableCreate,omitempty"`
	DisableDetail bool `json:"disableDetail,omitempty"`
	DisableUpdate bool `json:"disableUpdate,omitempty"`
	DisableDelete bool `json:"disableDelete,omitempty"`
}

func (c *Config) DisableFilter(b bool) *Config {
	c.Action.DisableFilter = b
	return c
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
