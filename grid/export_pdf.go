package grid

import (
	"fmt"
	"github.com/jung-kurt/gofpdf"
	"github.com/patrickascher/gofw/controller/context"
	"reflect"
	"time"
)

func init() {
	_ = context.Register("gridPdf", newPdf)
}

// New satisfies the config.provider interface.
func newPdf() context.Interface {
	return &pdfWriter{}
}

type pdfWriter struct {
}

func (pw *pdfWriter) Name() string {
	return "PDF"
}

func (pw *pdfWriter) Icon() string {
	return "mdi-file-pdf-outline"
}

func (pw *pdfWriter) Write(r *context.Response) error {

	// TODO config image?, title, desc, image
	// TODO improve table span

	r.Raw().Header().Set("Content-Type", "application/pdf")
	r.Raw().Header().Set("Content-Disposition", "attachment; filename=\"export.pdf\"")

	var header []Field
	for _, h := range r.Data("head").([]Field) {
		if h.IsRemoved() {
			continue
		}
		header = append(header, h)
	}
	data := r.Data("data")

	// pdf general options
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Times", "B", 28)
	pdf.Cell(40, 10, "Export")
	pdf.Ln(12)
	pdf.SetFont("Times", "", 20)
	pdf.Cell(40, 10, time.Now().Format("Mon Jan 2, 2006"))
	pdf.Ln(20)

	// calculate size
	cellWidth := float64(264 / len(header))
	cellHeight := float64(5)
	fontSize := float64(6)

	// header
	pdf.SetFont("Times", "B", fontSize)
	pdf.SetFillColor(240, 240, 240)
	for _, head := range header {
		pdf.CellFormat(cellWidth, cellHeight, head.Title(), "1", 0, "", true, 0, "")
	}
	pdf.Ln(-1)

	// body
	pdf.SetFont("Times", "", fontSize)
	pdf.SetFillColor(255, 255, 255)
	rData := reflect.ValueOf(data)
	for i := 0; i < rData.Len(); i++ {
		for _, head := range header {
			var body string
			if rData.Index(i).Type().Kind().String() == "struct" {
				body = fmt.Sprint(rData.Index(i).FieldByName(head.id).Interface())
			} else {
				body = fmt.Sprint(reflect.ValueOf(rData.Index(i).Interface()).MapIndex(reflect.ValueOf(head.id)).Interface())
			}
			pdf.CellFormat(cellWidth, cellHeight, body, "1", 0, "", true, 0, "")
		}
		pdf.Ln(-1)
	}

	// image
	//pdf.ImageOptions("stats.png", 225, 10, 25, 25, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")

	return pdf.Output(r.Raw())
}
