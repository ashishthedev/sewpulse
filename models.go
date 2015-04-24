package sewpulse

import (
	"encoding/json"
	"errors"
	"net/http"
)

func ExtractModelFromRequest(r *http.Request) (*Model, error) {
	newMod := NewModel()
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&newMod); err != nil {
		return nil, err
	}
	return newMod, nil
}

func CreateDecodedNewModel(newMod *Model, r *http.Request) error {
	bom, err := GetOrCreateBOMFromDS(r)
	if err != nil {
		return err
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
