package export

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/patrickascher/gofw/controller/context"
	"reflect"
	"time"
)

func init() {
	_ = context.Register("gridExcel", newExcel)
}

// New satisfies the config.provider interface.
func newExcel() context.Interface {
	return &excelWriter{}
}

type excelWriter struct {
}

func (ew *excelWriter) Name() string {
	return "Excel"
}

func (ew *excelWriter) Icon() string {
	return "mdi-microsoft-excel"
}

func (ew *excelWriter) Write(r *context.Response) error {

	r.Raw().Header().Set("Content-Type", "application/octet-stream")
	r.Raw().Header().Set("Content-Disposition", "attachment; filename=\"export.xlsx\"")

	f := excelize.NewFile()
	worksheet := "Sheet1"
	// Create a new sheet.
	index := f.NewSheet(worksheet)

	header := r.Data("head").([]string)
	data := r.Data("data")

	// adding header data
	i := 1
	fmt.Println(header)
	for _, head := range header {
		cell, err := excelize.CoordinatesToCellName(i, 1)
		if err != nil {
			return err
		}
		err = f.SetCellValue(worksheet, cell, head)
		if err != nil {
			return err
		}
		i++
	}

	// adding body
	rData := reflect.ValueOf(data)
	line := 2
	for i := 0; i < rData.Len(); i++ {
		n := 1
		for _, head := range header {
			cell, err := excelize.CoordinatesToCellName(n, line)
			if err != nil {
				return err
			}

			var value interface{}
			if rData.Index(i).Type().Kind().String() == "struct" {
				value = rData.Index(i).FieldByName(head).Interface()
			} else {
				value = reflect.ValueOf(rData.Index(i).Interface()).MapIndex(reflect.ValueOf(head)).Interface()
			}

			// excel only allows UTC times.
			typ := reflect.TypeOf(value)
			if typ != nil && typ.String() == "time.Time" {
				value = value.(time.Time).String()
			}

			err = f.SetCellValue(worksheet, cell, value)
			if err != nil {
				return err
			}
			n++
		}
		line++
	}

	// Set active sheet of the workbook.
	f.SetActiveSheet(index)

	err := f.Write(r.Raw())

	f = nil
	header = make([]string, 0)
	data = make([]interface{}, 0)

	return err
}
