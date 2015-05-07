package sewpulse

import (
	"html/template"
	"net/http"
)

type urlStruct struct {
	handler      func(w http.ResponseWriter, r *http.Request)
	templatePath string
}

type apiStruct struct {
	handler func(w http.ResponseWriter, r *http.Request)
}

var urlMaps map[string]urlStruct
var apiMaps map[string]apiStruct
var PRE_BUILT_TEMPLATES = make(map[string]*template.Template)
var PAGE_NOT_FOUND_TEMPLATE = template.Must(template.ParseFiles("templates/pageNotFound.html"))

const API_BOM_ARTICLE_SLASH_END = "/api/bom/article/"
const API_BOM_ARTICLE_END = "/api/bom/article"

const API_BOM_MODEL_SLASH_END = "/api/bom/model/"
const API_BOM_MODEL_END = "/api/bom/model"

const API_RRK_SALE_INVOICE_SALSH_END = "/api/rrk/saleInvoice/"
const API_RRK_SALE_INVOICE_END = "/api/rrk/saleInvoice"
const HTTP_RRK_SALE_INVOICE_SLASH_END = "/rrk/saleInvoice/"

const API_RRK_PURCHASE_INVOICE_SALSH_END = "/api/rrk/purchaseInvoice/"
const API_RRK_PURCHASE_INVOICE_END = "/api/rrk/purchaseInvoice"
const HTTP_RRK_PURCHASE_INVOICE_SLASH_END = "/rrk/purchaseInvoice/"

const API_RRK_RM_INWARD_STK_TRFR_SLASH_END = "/api/rrk/rmInwardStkTrfInvoice/"
const API_RRK_RM_INWARD_STK_TRFR_END = "/api/rrk/rmInwardStkTrfInvoice"

const API_RRK_RM_OUTWARD_STK_TRFR_SLASH_END = "/api/rrk/rmOutwardStkTrfInvoice/"
const API_RRK_RM_OUTWARD_STK_TRFR_END = "/api/rrk/rmOutwardStkTrfInvoice"

const API_RRK_FP_INWARD_STK_TRFR_SLASH_END = "/api/rrk/fpInwardStkTrfInvoice/"
const API_RRK_FP_INWARD_STK_TRFR_END = "/api/rrk/fpInwardStkTrfInvoice"

const API_RRK_FP_OUTWARD_STK_TRFR_SLASH_END = "/api/rrk/fpOutwardStkTrfInvoice/"
const API_RRK_FP_OUTWARD_STK_TRFR_END = "/api/rrk/fpOutwardStkTrfInvoice"

const API_RRK_STOCK_POSITION_FOR_DATE_SLASH_END = "/api/rrk/stock-position-for-date/"

func initDynamicHTMLUrlMaps() {

	http.HandleFunc(HTTP_RRK_SALE_INVOICE_SLASH_END, HTTPSingleSaleInvoiceHandler)
	http.HandleFunc(HTTP_RRK_PURCHASE_INVOICE_SLASH_END, HTTPSinglePurchaseInvoiceHandler)
	http.HandleFunc(API_RRK_STOCK_POSITION_FOR_DATE_SLASH_END, rrkStockPositionForDateSlashApiHandler)
}

func initStaticHTMLUrlMaps() {
	urlMaps := map[string]urlStruct{
		"/":                                             {generalPageHandler, "templates/home.html"},
		"/a":                                            {generalPageHandler, "templates/admin/admin.html"},
		"/a/bom/view":                                   {generalPageHandler, "templates/admin/bom_view.html"},
		"/a/bom/new-model":                              {generalPageHandler, "templates/admin/create_model.html"},
		"/a/bom/new-article":                            {generalPageHandler, "templates/admin/create_article.html"},
		"/a/gzb/view-unsettled-advance":                 {generalPageHandler, "templates/admin/gzb_admin_view_unsettled_advance.html"},
		"/a/rrk/all-sale-invoices":                      {generalPageHandler, "templates/admin/rrk_sale_invoice_all.html"},
		"/a/rrk/all-purchase-invoices":                  {generalPageHandler, "templates/admin/rrk_purchase_invoice_all.html"},
		"/a/rrk/view-unsettled-advance":                 {generalPageHandler, "templates/admin/rrk_admin_view_unsettled_advance.html"},
		"/rrk/daily-polish":                             {generalPageHandler, "templates/rrk_daily_polish.html"},
		"/rrk/daily-assembly":                           {generalPageHandler, "templates/rrk_daily_assembly.html"},
		"/rrk/daily-sale":                               {generalPageHandler, "templates/rrk_daily_sale.html"},
		"/rrk/enter-purchase-invoice":                   {generalPageHandler, "templates/rrk_enter_purchase.html"},
		"/rrk/stock-position":                           {generalPageHandler, "templates/rrk_view_stock_position.html"},
		"/rrk":                                          {generalPageHandler, "templates/rrk.html"},
		"/rrk/daily-cash":                               {generalPageHandler, "templates/rrk_daily_cash.html"},
		"/rrk/raw-material-outward-stock-transfer":      {generalPageHandler, "templates/rrk_rm_outward_stock_transfer.html"},
		"/rrk/raw-material-inward-stock-transfer":       {generalPageHandler, "templates/rrk_rm_inward_stock_transfer.html"},
		"/rrk/finished-products-inward-stock-transfer":  {generalPageHandler, "templates/rrk_fp_inward_stock_transfer.html"},
		"/rrk/finished-products-outward-stock-transfer": {generalPageHandler, "templates/rrk_fp_outward_stock_transfer.html"},
		"/gzb":                    {generalPageHandler, "templates/gzb.html"},
		"/gzb/daily-cash":         {generalPageHandler, "templates/gzb_daily_cash.html"},
		"/gzb/daily-mfg-sale":     {generalPageHandler, "templates/gzb_daily_mfg_sale.html"},
		"/gzb/daily-trading-sale": {generalPageHandler, "templates/gzb_daily_trading_sale.html"},
	}

	for path, urlBlob := range urlMaps {
		templatePath := urlBlob.templatePath
		PRE_BUILT_TEMPLATES[path] = template.Must(template.ParseFiles(templatePath))
	}

	for path, urlBlob := range urlMaps {
		http.HandleFunc(path, urlBlob.handler)
	}
	return
}

