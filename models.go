package sewpulse

import (
	"encoding/json"
	"errors"
	//"fmt"
	"net/http"
)

func CreateDecodedNewModel(newMod *Model, r *http.Request) error {
	bom, err := GetOrCreateBOMFromDS(r)
	if err != nil {
		return err
	}

	if _, present := bom.Models[newMod.Name]; present {
		return errors.New("Model `" + newMod.Name + "` is already created.")
	}

	newMod.ArticleAndQty = make(QtyMap)

	for an, _ := range bom.AML {
		newMod.ArticleAndQty[an] = 0
	}

	bom.Models[newMod.Name] = *newMod

	return SaveBomInDS(bom, r)
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

func bomModelWithSlashAPIHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		newMod := NewModel()
		if err := HTTPBodyToStruct(r, newMod); err != nil {
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
	case "GET":
		bom, err := GetOrCreateBOMFromDS(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(bom.Models); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	case "POST":
		newMod := NewModel()
		if err := HTTPBodyToStruct(r, newMod); err != nil {
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
