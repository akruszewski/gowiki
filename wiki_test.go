package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
)

var testCases = []struct {
	title    string
	document string
	json     []byte
}{
	{
		"test_index",
		"index",
		[]byte(`{"title":"test_index","document":"index","updated":""}`),
	},
	{
		"test_index 2",
		"index2",
		[]byte(`{"title":"test_index 2","document":"index2","updated":""}`),
	},
	{
		"test_index.3",
		"index3",
		[]byte(`{"title":"test_index.3","document":"index3","updated":""}`),
	},
}

func TestPage(t *testing.T) {
	var (
		err    error
		pages  []Page
		tmp    Page
		tmp_js []byte
	)

	// TODO: create tmp folder for tests

	// Test save page method (for page file creation)
	for _, test := range testCases {
		tmp = Page{Title: test.title, Document: test.document}
		err = tmp.save()
		if err != nil {
			t.Errorf("%+v.save() - error occurs: %v", tmp, err)
		}
	}

	// Test loadPage function (for page file read)
	for _, test := range testCases {
		tmp, err := loadPage(test.title)
		if err != nil {
			t.Errorf("loadPage(\"%v\") - error occurs: %v", test.title, err)
		}
		if strings.Compare(tmp.Document, test.document) != 0 {
			t.Errorf(
				"loadPage(%v) - expected document: %s, found %s",
				test.title,
				test.document,
				tmp.Document,
			)
		}
		pages = append(pages, *tmp)
	}

	// Test .toJSON page struct method
	for i, test := range testCases {
		tmp_js, err = pages[i].toJSON()
		if err != nil || bytes.Compare(tmp_js, test.json) != 0 {
			t.Errorf("Page%v(\"%v\") - error occurs: %v", pages[i], test.title, err)
		}
	}

	// Test .fromJSON page struct method
	for i, test := range testCases {
		err = tmp.fromJSON(test.json)
		if err != nil || strings.Compare(tmp.Document, test.document) != 0 {
			t.Errorf("Page%v(\"%v\") - error occurs: %v", pages[i], test.title, err)
		}
	}

	// Test .fromJSON page struct method
	for i, test := range testCases {
		err = tmp.fromJSON(test.json)
		if err != nil || strings.Compare(tmp.Document, test.document) != 0 {
			t.Errorf("Page%v(\"%v\") - error occurs: %v", pages[i], test.title, err)
		}
	}

	// Test removePage function (for page file remove)
	for _, test := range testCases {
		err = removePage(test.title)
		if err != nil {
			t.Errorf(
				"removePage(%v) != nil, error message: %v",
				test.title,
				err,
			)
		}
	}

}

//Test http view handlers.
func TestHandlers(t *testing.T) {
	var (
		ps          = httprouter.Params{httprouter.Param{Key: "title", Value: "test_index"}}
		newDocument = []byte(`{"title":"test_index","document":"asd","updated":""}`)
	)

	//Test PageUpdateHandler http view function.

	//Test creating new page.
	req, err := http.NewRequest(
		"POST",
		"http://localhost:8080/test_index",
		bytes.NewBuffer(testCases[0].json),
	)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	PageUpdateHandler(w, req, ps)
	if w.Code != 200 {
		t.Errorf("Expected 200 HTTP status code, gets %d", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), testCases[0].json) {
		t.Error("Response body is different from expected one!")
	}

	//Test updating existing page.
	req, err = http.NewRequest(
		"POST",
		"http://localhost:8080/test_index",
		bytes.NewBuffer([]byte(`{"document": "asd"}`)),
	)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	PageUpdateHandler(w, req, ps)
	bd := w.Body.Bytes()
	if w.Code != 200 {
		t.Errorf("Expected 200 HTTP status code, gets %d", w.Code)
	}
	if !bytes.Contains(
		bd,
		newDocument,
	) {
		t.Errorf("Response body is different from expected one! %s", bd)
	}

	//Test PageGetHandler http view function.

	// Gets existing page.
	req, err = http.NewRequest(
		"GET",
		"http://localhost:8080/test_index",
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	w = httptest.NewRecorder()
	PageGetHandler(w, req, ps)
	if w.Code != 200 {
		t.Errorf("Expected 200 HTTP status code, gets %d", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), newDocument) {
		t.Error("Response body is different from expected one!")
	}

	// Gets not existing page.
	req, err = http.NewRequest(
		"GET",
		"http://localhost:8080/test_not_existing",
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	w = httptest.NewRecorder()
	PageGetHandler(
		w,
		req,
		httprouter.Params{httprouter.Param{Key: "title", Value: "test_not_existing"}},
	)
	if w.Code != 404 {
		t.Errorf("Expected 404 HTTP status code, gets %d", w.Code)
	}

	//Test PageListHandler
	req, err = http.NewRequest(
		"GET",
		"http://localhost:8080/",
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	w = httptest.NewRecorder()
	PageListHandler(w, req, httprouter.Params{})
	if w.Code != 200 {
		t.Errorf("Expected 200 HTTP status code, gets %d", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`["test_index"]`)) {
		t.Error("Response body is different from expected one!")
	}

	//Test PageDeleteHandler

	// Deletes existing page.
	req, err = http.NewRequest(
		"DELETE",
		"http://localhost:8080/test_index",
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	w = httptest.NewRecorder()
	PageDeleteHandler(w, req, ps)
	if w.Code != 200 {
		t.Errorf("Expected 200 HTTP status code, gets %d", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("{\"message\": \"Page removed\"}")) {
		t.Error("Response body is different from expected one!")
	}
}
