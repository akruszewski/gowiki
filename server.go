package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func RunServer() {
	r := httprouter.New()

	r.GET("/api/wiki_log/", LogHandler)
	r.GET("/api/wiki/:title/log/", PageLogHandler)
	//    r.GET("/api/wiki/:title/:commit_id", PageCommitHandler)

	r.GET("/api/wiki/:title", PageGetHandler)
	r.POST("/api/wiki/:title", PageUpdateHandler)
	r.DELETE("/api/wiki/:title", PageDeleteHandler)

	r.GET("/api/wiki/", PageListHandler)
	//
	//    r.GET("/api/settings/", SettingsGetHandler)
	//    r.POST("/api/settings/", SettingsUpdateHandler)
	//
	//    r.GET("/api/sync_wiki/", SyncWikiHandler)

	log.Println("Starting server on :8080")
	http.ListenAndServe(":8080", r)
}

func LogHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	rep, err := GetWiki(WikiPath)
	if err != nil {
		http.Error(rw, "Wiki isn't configured!", http.StatusInternalServerError)
		return
	}
	lg, err := GetWikiLog(rep)
	if err != nil {
		http.Error(rw, "Can't fetch Wiki log", http.StatusInternalServerError)
	}
	js, err := json.Marshal(lg)
	if err != nil {
		http.Error(rw, "Can't fetch Wiki log", http.StatusInternalServerError)
	}
	rw.Write(js)
}

func PageLogHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	rep, err := GetWiki(WikiPath)
	if err != nil {
		http.Error(rw, "Wiki isn't configured!", http.StatusInternalServerError)
		return
	}
	lg, err := GetFileLog(rep, p.ByName("title")+".wiki")
	if err != nil {
		http.Error(rw, "Can't fetch log for Page.", http.StatusInternalServerError)
	}
	js, err := json.Marshal(lg)
	if err != nil {
		http.Error(rw, "Can't fetch log for Page.", http.StatusInternalServerError)
	}
	rw.Write(js)
}

//
//func PageCommitHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
//}
//

// Http view handler for retrieving wiki page
func PageGetHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	title := p.ByName("title")
	rep, err := GetWiki(WikiPath)
	if err != nil {
		http.Error(rw, "Wiki improperly configured.", http.StatusInternalServerError)
	}
	page, err := loadPage(title, rep)
	if err != nil {
		http.Error(rw, "Page not found.", http.StatusNotFound)
		return
	}
	js, err := page.toJSON()
	if err != nil {
		http.Error(rw, "Decoding error", http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(js)
}

// Http view handler for updating wiki page
func PageUpdateHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	title := p.ByName("title")
	reqBody, err := ioutil.ReadAll(r.Body)
	page := Page{Title: title}

	err = page.fromJSON(reqBody)
	if err != nil {
		http.Error(rw, "Decoding error", http.StatusInternalServerError)
		return
	}

	rep, err := GetWiki(WikiPath)
	if err != nil {
		http.Error(rw, "Wiki improperly configured.", http.StatusInternalServerError)
	}
	page.save()
	page.commit(page.Message, rep)
	page.loadLog(rep)

	js, err := page.toJSON()
	if err != nil {
		http.Error(rw, "Encoding error", http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(js)
}

// Http view handler for deleting wiki page
func PageDeleteHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	rep, err := GetWiki(WikiPath)
	if err != nil {
		http.Error(rw, "Wiki improperly configured.", http.StatusInternalServerError)
	}
	err = removePage(p.ByName("title"), rep)
	if err != nil {
		log.Printf("Error deleting page: %v", err)
		http.Error(rw, "can't delete page", http.StatusBadRequest)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write([]byte("{\"message\": \"Page removed\"}"))
}

// Http view handler for listing wiki pages
func PageListHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	files, err := ioutil.ReadDir(WikiPath)
	var pages []string
	if err != nil {
		log.Printf("Error listing pages: %v", err)
		http.Error(rw, "Can't load page list.", http.StatusInternalServerError)
	}
	// TODO: figure out why this range didn't work
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

//
//func SettingsGetHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
//}
//
//func SettingsUpdateHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
//}
//
//func SyncWikiHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
//}
