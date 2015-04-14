package sewpulse

import (
	"time"
)

type PurchaseItem struct {
	Name     string
	Rate     float64
	Quantity int
}

type _PurchaseInvoice struct {
	Items                []PurchaseItem
	Number               string
	DateValue            time.Time
	JSDateValueAsSeconds int64
	GoodsValue           float64
	GrandTotal           float64
	SupplierName         string
	TotalTax             float64
	TotalFreight         float64
	Remarks              string
	UID                  string
	DD_MMM_YY            string
}

type GZBPurchaseInvoice struct {
	_PurchaseInvoice
}

type RRKPurchaseInvoice struct {
	_PurchaseInvoice
}
