package strings

import (
	"github.com/jinzhu/inflection"
	"github.com/serenize/snaker"
)

func CamelToSnake(s string) string {
	return snaker.CamelToSnake(s)
}

func SnakeToCamel(s string) string {
	return snaker.SnakeToCamel(s)
}

func Plural(s string) string {
	return inflection.Plural(s)
}

func Singular(s string) string {
	return inflection.Singular(s)
}
