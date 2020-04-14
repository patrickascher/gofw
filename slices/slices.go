package slices

import "strings"

func Exists(slice []string, search string) (int, bool) {
	for i, s := range slice {
		if s == search {
			return i, true
		}
	}
	return 0, false
}

func PrefixedWith(slice []string, search string) []string {
	var rv []string
	for _, s := range slice {
		if strings.HasPrefix(s, search) {
			rv = append(rv, s)
		}
	}
	return rv
}

func MergeUnique(slice []string, slice2 []string) []string {
	s := append(slice, slice2...)
	return Unique(s)
}

func Unique(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
