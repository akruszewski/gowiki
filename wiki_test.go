package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
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

func tearUp() {
	for _, page := range testCases {
		_ = ioutil.WriteFile(
			WikiPath+page.title+".wiki",
			[]byte(page.document),
			0600,
		)
	}
}

func tearDown() {
	for _, page := range testCases {
		_ = os.Remove(WikiPath + page.title + ".wiki")
	}
}

// Test .save Page type method (for Page file creation)
func TestSavePage(t *testing.T) {
	for _, test := range testCases {
		tmp := Page{Title: test.title, Document: test.document}
		err := tmp.save()
		if err != nil {
			t.Errorf("%+v.save() - error occurs: %v", tmp, err)
		}
	}
	tearDown()
}

// Test loadPage function (for Page file read)
func TestLoadPage(t *testing.T) {
	tearUp()
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
	}
	tearDown()
}

// Test .toJSON Page type method
func TestToJsonPage(t *testing.T) {
	tearUp()
	for _, test := range testCases {
		tmp, _ := loadPage(test.title)
		tmp_js, err := tmp.toJSON()
		if err != nil || bytes.Compare(tmp_js, test.json) != 0 {
			t.Errorf("Page%v(\"%v\") - error occurs: %v", tmp, test.title, err)
		}
	}
	tearDown()
}

// Test .fromJSON Page type method
func TestFromJsonPage(t *testing.T) {
	tearUp()
	for _, test := range testCases {
		tmp, _ := loadPage(test.title)
		err := tmp.fromJSON(test.json)
		if err != nil || strings.Compare(tmp.Document, test.document) != 0 {
			t.Errorf("Page%v(\"%v\") - error occurs: %v", tmp, test.title, err)
		}
	}
	tearDown()
}

// Test removePage function (for page file remove)
func TestRemovePage(t *testing.T) {
	tearUp()
	for _, test := range testCases {
		err := removePage(test.title)
		if err != nil {
			t.Errorf(
				"removePage(%v) != nil, error message: %v",
				test.title,
				err,
			)
		}
	}

}

//Test PageUpdateHandler http view function.
func TestPageUpdateHandler(t *testing.T) {
	tearUp()

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
		if !bytes.Contains(w.Body.Bytes(), test.json) {
			t.Error("Response body is different from expected one!")
		}
	}

	//Test updating new page.
	ps := httprouter.Params{httprouter.Param{Key: "title", Value: testCases[0].title}}
	newDocument := []byte(`{"title":"test_index","document":"asd","updated":""}`)
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
	tearDown()
}

//Test PageGetHandler http view function.
func TestPageGetHandler(t *testing.T) {
	tearUp()

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
		if !bytes.Contains(w.Body.Bytes(), test.json) {
			t.Error("Response body is different from expected one!")
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
	tearDown()
}

//Test PageListHandler http view function.
func TestPageListHandler(t *testing.T) {
	tearUp()
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
	tearDown()
}

//Test PageDeleteHandler http view function.
func TestPageDeleteHandler(t *testing.T) {
	tearUp()
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
