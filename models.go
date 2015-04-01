package sewpulse

import (
	"encoding/json"
	"errors"
	"net/http"
)

func DeleteModel(r *http.Request) (err error) {
	panic("Not implemented... yet")
}

func GetSingleModel(r *http.Request) (err error) {
	return
}

func SaveExistingModel(r *http.Request) (err error) {
	return
}

func CreateNewModel(r *http.Request) error {
	newMod := NewModel()
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&newMod); err != nil {
		return err
	}
	return CreateDecodedNewModel(newMod, r)
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
