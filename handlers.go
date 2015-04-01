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

const API_ARTICLE = "/api/bom/article/"

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
		"/rrk/daily-sale-old": urlStruct{
			handler:      generalPageHandler,
			templatePath: "templates/rrk_daily_sale_old.html",
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
		"/api/bom/model": apiStruct{
			handler: bomSingleModelAPIHandler,
		},
		"/api/bom/articlesml": apiStruct{
			handler: bomArticlesMasterListAPIHandler,
		},
		//"/api/bom/article/delete": apiStruct{
		//	handler: bomDeleteArticleAPIHandler,
		//},
		API_ARTICLE: apiStruct{
			handler: bomArticleAPIHandler,
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
		"/gzb/update": apiStruct{
			handler: gzbDailyCashUpdateModelApiHandler,
		},
		"/api/rrkDailyPolishEmailSendApi": apiStruct{
			handler: rrkDailyPolishEmailSendApiHandler,
		},
		"/api/rrkDailyAssemblyEmailSendApi": apiStruct{
			handler: rrkDailyAssemblyEmailSendApiHandler,
		},
		"/api/rrkDailySaleEmailSendApiOld": apiStruct{
			handler: rrkDailySaleEmailSendApiHandlerOld,
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
	var t *template.Template = nil
	for k, v := range PRE_BUILT_TEPLATES {
		if r.URL.Path == k {
			t = v
			break
		}
	}

	if t == nil {
		t = PAGE_NOT_FOUND_TEMPLATE
	}

	if err := t.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
