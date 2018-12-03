package webservice

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	page "github.com/akruszewski/awiki/page/git"
	"github.com/akruszewski/awiki/settings"
	"github.com/akruszewski/awiki/webservice/auth"
	"github.com/akruszewski/awiki/webservice/auth/jwt"
	"github.com/julienschmidt/httprouter"
)

func RunServer() {
	r := httprouter.New()

	r.GET("/api/wiki_log/", wikiLog)
	r.GET("/api/wiki/:title/log/", pageLog)

	r.GET("/api/wiki/:title", pageGet)
	r.POST("/api/wiki/:title", pageUpdate)
	r.DELETE("/api/wiki/:title", pageDelete)

	r.GET("/api/wiki/", pageList)

	r.GET("/api/auth/login/", login)
	r.GET("/api/auth/logout/", logout)
	r.GET("/api/auth/refresh-token", refreshToken)

	log.Printf("Starting server on :%v", settings.Port)
	http.ListenAndServe(fmt.Sprintf(":%v", settings.Port), r)
}

// Http view handler for retrieving wiki log
func wikiLog(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	rep, err := page.Wiki(settings.WikiPath)
	if err != nil {
		http.Error(rw, "Wiki isn't configured!", http.StatusInternalServerError)
		return
	}
	lg, err := page.WikiLog(rep)
	if err != nil {
		http.Error(rw, "Can't fetch Wiki log", http.StatusInternalServerError)
		return
	}
	js, err := json.Marshal(lg)
	if err != nil {
		http.Error(rw, "Can't fetch Wiki log", http.StatusInternalServerError)
		return
	}
	rw.Write(js)
}

// Http view handler for retrieving page log.
func pageLog(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	rep, err := page.Wiki(settings.WikiPath)
	if err != nil {
		http.Error(rw, "Wiki isn't configured!", http.StatusInternalServerError)
		return
	}
	lg, err := page.FileLog(rep, p.ByName("title")+".wiki")
	if err != nil {
		http.Error(rw, "Can't fetch log for Page.", http.StatusInternalServerError)
		return
	}
	js, err := json.Marshal(lg)
	if err != nil {
		http.Error(rw, "Can't fetch log for Page.", http.StatusInternalServerError)
		return
	}
	rw.Write(js)
}

// Http view handler for retrieving wiki page
func pageGet(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	title := p.ByName("title")
	rep, err := page.Wiki(settings.WikiPath)
	if err != nil {
		http.Error(rw, "Wiki improperly configured.", http.StatusInternalServerError)
		return
	}
	page, err := page.Load(title, settings.WikiPath, rep)
	if err != nil {
		http.Error(rw, "Page not found.", http.StatusNotFound)
		return
	}
	js, err := json.Marshal(page)
	if err != nil {
		http.Error(rw, "Decoding error", http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(js)
}

// Http view handler for updating wiki page
func pageUpdate(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	title := p.ByName("title")
	reqBody, err := ioutil.ReadAll(r.Body)
	pg := page.Page{Title: title}

	err = json.Unmarshal(reqBody, &pg)
	if err != nil {
		http.Error(rw, "Decoding error", http.StatusInternalServerError)
		return
	}

	rep, err := page.Wiki(settings.WikiPath)
	if err != nil {
		http.Error(rw, "Wiki improperly configured.", http.StatusInternalServerError)
		return
	}
	pg.Save(settings.WikiPath, rep)
	pg.LoadLog(rep, settings.WikiPath)

	js, err := json.Marshal(pg)
	if err != nil {
		http.Error(rw, "Encoding error", http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(js)
}

// Http view handler for deleting wiki page
func pageDelete(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	rep, err := page.Wiki(settings.WikiPath)
	if err != nil {
		http.Error(rw, "Wiki improperly configured.", http.StatusInternalServerError)
		return
	}
	err = page.Remove(p.ByName("title"), settings.WikiPath, rep)
	if err != nil {
		log.Printf("Error deleting page: %v", err)
		http.Error(rw, "can't delete page", http.StatusBadRequest)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write([]byte("{\"message\": \"Page removed\"}"))
}

// Http view handler for listing wiki pages
func pageList(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	files, err := ioutil.ReadDir(settings.WikiPath)
	var pages []string
	if err != nil {
		log.Printf("Error listing pages: %v", err)
		http.Error(rw, "Can't load page list.", http.StatusInternalServerError)
		return
	}
	for _, file := range files {
		fileName := file.Name()
		if strings.HasSuffix(fileName, ".wiki") {
			pages = append(pages, strings.TrimSuffix(fileName, ".wiki"))
		}
	}
	js, err := json.Marshal(pages)
	if err != nil {
		log.Printf("Error marshaling pages: %v", err)
		http.Error(rw, "Can't load page list.", http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(js)
}

func login(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	requestUser := new(auth.User)
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&requestUser)

	responseStatus, token := jwt.Login(requestUser)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseStatus)
	w.Write(token)
}

func refreshToken(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	requestUser := new(auth.User)
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&requestUser)

	w.Header().Set("Content-Type", "application/json")
	w.Write(jwt.RefreshToken(requestUser))
}

func logout(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := jwt.Logout(r)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
