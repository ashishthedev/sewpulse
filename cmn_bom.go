package sewpulse

import (
	"encoding/json"
	"net/http"
)

func bomResetToSampleState(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		if err := ResetBOMToSampleState(r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, r.Method+" Not implemented", http.StatusNotImplemented)
		return
	}
}

func bomResetAPIHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		if err := ResetBOM(r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, r.Method+" Not implemented", http.StatusNotImplemented)
		return
	}
}

func bomArticleWithoutSalshAPIHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		bom, err := GetOrCreateBOMFromDS(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(bom.AML)
		return
	case "POST":
		//Create and article with data inside post
		newArticle := new(Article)
		if err := HTTPBodyToStruct(r, newArticle); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := CreateDecodedNewArticle(newArticle, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, r.Method+" Not implemented", http.StatusNotImplemented)
		return
	}
}

func bomArticleWithSlashAPIHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		an := r.URL.Path[len(API_BOM_ARTICLE_SLASH_END):]
		bom, err := GetOrCreateBOMFromDS(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, article := range bom.AML {
			if article.Name == an {
				json.NewEncoder(w).Encode(article)
				return
			}
		}
		http.Error(w, "Article "+an+" not present in BOM", http.StatusInternalServerError)
		return

	case "POST":
		newArticle := new(Article)
		if err := HTTPBodyToStruct(r, newArticle); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		newArticle.Name = r.URL.Path[len(API_BOM_ARTICLE_SLASH_END):]

		if err := CreateDecodedNewArticle(newArticle, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "DELETE":
		if err := DeleteArticle(r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, r.Method+" Not implemented", http.StatusNotImplemented)
		return
	}
}

func bomAPIHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		if err := SaveBOMRcvdInHttpRequest(r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case "GET":
		bom, err := GetOrCreateBOMFromDS(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(*bom)
		return

	default:
		http.Error(w, r.Method+" Not implemented", http.StatusNotImplemented)
		return
	}
}
