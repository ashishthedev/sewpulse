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

func initRootUrlMaps() {
	urlMaps := map[string]urlStruct{
		"/": urlStruct{
			handler:      generalPageHandler,
			templatePath: "templates/home.html",
		},
		"/a": urlStruct{
			handler:      generalPageHandler,
			templatePath: "templates/admin/admin.html",
		},
		"/a/gzb": urlStruct{
			handler:      generalPageHandler,
			templatePath: "templates/admin/gzb_admin.html",
		},
		"/a/bom": urlStruct{
			handler:      generalPageHandler,
			templatePath: "templates/admin/bom_admin.html",
		},
		"/a/bom/view": urlStruct{
			handler:      generalPageHandler,
			templatePath: "templates/admin/bom_view.html",
		},
		"/a/bom/new-model": urlStruct{
			handler:      generalPageHandler,
			templatePath: "templates/admin/create_model.html",
		},
		"/a/bom/new-article": urlStruct{
			handler:      generalPageHandler,
			templatePath: "templates/admin/create_article.html",
		},
		"/a/gzb/view-unsettled-advance": urlStruct{
			handler:      generalPageHandler,
			templatePath: "templates/admin/gzb_admin_view_unsettled_advance.html",
		},
		"/a/rrk": urlStruct{
			handler:      generalPageHandler,
			templatePath: "templates/admin/rrk_admin.html",
		},
		"/a/rrk/view-unsettled-advance": urlStruct{
			handler:      generalPageHandler,
			templatePath: "templates/admin/rrk_admin_view_unsettled_advance.html",
		},
		"/gzb": urlStruct{
			handler:      generalPageHandler,
			templatePath: "templates/gzb.html",
		},
		"/gzb/daily-cash": urlStruct{
			handler:      generalPageHandler,
			templatePath: "templates/gzb_daily_cash.html",
		},
		"/gzb/daily-sale": urlStruct{
			handler:      generalPageHandler,
			templatePath: "templates/gzb_daily_sale.html",
		},
		"/rrk/daily-polish": urlStruct{
			handler:      generalPageHandler,
			templatePath: "templates/rrk_daily_polish.html",
		},
		"/rrk/daily-assembly": urlStruct{
			handler:      generalPageHandler,
			templatePath: "templates/rrk_daily_assembly.html",
		},
		"/rrk/daily-sale": urlStruct{
			handler:      generalPageHandler,
			templatePath: "templates/rrk_daily_sale.html",
		},
		"/rrk": urlStruct{
			handler:      generalPageHandler,
			templatePath: "templates/rrk.html",
		},
		"/rrk/daily-cash": urlStruct{
			handler:      generalPageHandler,
			templatePath: "templates/rrk_daily_cash.html",
		},
	}

	for path, urlBlob := range urlMaps {
		templatePath := urlBlob.templatePath
		PRE_BUILT_TEPLATES[path] = template.Must(template.ParseFiles(templatePath))
	}

	for path, urlBlob := range urlMaps {
		http.HandleFunc(path, urlBlob.handler)
	}
	return
}

func initRootApiMaps() {
	apiMaps := map[string]apiStruct{
		"/api/bom/reset": apiStruct{
			handler: bomResetAPIHandler,
		},
		"/api/bom/resetToSampleBOM": apiStruct{
			handler: bomResetToSampleState,
		},
		API_BOM_MODEL_END: apiStruct{
			handler: bomModelWithoutSlashAPIHandler,
		},
		API_BOM_MODEL_SLASH_END: apiStruct{
			handler: bomModelWithSlashAPIHandler,
		},
		API_BOM_ARTICLE_END: apiStruct{
			handler: bomArticleWithoutSalshAPIHandler,
		},
		API_BOM_ARTICLE_SLASH_END: apiStruct{
			handler: bomArticleWithSlashAPIHandler,
		},
		"/api/bom": apiStruct{
			handler: bomAPIHandler,
		},
		"/api/gzbCashBookStoreAndEmailApi": apiStruct{
			handler: gzbCashBookStoreAndEmailApiHandler,
		},
		"/api/gzbDailyCashOpeningBalanceApi": apiStruct{
			handler: gzbDailyCashGetOpeningBalanceHandler,
		},
		"/api/gzbDailyCashGetUnsettledAdvancesApi": apiStruct{
			handler: gzbDailyCashGetUnsettledAdvancesHandler,
		},
		"/api/gzbDailyCashSettleAccForOneEntryApi": apiStruct{
			handler: gzbDailyCashSettleAccForOneEntryApiHandler,
		},
		"/api/gzbDailySaleEmailSendApi": apiStruct{
			handler: gzbDailySaleEmailSendApiHandler,
		},
		"/api/gzbGetModelApi": apiStruct{
			handler: gzbGetModelApiHandler,
		},
		"/gzb/update": apiStruct{
			handler: gzbDailyCashUpdateModelApiHandler,
		},
		"/api/rrkDailyPolishEmailSendApi": apiStruct{
			handler: rrkDailyPolishEmailSendApiHandler,
		},
		"/api/rrkDailyAssemblyEmailSendApi": apiStruct{
			handler: rrkDailyAssemblyEmailSendApiHandler,
		},
		"/api/rrkDailySaleEmailSendApi": apiStruct{
			handler: rrkDailySaleEmailSendApiHandler,
		},
		"/api/rrkGetModelApi": apiStruct{
			handler: rrkGetModelApiHandler,
		},
		"/api/rrkAddModelNameApi": apiStruct{
			handler: rrkAddModelNameApiHandler,
		},
		"/api/rrkCashBookStoreAndEmailApi": apiStruct{
			handler: rrkCashBookStoreAndEmailApiHandler,
		},
		"/api/rrkDailyCashOpeningBalanceApi": apiStruct{
			handler: rrkDailyCashGetOpeningBalanceHandler,
		},
		"/api/rrkDailyCashGetUnsettledAdvancesApi": apiStruct{
			handler: rrkDailyCashGetUnsettledAdvancesHandler,
		},
		"/api/rrkDailyCashSettleAccForOneEntryApi": apiStruct{
			handler: rrkDailyCashSettleAccForOneEntryApiHandler,
		},
		"/rrk/update": apiStruct{
			handler: rrkDailyCashUpdateModelApiHandler,
		},
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
