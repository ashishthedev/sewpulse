package sewpulse

import (
	"time"
)

type InvoiceItem struct {
	Name      string
	Rate      float64
	Quantity  int
	ModelVal  Model `datastore:"-"`
	ModelJSON []byte
}

type SaleInvoice struct {
	Items                []InvoiceItem
	Number               string
	DateValue            time.Time
	JSDateValueAsSeconds int64
	GoodsValue           float64
	GrandTotal           float64
	CustomerName         string
	TotalTax             float64
	TotalFreight         float64
	Remarks              string
	UID                  string
	DD_MMM_YY            string
}
