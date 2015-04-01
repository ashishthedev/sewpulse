package sewpulse

import (
	"encoding/json"
	"net/http"
)

func bomResetToSampleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, r.Method+" Not implemented", http.StatusNotImplemented)
		return
	}
	if err := ResetBOMToSampleState(r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func bomResetAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, r.Method+" Not implemented", http.StatusNotImplemented)
		return
	}
	if err := ResetBOM(r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func bomSingleModelAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := CreateNewModel(r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if r.Method == "DELETE" {
		if err := DeleteModel(r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if r.Method == "GET" {
		if err := GetSingleModel(r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if r.Method == "PUT" {
		if err := SaveExistingModel(r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func bomArticlesMasterListAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		aml, err := GetorCreateArticleMasterListFromDS(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(*aml)
	}
	return
}

func bomArticleAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := CreateArticle(r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if r.Method == "DELETE" {
		if err := DeleteArticle(r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func bomAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := SaveBOMRcvdInHttpRequest(r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if r.Method == "GET" {
		bom, err := GetOrCreateBOMFromDS(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(*bom)
	}

	return
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
