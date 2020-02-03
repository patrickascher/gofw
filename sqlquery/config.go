package sqlquery

import (
	"fmt"
	"github.com/patrickascher/gofw/logger"
)

// Database interface is provided that you can create your own Adapter which is maybe not supported yet by default.
type Database interface {
	Driver() string
	DSN() string
	Placeholder() *Placeholder
	QuoteCharacter() string
	Debugger() Debugger
	DbName() string
}

// Debugger interface to allow Custom Debuggers
type Debugger interface {
	Debug(msg string, args ...interface{})
}

// Config stores all information about the database.
type Config struct {
	Adapter     string      `json:"adapter"`
	Host        string      `json:"host"`
	Port        int         `json:"port"`
	Username    string      `json:"username"`
	Password    string      `json:"password"`
	Database    string      `json:"database"`
	QuoteChar   string      `json:"quoteCharacter"`
	Debug       bool        `json:"debug"`
	PlaceHolder Placeholder `json:"placeholder"`
}

// Driver must return the *sql.DB driver name which is needed in the *sql.Open method
func (c *Config) Driver() string {
	return c.Adapter
}

// DbName returns the database name of the config
func (c *Config) DbName() string {
	return c.Database
}

// QuoteCharacter returns the quote for identifiers
func (c *Config) QuoteCharacter() string {
	return c.QuoteChar
}

// Placeholder configures and returns a *Placeholder for the given database adapter.
func (c *Config) Placeholder() *Placeholder {
	return &Placeholder{Numeric: c.PlaceHolder.Numeric, Char: c.PlaceHolder.Char}
}

// Debugger returns the a console Debugger
func (c *Config) Debugger() Debugger {
	if c.Debug {
		log, _ := logger.Get(logger.CONSOLE)
		return log
	}
	return nil
}

// DSN returns the DSN configuration which is needed in the *sql.Open method
func (c *Config) DSN() string {
	switch c.Adapter {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true", c.Username, c.Password, c.Host, c.Port, c.Database)
	case "postgres":

		password := "password=%s"
		if c.Password == "" {
			password = "%s"
		}
		database := "dbname=%s"
		if c.Database == "" {
			database = "%s"
		}
		return fmt.Sprintf("host=%s port=%d user=%s "+password+" "+database+" sslmode=disable", c.Host, c.Port, c.Username, c.Password, c.Database)
	}
	return ""
}
