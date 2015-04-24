package sewpulse

import (
	"time"
)

type SoldItem struct {
	Name             string
	Rate             float64
	Quantity         float64
	ModelWithFullBOM Model
}

type _SaleInvoice struct {
	Items                []SoldItem
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

type GZBSaleInvoice struct {
	_SaleInvoice
}

type RRKSaleInvoice struct {
	_SaleInvoice
}
