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
var PRE_BUILT_TEPLATES = make(map[string]*template.Template)
var PAGE_NOT_FOUND_TEMPLATE = template.Must(template.ParseFiles("templates/pageNotFound.html"))

const API_BOM_ARTICLE_SLASH_END = "/api/bom/article/"
const API_BOM_ARTICLE_END = "/api/bom/article"

const API_BOM_MODEL_SLASH_END = "/api/bom/model/"
const API_BOM_MODEL_END = "/api/bom/model"

const API_RRK_SALE_INVOICE_SALSH_END = "/api/rrk/saleInvoice/"
const API_RRK_SALE_INVOICE_END = "/api/rrk/saleInvoice"

func initRootUrlMaps() {
	urlMaps := map[string]urlStruct{
		"/":                             {generalPageHandler, "templates/home.html"},
		"/a":                            {generalPageHandler, "templates/admin/admin.html"},
		"/a/gzb":                        {generalPageHandler, "templates/admin/gzb_admin.html"},
		"/a/bom":                        {generalPageHandler, "templates/admin/bom_admin.html"},
		"/a/bom/view":                   {generalPageHandler, "templates/admin/bom_view.html"},
		"/a/bom/new-model":              {generalPageHandler, "templates/admin/create_model.html"},
		"/a/bom/new-article":            {generalPageHandler, "templates/admin/create_article.html"},
		"/a/gzb/view-unsettled-advance": {generalPageHandler, "templates/admin/gzb_admin_view_unsettled_advance.html"},
		"/a/rrk":                        {generalPageHandler, "templates/admin/rrk_admin.html"},
		"/a/rrk/all-sale-invoices":      {generalPageHandler, "templates/admin/rrk_all_sale_invoices.html"},
		"/a/rrk/view-unsettled-advance": {generalPageHandler, "templates/admin/rrk_admin_view_unsettled_advance.html"},
		"/rrk/daily-polish":             {generalPageHandler, "templates/rrk_daily_polish.html"},
		"/rrk/daily-assembly":           {generalPageHandler, "templates/rrk_daily_assembly.html"},
		"/rrk/daily-sale":               {generalPageHandler, "templates/rrk_daily_sale.html"},
		"/rrk":                          {generalPageHandler, "templates/rrk.html"},
		"/rrk/daily-cash":               {generalPageHandler, "templates/rrk_daily_cash.html"},
		"/gzb":                          {generalPageHandler, "templates/gzb.html"},
		"/gzb/daily-cash":               {generalPageHandler, "templates/gzb_daily_cash.html"},
		"/gzb/daily-mfg-sale":           {generalPageHandler, "templates/gzb_daily_mfg_sale.html"},
		"/gzb/daily-trading-sale":       {generalPageHandler, "templates/gzb_daily_trading_sale.html"},
	}

	for path, urlBlob := range urlMaps {
		templatePath := urlBlob.templatePath
		PRE_BUILT_TEPLATES[path] = template.Must(template.ParseFiles(templatePath))
	}

	for path, urlBlob := range urlMaps {
		http.HandleFunc(path, urlBlob.handler)
	}
	http.HandleFunc("/rrk/saleinvoice/", singleInvoiceHandler)
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
		"/api/rrkDailyAssemblyEmailSendApi":        {rrkDailyAssemblyEmailSendApiHandler},
		"/api/rrk/saleInvoice":                     {rrkSaleInvoiceApiHandler},
		API_RRK_SALE_INVOICE_SALSH_END:             {rrkSaleInvoiceWithSalshApiHandler},
		"/api/rrkGetModelApi":                      {rrkGetModelApiHandler},
		"/api/rrkAddModelNameApi":                  {rrkAddModelNameApiHandler},
		"/api/rrkCashBookStoreAndEmailApi":         {rrkCashBookStoreAndEmailApiHandler},
		"/api/rrkDailyCashOpeningBalanceApi":       {rrkDailyCashGetOpeningBalanceHandler},
		"/api/rrkDailyCashGetUnsettledAdvancesApi": {rrkDailyCashGetUnsettledAdvancesHandler},
		"/api/rrkDailyCashSettleAccForOneEntryApi": {rrkDailyCashSettleAccForOneEntryApiHandler},
		"/rrk/update":                              {rrkDailyCashUpdateModelApiHandler},
	}
	for path, apiBlob := range apiMaps {
		http.HandleFunc(path, apiBlob.handler)
	}
	return
}

func init() {
	initRootApiMaps()
	initRootUrlMaps()
	return
}

func generalPageHandler(w http.ResponseWriter, r *http.Request) {
	t := PRE_BUILT_TEPLATES[r.URL.Path]
	if t == nil {
		t = PAGE_NOT_FOUND_TEMPLATE
	}

	if err := t.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func singleInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/rrk_sale_invoice.html"))
	if err := t.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
