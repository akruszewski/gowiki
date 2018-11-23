package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
)

//Test PageUpdateHandler http view function.
func TestPageUpdateHandler(t *testing.T) {
	_ = setupTestRepo()
	defer tearDownTestRepo()

	//Test creating new page.
	for _, test := range testCases {
		ps := httprouter.Params{httprouter.Param{Key: "title", Value: test.title}}
		req, err := http.NewRequest(
			"POST",
			fmt.Sprintf("http://localhost:8080/%s", test.title),
			bytes.NewBuffer(test.json),
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
		page := Page{}
		err = page.fromJSON(w.Body.Bytes())
		if err != nil {
			t.Error(err)
		}
		if !strings.Contains(page.Document, test.document) {
			t.Error("Response body is different from expected one!")
		}
	}

	//Test updating new page.
	ps := httprouter.Params{httprouter.Param{Key: "title", Value: testCases[0].title}}
	newDocument := []byte(`{"document":"asd"}`)
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("http://localhost:8080/%s", testCases[0].title),
		bytes.NewBuffer(newDocument),
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

	page := Page{}
	err = page.fromJSON(w.Body.Bytes())
	if err != nil {
		t.Errorf("Error occures: %s.", err)
	}

	if !strings.Contains(
		page.Document,
		"asd",
	) {
		t.Errorf("Response body is different from expected one! %s", page)
	}
}

//Test PageGetHandler http view function.
func TestPageGetHandler(t *testing.T) {
	_ = setupTestRepo()

	// Gets existing page.
	for _, test := range testCases {
		ps := httprouter.Params{httprouter.Param{Key: "title", Value: test.title}}
		req, err := http.NewRequest(
			"GET",
			fmt.Sprintf("http://localhost:8080/%s", test.title),
			nil,
		)
		if err != nil {
			log.Fatal(err)
		}
		w := httptest.NewRecorder()
		PageGetHandler(w, req, ps)
		if w.Code != 200 {
			t.Errorf("Expected 200 HTTP status code, gets %d", w.Code)
		}

		page := Page{}
		err = page.fromJSON(w.Body.Bytes())
		if err != nil {
			t.Errorf("Error occures: %s.", err)
		}

		if !strings.Contains(page.Document, test.document) {
			t.Error("Response document body is different from expected one!")
		}
	}

	// Gets not existing page.
	req, err := http.NewRequest(
		"GET",
		"http://localhost:8080/test_not_existing",
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	w := httptest.NewRecorder()
	PageGetHandler(
		w,
		req,
		httprouter.Params{httprouter.Param{Key: "title", Value: "test_not_existing"}},
	)
	if w.Code != 404 {
		t.Errorf("Expected 404 HTTP status code, gets %d", w.Code)
	}
	tearDownTestRepo()
}

//Test PageListHandler http view function.
func TestPageListHandler(t *testing.T) {
	_ = setupTestRepo()
	req, err := http.NewRequest(
		"GET",
		"http://localhost:8080/",
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	w := httptest.NewRecorder()
	PageListHandler(w, req, httprouter.Params{})
	if w.Code != 200 {
		t.Errorf("Expected 200 HTTP status code, gets %d", w.Code)
	}
	bd := w.Body.Bytes()
	if !bytes.Contains(bd, []byte(`["test_index 2","test_index.3","test_index"]`)) {
		t.Errorf("Response body is different (%s) from expected one!", bd)
	}
	tearDownTestRepo()
}

//Test PageDeleteHandler http view function.
func TestPageDeleteHandler(t *testing.T) {
	_ = setupTestRepo()
	// Deletes existing page.
	for _, test := range testCases {
		ps := httprouter.Params{httprouter.Param{Key: "title", Value: test.title}}
		req, err := http.NewRequest(
			"DELETE",
			fmt.Sprintf("http://localhost:8080/%s", test.title),
			nil,
		)
		if err != nil {
			log.Fatal(err)
		}
		w := httptest.NewRecorder()
		PageDeleteHandler(w, req, ps)
		if w.Code != 200 {
			t.Errorf("Expected 200 HTTP status code, gets %d", w.Code)
		}
		if !bytes.Contains(w.Body.Bytes(), []byte("{\"message\": \"Page removed\"}")) {
			t.Error("Response body is different from expected one!")
		}
	}
}
