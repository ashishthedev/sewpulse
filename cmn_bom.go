package sewpulse

import (
	"encoding/json"
	"net/http"
)

func bomResetToSampleState(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		if err := ResetBOMToSampleState(r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, r.Method+" Not implemented", http.StatusNotImplemented)
		return
	}
}

func bomResetAPIHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		if err := ResetBOM(r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, r.Method+" Not implemented", http.StatusNotImplemented)
		return
	}
}

func bomModelWithSlashAPIHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		newMod, err := ExtractModelFromRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		newMod.Name = r.URL.Path[len(API_BOM_MODEL_SLASH_END):]
		if err := CreateDecodedNewModel(newMod, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, r.Method+" Not implemented", http.StatusNotImplemented)
		return
	}

}
func bomModelWithoutSlashAPIHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		newMod, err := ExtractModelFromRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := CreateDecodedNewModel(newMod, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, r.Method+" Not implemented", http.StatusNotImplemented)
		return
	}
}

func bomArticleWithoutSalshAPIHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		aml, err := GetorCreateArticleMasterListFromDS(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(*aml)
		return
	case "POST":
		//Create and article with data inside post
		article, err := ExtractArticleFromPostData(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := CreateDecodedNewArticle(article, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, r.Method+" Not implemented", http.StatusNotImplemented)
		return
	}
}

func bomArticleWithSlashAPIHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		an := r.URL.Path[len(API_BOM_ARTICLE_SLASH_END):]
		bom, err := GetOrCreateBOMFromDS(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, article := range bom.AML.Articles {
			if article.Name == an {
				json.NewEncoder(w).Encode(article)
				return
			}
		}
		http.Error(w, "Article "+an+" not present in BOM", http.StatusInternalServerError)
		return

	case "POST":
		article, err := ExtractArticleFromPostData(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		article.Name = r.URL.Path[len(API_BOM_ARTICLE_SLASH_END):]

		if err := CreateDecodedNewArticle(article, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "DELETE":
		if err := DeleteArticle(r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, r.Method+" Not implemented", http.StatusNotImplemented)
		return
	}
}

func bomAPIHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		if err := SaveBOMRcvdInHttpRequest(r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case "GET":
		bom, err := GetOrCreateBOMFromDS(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(*bom)
		return

	default:
		http.Error(w, r.Method+" Not implemented", http.StatusNotImplemented)
		return
	}
}

//======================================================
// Design:
// The whole BOM can be visualized as a table with rows and columns being articles and models. Every individual cell represents the amount of articles used in that particular model.
// Bom is central and contain all the models
// Models contain articles
// Article list has to be same across all models. We accomplish this by carefully monitoring add and delete operations on modesl
// A master article list is maintained which serves as a reference point for any other article based operation.
//
//======================================================

type Article struct {
	Name              string
	Unit              string
	ReorderPoint      int
	MinimumAlarmPoint int
	MaximumAlarmPoint int
	EOQ               int
}

func NewArticle() *Article {
	return new(Article)
}

type ArticleMasterList struct {
	Articles map[string]Article
}

func NewArticleMasterList() *ArticleMasterList {
	aml := new(ArticleMasterList)
	aml.Articles = make(map[string]Article)
	return aml
}

type QtyMap map[string]int

type Model struct {
	//ModelKey  *datastore.Key //If the bom gets too big, store each model as an independent entity in datastore
	Name          string
	Unit          string
	ArticleAndQty QtyMap
}

func NewModel() *Model {
	newModel := new(Model)
	newModel.ArticleAndQty = make(QtyMap)
	return newModel
}

type BOM struct {
	Models map[string]Model
	AML    *ArticleMasterList
}

func NewBOM() *BOM {
	bom := new(BOM)
	bom.Models = make(map[string]Model)
	bom.AML = NewArticleMasterList()
	return bom
}

type ArticleLister interface {
	ArticleList() []string
}

func (x Model) ArticleList() []string {
	var sl []string
	for articleName, _ := range x.ArticleAndQty {
		sl = append(sl, articleName)
	}
	return sl
}

func (x *ArticleMasterList) ArticleList() []string {
	var sl []string
	for an, _ := range x.Articles {
		sl = append(sl, an)
	}
	return sl
}
