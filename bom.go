package sewpulse

import (
	"appengine"
	"appengine/datastore"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var GLOBAL_BOM *BOM = nil

func BOMAsBytesKey(r *http.Request) *datastore.Key {
	const BOMAsBytesType = "BOMAsBytes"
	return CMN_SEWNewKey(BOMAsBytesType, "BOMAsBytesKey", 0, r)
}

type BOMAsBytes struct {
	Content []byte //GAE DS does not allow us to store slice of slices, therefore we need to convert the BOM as json string and store in the DS
}

func ResetBOMToSampleState(r *http.Request) error {
	if err := ResetBOM(r); err != nil {
		return err
	}
	mod := NewModel()
	for _, x := range []string{"PR+", "SURYA", "RUBY"} {
		mod.Name = x
		mod.Unit = "pc"
		if err := CreateDecodedNewModel(mod, r); err != nil {
			return err
		}
	}
	var article Article
	article.Unit = "pc"
	for _, x := range []string{"BODY_PR+", "BODY_SUP", "BODY_SUR", "BBT_280", "Valve_80"} {
		article.Name = x
		article.ReorderPoint = len(x)
		if err := CreateDecodedNewArticle(&article, r); err != nil {
			return err
		}
	}
	return nil
}
func ResetBOM(r *http.Request) error {
	return SaveBomInDS(NewBOM(), r)
}

func GetOrCreateBOMFromDS(r *http.Request) (bom *BOM, err error) {
	if GLOBAL_BOM != nil {
		return GLOBAL_BOM, nil
	}
	c := appengine.NewContext(r)
	k := BOMAsBytesKey(r)
	e := &BOMAsBytes{}

	bom = NewBOM()
	err = datastore.Get(c, k, e)
	if err == datastore.ErrNoSuchEntity {
		err = SaveBomInDS(NewBOM(), r)
		return
	} else if err != nil {
		return
	}

	b := bytes.NewBuffer(e.Content)
	if err = json.NewDecoder(b).Decode(bom); err != nil {
		return
	}

	GLOBAL_BOM = bom
	return
}

func SaveBOMRcvdInHttpRequest(r *http.Request) error {
	bom := NewBOM()
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(bom); err != nil {
		return err
	}
	return SaveBomInDS(bom, r)

}
func SaveBomInDS(bom *BOM, r *http.Request) error {
	bomAsBytes := &BOMAsBytes{Content: []byte{}}
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(*bom); err != nil {
		return err
	}
	bomAsBytes.Content = b.Bytes()
	if _, err := datastore.Put(appengine.NewContext(r), BOMAsBytesKey(r), bomAsBytes); err != nil {
		return err
	}

	GLOBAL_BOM = bom
	return nil
}

func (bom *BOM) String() string {
	if bom == nil {
		return "nil"
	}
	s := "BOM: "
	for _, model := range bom.Models {
		s += "\n" + model.Name + "\n"
		for an, qty := range model.ArticleAndQty {
			s += fmt.Sprintf("  %v:%v", an, qty)
		}
	}
	return s
}
func GetModelWithName(r *http.Request, modelName string) (Model, error) {
	bom, err := GetOrCreateBOMFromDS(r)
	if err != nil {
		return Model{}, err
	}
	for _, model := range bom.Models {
		if model.Name == modelName {
			return model, nil
		}
	}
	return Model{}, errors.New("No model exists with name: " + modelName)

}
