package export

import (
	"encoding/csv"
	"fmt"
	"github.com/patrickascher/gofw/controller/context"
	"reflect"
)

func init() {
	_ = context.Register("gridCsv", newCsv)
}

// New satisfies the config.provider interface.
func newCsv() context.Interface {
	return &csvWriter{}
}

type csvWriter struct {
}

func (cw *csvWriter) Name() string {
	return "Csv"
}

func (cw *csvWriter) Icon() string {
	return "mdi-file-delimited"
}

func (cw *csvWriter) Write(r *context.Response) error {

	// TODO define separator
	// TODO define CRLF

	r.Raw().Header().Set("Content-Type", "text/csv")
	r.Raw().Header().Set("Content-Disposition", "attachment; filename=\"export.csv\"")

	w := csv.NewWriter(r.Raw())

	header := r.Data("head").([]string)
	data := r.Data("data")

	//header
	if err := w.Write(header); err != nil {
		return err
	}

	// adding body
	rData := reflect.ValueOf(data)
	for i := 0; i < rData.Len(); i++ {
		var body []string

		for _, head := range header {
			if rData.Index(i).Type().Kind().String() == "struct" {
				body = append(body, fmt.Sprint(rData.Index(i).FieldByName(head).Interface()))
			} else {
				body = append(body, fmt.Sprint(reflect.ValueOf(rData.Index(i).Interface()).MapIndex(reflect.ValueOf(head)).Interface()))
			}
		}

		if err := w.Write(body); err != nil {
			return err
		}
	}

	w.Flush()

	return nil
}
