package sewpulse

import (
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
	aml, err := GetorCreateArticleMasterListFromDS(r)
	if err != nil {
		return err
	}

	_, present := aml.Articles[articleName]
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
	aml, err := GetorCreateArticleMasterListFromDS(r)
	if err != nil {
		return err
	}

	for an, _ := range aml.Articles {
		if articleName == an {
			delete(aml.Articles, articleName)
		}
	}

	if err := SaveArticleMasterList(aml, r); err != nil {
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
	article := NewArticle()
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

	aml, err := GetorCreateArticleMasterListFromDS(r)
	if err != nil {
		return err
	}

	an := article.Name
	_, present := aml.Articles[an]
	if present {
		return errors.New("Article `" + an + "` already created.")
	}

	//Add it to article_master_list
	if err := AddArticleToMasterList(r, article); err != nil {
		return err
	}

	if err := AddArticleToExistingModels(r, article); err != nil {
		return err
	}
	//TODO:
	//Bug: Do in a transaction
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

func (aml *ArticleMasterList) String() string {
	if aml == nil {
		return "nil"
	}
	s := "ArticleMasterList: "
	for _, article := range aml.Articles {
		s += fmt.Sprintf("\n %#v", article)
	}
	return s
}

func AddArticleToMasterList(r *http.Request, article *Article) error {
	aml, err := GetorCreateArticleMasterListFromDS(r)
	if err != nil {
		return err
	}

	an := article.Name
	_, present := aml.Articles[an]
	if present {
		return errors.New("Article " + an + "already created.")
	}

	aml.Articles[an] = *article

	if err := SaveArticleMasterList(aml, r); err != nil {
		return err
	}
	return nil
}

func GetorCreateArticleMasterListFromDS(r *http.Request) (*ArticleMasterList, error) {
	bom, err := GetOrCreateBOMFromDS(r)
	if err != nil {
		return nil, err
	}
	return bom.AML, nil
}

func SaveArticleMasterList(masterList *ArticleMasterList, r *http.Request) error {
	bom, err := GetOrCreateBOMFromDS(r)
	if err != nil {
		return err
	}

	bom.AML = masterList
	if err := SaveBomInDS(bom, r); err != nil {
		return err
	}
	return nil
}

func PrintArticleMasterListFromDS(r *http.Request) error {
	aml, err := GetorCreateArticleMasterListFromDS(r)
	if err != nil {
		return err
	}
	myDebug(r, fmt.Sprintf("%#v", aml))
	return nil
}
