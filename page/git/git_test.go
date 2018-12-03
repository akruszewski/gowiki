package git

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"testing"

	git "gopkg.in/src-d/go-git.v4"
)

var testCases = []struct {
	title    string
	document string
	json     []byte
}{
	{
		"test_index",
		"index",
		[]byte(`{"title":"test_index","document":"index"}`),
	},
	{
		"test_index 2",
		"index2",
		[]byte(`{"title":"test_index 2","document":"index2"}`),
	},
	{
		"test_index.3",
		"index3",
		[]byte(`{"title":"test_index.3","document":"index3"}`),
	},
}

// Prepares directory and repository for wiki.
func prepareTestWiki() (*git.Repository, string, error) {
	dir, err := ioutil.TempDir("", "tmp_wiki")
	if err != nil {
		log.Fatal(err)
		return nil, "", err
	}
	repo, err := Init(dir)
	if err != nil {
		log.Fatalf("Can't initialize wiki repo %s", err)
		return nil, "", err
	}
	return repo, dir, nil
}

// Prepares test repository and returns its object.
func setupTestRepoWithFiles() (*git.Repository, string) {
	repo, dir, _ := prepareTestWiki()
	for _, pg := range testCases {
		fileName := path.Join(dir, pg.title+".wiki")
		_ = ioutil.WriteFile(
			fileName,
			[]byte(pg.document),
			0600,
		)
		_, err := CommitFile(
			repo,
			fileName,
			fmt.Sprint(fileName, " created."),
			"John Doe",
			"john@doe.com",
		)
		if err != nil {
			log.Fatalf("Can't commit file %s, %s", fileName, err)
		}
	}
	return repo, dir
}

// Removes tmp test directory
func tearDownTestRepo(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		log.Fatalf("Can't remove git repository %s", dir)
	}
}

// Test .save Page type method (for Page file creation)
func TestSave(t *testing.T) {
	repo, dir, _ := prepareTestWiki()
	defer tearDownTestRepo(dir)
	for _, test := range testCases {
		tmp := Page{Title: test.title, Document: test.document}
		err := tmp.Save(dir, repo)
		if err != nil {
			t.Errorf("%+v.save() - error occurs: %v", tmp, err)
		}
	}
}

// Test loadPage function (for Page file read)
func TestLoad(t *testing.T) {
	repo, dir := setupTestRepoWithFiles()
	defer tearDownTestRepo(dir)

	//Try to load page which doesn't exists
	notExistingFile := "ImNotExists"
	_, err := Load(notExistingFile, dir, repo)
	if err == nil {
		t.Errorf("Load(\"%v\") should return err", notExistingFile)
	}

	for _, test := range testCases {
		tmp, err := Load(test.title, dir, repo)
		if err != nil {
			t.Errorf("Load(\"%v\") - error occurs: %v", test.title, err)
		}
		if strings.Compare(tmp.Document, test.document) != 0 {
			t.Errorf(
				"Load(%v) - expected document: %s, found %s",
				test.title,
				test.document,
				tmp.Document,
			)
		}
	}
}

// Test Remove function (for page file remove)
func TestRemove(t *testing.T) {
	repo, dir := setupTestRepoWithFiles()
	defer tearDownTestRepo(dir)
	for _, test := range testCases {
		err := Remove(test.title, dir, repo)
		if err != nil {
			t.Errorf(
				"Remove(%v) != nil, error message: %v",
				test.title,
				err,
			)
		}
	}
}

// Tests FileLog
func TestFileLog(t *testing.T) {
	rep, dir := setupTestRepoWithFiles()
	defer tearDownTestRepo(dir)

	// Get log for existing file.
	ll, err := FileLog(rep, testCases[0].title)
	if err != nil {
		t.Error("Can't get file log: ", err)
	}
	for _, logEntry := range ll {
		if strings.Compare(
			logEntry.Message,
			fmt.Sprintf("%s created.", testCases[0].title),
		) != 0 {
			t.Error("Invalid log message", logEntry.Message)
		}
	}

	// Get log for not existing file.
	ll, err = FileLog(rep, "NotExisitngFile")
	if err != nil {
		t.Error("Can't get file log: ", err)
	}
	if len(ll) > 0 {
		t.Error("Log found for not existing file!")
	}
}

func TestCommitFile(t *testing.T) {
	repo, dir, _ := prepareTestWiki()
	defer tearDownTestRepo(dir)

	fileName := path.Join(dir, testCases[0].title+".wiki")
	message := fmt.Sprint(fileName, " created.")
	err := ioutil.WriteFile(
		fileName,
		[]byte(testCases[0].document),
		0600,
	)
	if err != nil {
		t.Error("Can't create file.")
	}
	logEntry, err := CommitFile(
		repo,
		fileName,
		message,
		"John Doe",
		"john@doe.com",
	)
	if err != nil {
		t.Errorf("Can't commit file %s, %s", fileName, err)
	}
	if strings.Compare(message, logEntry.Message) != 0 {
		t.Errorf(
			"Commit message is different from given one (expect: %s, is: %s)",
			message,
			logEntry.Message,
		)
	}
}

// Tests Init function.
func TestInit(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmp_wiki")
	if err != nil {
		t.Error("Can't create temp dir for wiki, ", err)
	}
	defer os.RemoveAll(dir)

	_, err = Init(dir)
	if err != nil {
		t.Error("Error occurs on git init: ", err)
	}
	_, err = Init(dir)
	if err == nil {
		t.Error("Should get error: Wiki already exists, get <nil> instead.")
	}
}

// Tests Wiki function.
func TestWiki(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmp_wiki")
	if err != nil {
		t.Error("Can't create temp dir for wiki, ", err)
	}

	_, err = Init(dir)
	if err != nil {
		t.Error("Error occurs on git init: ", err)
	}

	_, err = Wiki(dir)
	if err != nil {
		t.Error("Error occurs on retrieving git repository: ", err)
	}
	os.RemoveAll(dir)

	_, err = Wiki(dir)
	if err == nil {
		t.Error("Repository shouldn't be available here.")
	}
}

// Tests Log function.
func TestLog(t *testing.T) {
	rep, dir := setupTestRepoWithFiles()
	defer tearDownTestRepo(dir)
	lg, err := WikiLog(rep)
	if err != nil {
		t.Error("Can't get wiki log: ", err)
	}
	for _, logEntry := range lg {
		if !strings.Contains(logEntry.Message, " created.") {
			t.Error("Invalid log message", logEntry.Message)
		}
	}
}
