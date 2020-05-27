// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlquery

// Config stores all information about the database.
type Config struct {
	Name               string `json:"name"`
	Driver             string `json:"driver"`
	Host               string `json:"host"`
	Port               int    `json:"port"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	Database           string `json:"database"`
	Schema             string `json:"schema"`
	MaxOpenConnections int    `json:"maxOpenConnections"`
	MaxIdleConnections int    `json:"maxIdleConnections"`
	MaxConnLifetime    int    `json:"maxConnLifetime"` // in Minutes
	Debug              bool   `json:"debug"`
}
