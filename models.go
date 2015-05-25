package sewpulse

import (
	"encoding/json"
	"errors"
	"net/http"
)

func CreateDecodedNewModel(newMod *Model, r *http.Request) error {
	bom, err := GetOrCreateBOMFromDS(r)
	if err != nil {
		return err
	}

	aml, err := GetorCreateArticleMasterListFromDS(r)
	if err != nil {
		return err
	}

	for an, _ := range aml.Articles {
		newMod.ArticleAndQty[an] = 0
	}

	_, present := bom.Models[newMod.Name]
	if present {
		return errors.New("Model `" + newMod.Name + "` is already created.")
	}

	bom.Models[newMod.Name] = *newMod
	if err := SaveBomInDS(bom, r); err != nil {
		return err
	}
	return nil
}

func GetAllModelsFromBOM(r *http.Request) (ModelMap, error) {
	bom, err := GetOrCreateBOMFromDS(r)
	if err != nil {
		return nil, err
	}
	return bom.Models, nil
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
		if err := DecodeBodyToStruct(r, newMod); err != nil {
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
		models, err := GetAllModelsFromBOM(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(models); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	case "POST":
		newMod := NewModel()
		if err := DecodeBodyToStruct(r, newMod); err != nil {
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
