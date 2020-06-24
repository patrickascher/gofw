package grid

import (
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
	data := r.Data("data").([]interface{})

	// adding header data
	i := 1
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
	i = 2
	for _, body := range data {
		n := 1

		bodyx := body.(map[string]interface{})
		for _, head := range header {
			cell, err := excelize.CoordinatesToCellName(n, i)
			if err != nil {
				return err
			}

			// excel only allows UTC times.
			typ := reflect.TypeOf(bodyx[head])
			if typ != nil && typ.String() == "time.Time" {
				bodyx[head] = bodyx[head].(time.Time).String()
			}

			err = f.SetCellValue(worksheet, cell, bodyx[head])
			if err != nil {
				return err
			}
			n++
		}
		i++
	}

	// Set active sheet of the workbook.
	f.SetActiveSheet(index)

	err := f.Write(r.Raw())

	f = nil
	header = make([]string, 0)
	data = make([]interface{}, 0)

	return err
}
