package sewpulse

import (
	"time"
)

//======================================================
// Design:
// The whole BOM can be visualized as a table with rows and columns being articles and models. Every individual cell represents the amount of articles used in that particular model.
// Bom is central and contain all the models
// Models contain articles
// Article list has to be same across all models. We accomplish this by carefully monitoring add and delete operations on modesl
// A master article list is maintained which serves as a reference point for any other article based operation.
//
//======================================================

//======================================================
// Article
//======================================================
type Article struct {
	Name              string
	Unit              string
	ReorderPoint      int
	MinimumAlarmPoint int
	MaximumAlarmPoint int
	EOQ               int
}

type ArticleMap map[string]Article

//======================================================
// ArticleMasterList
//======================================================
type ArticleMasterList struct {
	Articles ArticleMap
}

type QtyMap map[string]float64

func NewArticleMasterList() *ArticleMasterList {
	aml := new(ArticleMasterList)
	aml.Articles = make(map[string]Article)
	return aml
}

func (x *ArticleMasterList) ArticleList() []string {
	var sl []string
	for an, _ := range x.Articles {
		sl = append(sl, an)
	}
	return sl
}

//======================================================
// Model
//======================================================
type Model struct {
	//ModelKey  *datastore.Key //If the bom gets too big, store each model as an independent entity in datastore
	Name          string
	Unit          string
	ArticleAndQty QtyMap
}

type ModelMap map[string]Model

func (x Model) ArticleList() []string {
	var sl []string
	for articleName, _ := range x.ArticleAndQty {
		sl = append(sl, articleName)
	}
	return sl
}

func NewModel() *Model {
	newModel := new(Model)
	newModel.ArticleAndQty = make(QtyMap)
	return newModel
}

//======================================================
// BOM
//======================================================
type BOM struct {
	Models ModelMap
	AML    *ArticleMasterList
}

func NewBOM() *BOM {
	bom := new(BOM)
	bom.Models = make(map[string]Model)
	bom.AML = NewArticleMasterList()
	return bom
}

//======================================================
// ArticleLister
//======================================================

type ArticleLister interface {
	ArticleList() []string
}

//======================================================
// SoldItem
//======================================================
type SoldItem struct {
	Name             string
	Rate             float64
	Quantity         float64
	ModelWithFullBOM Model
}

//======================================================
// _SaleInvoice
//======================================================

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

//======================================================
// GZBSaleInvoice
//======================================================

type GZBSaleInvoice struct {
	_SaleInvoice
}

//======================================================
// RRKSaleInvoice
//======================================================
type RRKSaleInvoice struct {
	_SaleInvoice
}
