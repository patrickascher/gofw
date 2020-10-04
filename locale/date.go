package locale

import (
	"fmt"
)

type DateFormat struct {
	GoFormat    string
	HumanFormat string
}

func DateFormats() []DateFormat {
	return []DateFormat{{GoFormat: "2006-01-02 15:04:05", HumanFormat: "YYYY-MM-DD H:I:S"}, {GoFormat: "02.01.2006 15:04:05", HumanFormat: "DD.MM.YYYY H:I:S"}}
}

func GoDateFormatFromHumanFormat(f string) (string, error) {
	for _, date := range DateFormats() {
		if date.HumanFormat == f {
			return date.GoFormat, nil
		}
	}
	return "", fmt.Errorf("locale: Date format %s does not exist", f)
}