func initRootApiMaps() {
	apiMaps := map[string]apiStruct{
		"/api/bom/reset":                           {bomResetAPIHandler},
		"/api/bom/resetToSampleBOM":                {bomResetToSampleState},
		API_BOM_MODEL_END:                          {bomModelWithoutSlashAPIHandler},
		API_BOM_MODEL_SLASH_END:                    {bomModelWithSlashAPIHandler},
		API_BOM_ARTICLE_END:                        {bomArticleWithoutSalshAPIHandler},
		API_BOM_ARTICLE_SLASH_END:                  {bomArticleWithSlashAPIHandler},
		"/api/bom":                                 {bomAPIHandler},
		"/api/gzbCashBookStoreAndEmailApi":         {gzbCashBookStoreAndEmailApiHandler},
		"/api/gzbDailyCashOpeningBalanceApi":       {gzbDailyCashGetOpeningBalanceHandler},
		"/api/gzbDailyCashGetUnsettledAdvancesApi": {gzbDailyCashGetUnsettledAdvancesHandler},
		"/api/gzbDailyCashSettleAccForOneEntryApi": {gzbDailyCashSettleAccForOneEntryApiHandler},
		"/api/gzbDailyMfgSaleEmailSendApi":         {gzbDailyMfgSaleEmailSendApiHandler},
		"/api/gzbDailyTradingSaleEmailSendApi":     {gzbDailyTradingSaleEmailSendApiHandler},
		"/api/gzbGetModelApi":                      {gzbGetModelApiHandler},
		"/gzb/update":                              {gzbDailyCashUpdateModelApiHandler},
		"/api/rrkDailyPolishEmailSendApi":          {rrkDailyPolishEmailSendApiHandler},
		"/api/rrkDailyAssemblySubmissionApi":       {rrkDailyAssemblySubmissionApiHandler},
		API_RRK_SALE_INVOICE_END:                   {rrkSaleInvoiceApiHandler},
		API_RRK_SALE_INVOICE_SALSH_END:             {rrkSaleInvoiceWithSalshApiHandler},
		API_RRK_PURCHASE_INVOICE_END:               {rrkPurchaseInvoiceApiHandler},
		API_RRK_PURCHASE_INVOICE_SALSH_END:         {rrkPurchaseInvoiceWithSalshApiHandler},
		"/api/rrkGetModelApi":                      {rrkGetModelApiHandler},
		"/api/rrkAddModelNameApi":                  {rrkAddModelNameApiHandler},
		"/api/rrkCashBookStoreAndEmailApi":         {rrkCashBookStoreAndEmailApiHandler},
		"/api/rrkDailyCashOpeningBalanceApi":       {rrkDailyCashGetOpeningBalanceHandler},
		"/api/rrkDailyCashGetUnsettledAdvancesApi": {rrkDailyCashGetUnsettledAdvancesHandler},
		"/api/rrkDailyCashSettleAccForOneEntryApi": {rrkDailyCashSettleAccForOneEntryApiHandler},
		"/api/rrk/stock-pristine-date":             {rrkStockPristineDateApiHandler},
		API_RRK_RM_OUTWARD_STK_TRFR_END:            {RRKRMOSTInvoiceNoSalshApiHandler},
		API_RRK_RM_INWARD_STK_TRFR_END:             {RRKRMISTInvoiceNoSalshApiHandler},
		API_RRK_FP_INWARD_STK_TRFR_END:             {RRKFPISTInvoiceNoSalshApiHandler},
		API_RRK_FP_OUTWARD_STK_TRFR_END:            {RRKFPOSTInvoiceNoSalshApiHandler},
		"/rrk/update":                              {rrkDailyCashUpdateModelApiHandler},
		"/api/":                                    {apiNotImplementedHandler},
	}
	for path, apiBlob := range apiMaps {
		http.HandleFunc(path, apiBlob.handler)
	}
	return
}

func init() {
	initRootApiMaps()
	initStaticHTMLUrlMaps()
	initDynamicHTMLUrlMaps()
	return
}

func generalPageHandler(w http.ResponseWriter, r *http.Request) {
	t := PRE_BUILT_TEMPLATES[r.URL.Path]
	if t == nil {
		t = PAGE_NOT_FOUND_TEMPLATE
	}

	if err := t.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
func apiNotImplementedHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, r.URL.Path+" not implemented", http.StatusNotImplemented)

}
