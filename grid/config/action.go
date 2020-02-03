package config

// Action
type Action struct {
	DisplayLeft bool `json:"l,omitempty"`

	New    New    `json:"new,omitempty"`
	Edit   Edit   `json:"edit,omitempty"`
	Delete Delete `json:"delete,omitempty"`
}

type New struct {
	Disable bool `json:"disable,omitempty"`
}

type Edit struct {
	Disable bool `json:"disable,omitempty"`
}

type Delete struct {
	Disable bool `json:"disable,omitempty"`
}
