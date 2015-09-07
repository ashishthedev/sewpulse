package sewpulse

import (
	"appengine"
	"appengine/datastore"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

var GLOBAL_BOM *BOM = nil

func ResetBOMToSampleState(r *http.Request) error {
	if err := ResetBOM(r); err != nil {
		return err
	}
	mod := NewModel()
	for _, x := range []string{"M1", "M2", "M3", "M4", "M5", "M6", "M7", "M8", "M9", "M10", "M11", "M12", "M13", "M14", "M15", "M16", "M17", "M18", "M19", "M20", "M21", "M22", "M23", "M24", "M25", "M26", "M27", "M28", "M29", "M30", "M31", "M32", "M33", "M34", "M35", "M36", "M37", "M38", "M39", "M40", "M41", "M42", "M43", "M44", "M45", "M46", "M47", "M48", "M49", "M50", "M51", "M52", "M53", "M54", "M55", "M56", "M57", "M58", "M59", "M60", "M61", "M62", "M63", "M64", "M65", "M66", "M67", "M68", "M69", "M70", "M71", "M72", "M73", "M74", "M75", "M76", "M77", "M78", "M79", "M80", "M81", "M82", "M83", "M84", "M85", "M86", "M87", "M88", "M89", "M90", "M91", "M92", "M93", "M94", "M95", "M96", "M97", "M98", "M99", "M100"} {
		mod.Name = x
		mod.Unit = "pc"
		if err := CreateDecodedNewModel(mod, r); err != nil {
			return err
		}
	}
	var article Article
	article.Unit = "pc"
	for _, x := range []string{"A1", "A2", "A3", "A4", "A5", "A6", "A7", "A8", "A9", "A10", "A11", "A12", "A13", "A14", "A15", "A16", "A17", "A18", "A19", "A20", "A21", "A22", "A23", "A24", "A25", "A26", "A27", "A28", "A29", "A30", "A31", "A32", "A33", "A34", "A35", "A36", "A37", "A38", "A39", "A40", "A41", "A42", "A43", "A44", "A45", "A46", "A47", "A48", "A49", "A50", "A51", "A52", "A53", "A54", "A55", "A56", "A57", "A58", "A59", "A60", "A61", "A62", "A63", "A64", "A65", "A66", "A67", "A68", "A69", "A70", "A71", "A72", "A73", "A74", "A75", "A76", "A77", "A78", "A79", "A80", "A81", "A82", "A83", "A84", "A85", "A86", "A87", "A88", "A89", "A90", "A91", "A92", "A93", "A94", "A95", "A96", "A97", "A98", "A99", "A100", "A101", "A102", "A103", "A104", "A105", "A106", "A107", "A108", "A109", "A110", "A111", "A112", "A113", "A114", "A115", "A116", "A117", "A118", "A119", "A120", "A121", "A122", "A123", "A124", "A125", "A126", "A127", "A128", "A129", "A130", "A131", "A132", "A133", "A134", "A135", "A136", "A137", "A138", "A139", "A140", "A141", "A142", "A143", "A144", "A145", "A146", "A147", "A148", "A149", "A150", "A151", "A152", "A153", "A154", "A155", "A156", "A157", "A158", "A159", "A160", "A161", "A162", "A163", "A164", "A165", "A166", "A167", "A168", "A169", "A170", "A171", "A172", "A173", "A174", "A175", "A176", "A177", "A178", "A179", "A180", "A181", "A182", "A183", "A184", "A185", "A186", "A187", "A188", "A189", "A190", "A191", "A192", "A193", "A194", "A195", "A196", "A197", "A198", "A199", "A200", "A201", "A202", "A203"} {
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

func SaveBOMRcvdInHttpRequest(r *http.Request) error {
	bom := NewBOM()
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(bom); err != nil {
		return err
	}
	return SaveBomInDS(bom, r)

}

//======================================================
// BOMAsBytes
//======================================================

type BOMAsBytes struct {
	Content []byte `datastore:",noindex"` //GAE DS does not allow us to store slice of slices, therefore we need to convert the BOM as json string and store in the DS
}

func BOMAsBytesKey(r *http.Request) *datastore.Key {
	const BOMAsBytesType = "BOMAsBytes"
	return CMN_SEWNewKey(BOMAsBytesType, "BOMAsBytesKey", 0, r)
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

func (bom *BOM) String() string {
	if bom == nil {
		return "nil"
	}
	s := "Models: "
	for _, model := range bom.Models {
		s += "\n\n" + model.Name
		for an, qty := range model.ArticleAndQty {
			s += fmt.Sprintf("\n%v:%v", an, qty)
		}
	}
	s += "\n\nAML: "

	for _, article := range bom.AML {
		s += "\n" + article.Name
	}
	return s
}
