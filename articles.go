package sewpulse

import (
	"appengine"
	"appengine/datastore"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func DeleteArticle(r *http.Request) error {
	articleName := r.URL.Path[len(API_BOM_ARTICLE_SLASH_END):]
	return DeleteDecodedArticle(articleName, r)
}

func DeleteDecodedArticle(articleName string, r *http.Request) error {
	//If already is not present, return error
	bom, err := GetOrCreateBOMFromDS(r)
	if err != nil {
		return err
	}

	_, present := bom.AML[articleName]
	if !present {
		return errors.New("Article `" + articleName + "` is not present in BOM.")
	}

	//Delete it to article_master_list
	if err := DeleteArticleFromMasterList(r, articleName); err != nil {
		return err
	}

	if err := DeleteArticleFromAllModels(r, articleName); err != nil {
		return err
	}

	//TODO:
	//Bug: Do in a transaction
	return nil
}

func DeleteArticleFromMasterList(r *http.Request, articleName string) error {
	bom, err := GetOrCreateBOMFromDS(r)
	if err != nil {
		return err
	}

	//TODO: Check for pre-exitense of article and throw an error maybe?
	delete(bom.AML, articleName)

	if err := SaveBomInDS(bom, r); err != nil {
		return err
	}

	return nil
}
func DeleteArticleFromAllModels(r *http.Request, articleName string) error {
	//Delete it to from  existing model specs
	bom, err := GetOrCreateBOMFromDS(r)
	if err != nil {
		return err
	}

	for _, model := range bom.Models {
		for an, _ := range model.ArticleAndQty {
			if an == articleName {
				delete(model.ArticleAndQty, articleName)
			}
		}
	}

	if err := SaveBomInDS(bom, r); err != nil {
		return err
	}

	return nil
}

func ExtractArticleFromPostData(r *http.Request) (*Article, error) {
	article := new(Article)
	if err := json.NewDecoder(r.Body).Decode(&article); err != nil {
		return nil, err
	}
	return article, nil
}

func CreateArticle(r *http.Request) error {
	article, err := ExtractArticleFromPostData(r)
	if err != nil {
		return err
	}
	return CreateDecodedNewArticle(article, r)
}

func CreateDecodedNewArticle(article *Article, r *http.Request) error {

	bom, err := GetOrCreateBOMFromDS(r)
	if err != nil {
		return err
	}

	an := article.Name
	_, present := bom.AML[an]
	if present {
		return errors.New("Article `" + an + "` already created.")
	}

	tx := func(c appengine.Context) error {

		//Add it to article_master_list
		bom.AML[article.Name] = *article
		if err := SaveBomInDS(bom, r); err != nil {
			return err
		}

		return AddArticleToExistingModels(r, article)
	}
	c := appengine.NewContext(r)
	if err := datastore.RunInTransaction(c, tx, nil); err == nil {
		return err
	}
	return nil
}

func AddArticleToExistingModels(r *http.Request, article *Article) error {
	//Append it to all existing model specs
	bom, err := GetOrCreateBOMFromDS(r)
	if err != nil {
		return err
	}

	for _, model := range bom.Models {
		model.ArticleAndQty[article.Name] = 0
	}

	if err := SaveBomInDS(bom, r); err != nil {
		return err
	}

	return nil
}

func (aml ArticleMap) String() string {
	if aml == nil {
		return "nil"
	}
	s := "ArticleMap: "
	for _, article := range aml {
		s += fmt.Sprintf("\n %#v", article)
	}
	return s
}

func AddArticleAML(r *http.Request, article *Article) error {
	bom, err := GetOrCreateBOMFromDS(r)
	if err != nil {
		return err
	}

	an := article.Name
	_, present := bom.AML[an]
	if present {
		return errors.New("Article " + an + "already created.")
	}

	bom.AML[an] = *article

	if err := SaveBomInDS(bom, r); err != nil {
		return err
	}
	return nil
}

func PrintArticleMasterListFromDS(r *http.Request) error {
	bom, err := GetOrCreateBOMFromDS(r)
	if err != nil {
		return err
	}
	myDebug(r, fmt.Sprintf("%#v", bom.AML))
	return nil
}
