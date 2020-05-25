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

func ExistInt(slice []int, search int) (int, bool) {
	for i, s := range slice {
		if s == search {
			return i, true
		}
	}
	return 0, false
}

func ExistInterface(slice []interface{}, search interface{}) (int, bool) {
	for i, s := range slice {
		if s == search {
			return i, true
		}
	}
	return 0, false
}

func Reverse(numbers []string) []string {
	newNumbers := make([]string, len(numbers))
	for i, j := 0, len(numbers)-1; i <= j; i, j = i+1, j-1 {
		newNumbers[i], newNumbers[j] = numbers[j], numbers[i]
	}
	return newNumbers
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
