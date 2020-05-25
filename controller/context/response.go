// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context

import (
	"encoding/json"
	"github.com/360EntSecGroup-Skylar/excelize"
	"net/http"
	"reflect"
	"time"
)

// Response struct.
type Response struct {
	raw  http.ResponseWriter
	data map[string]interface{}
}

// newResponse initialization the Response struct.
func newResponse(raw http.ResponseWriter) *Response {
	return &Response{raw: raw, data: make(map[string]interface{})}
}

// SetData by key and value.
func (o *Response) SetData(key string, value interface{}) {
	o.data[key] = value
}

// Data returned by the key.
// If the key does not exist, nil will return.
func (o *Response) Data(key string) interface{} {
	if val, ok := o.data[key]; ok {
		return val
	}
	return nil
}

// Raw returns the original *http.ResponseWriter
func (o *Response) Raw() http.ResponseWriter {
	return o.raw
}

// Render the response with the given renderType.
// TODO create a Interface render.Write. So that PDF,Excel and other Exporters can make use of it.
func (o *Response) Render(renderType string) error {
	var err error

	switch renderType {
	case "excel":
		err = o.renderExcel()
	default:
		//TODO: only JSON is defined at the moment
		err = o.renderJson()
	}

	return err
}

// renderJson render the given data to json.
// It sets an content header and marshals the data.
// TODO: this should also be done in the new render package.
func (o *Response) renderJson() error {
	o.Raw().Header().Set("Content-Type", "application/json")
	js, err := json.Marshal(o.data)
	if err != nil {
		return err
	}
	_, err = o.Raw().Write(js)
	return err
}

func (o *Response) renderExcel() error {

	o.Raw().Header().Set("Content-Type", "application/octet-stream")
	o.Raw().Header().Set("Content-Disposition", "attachment; filename=\"export.xlsx\"")

	f := excelize.NewFile()
	worksheet := "Sheet1"
	// Create a new sheet.
	index := f.NewSheet(worksheet)

	header := o.data["head"].([]string)
	data := o.data["data"].([]interface{})

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

	return f.Write(o.Raw())
}
