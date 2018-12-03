package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/akruszewski/awiki/page"
	"github.com/pkg/errors"
	git "gopkg.in/src-d/go-git.v4"
	gitObject "gopkg.in/src-d/go-git.v4/plumbing/object"
)

type Page page.Page
type LogEntry page.LogEntry
type Log page.Log

// Saves Document file in WikiPath with Page Title and '.wiki' suffix
func (p *Page) Save(wikiPath string, repo *git.Repository) error {
	err := ioutil.WriteFile(
		path.Join(wikiPath, p.Title+".wiki"),
		[]byte(p.Document),
		0600,
	)
	if err != nil {
		return errors.Wrap(err, "Can't save wiki page")
	}
	logEntry, err := CommitFile(
		repo,
		p.Title,
		fmt.Sprintf("Page %s saved.", p.Title),
		os.Getenv("GIT_USERNAME"),
		os.Getenv("GIT_EMAIL"),
	)
	if err != nil {
		return errors.Wrap(err, "Can't commit wiki page")
	}
	p.Log = append(p.Log, *logEntry)
	return nil
}

// Load Page file commits log
func (p *Page) LoadLog(r *git.Repository, wikiPath string) error {
	lg, err := FileLog(r, p.Title+".wiki")
	if err != nil {
		return errors.Wrap(err, "Can't load wiki page log")
	}
	p.Log = append(p.Log, lg...)
	return nil
}

// Returns Page type variable loaded from WikiPath with given title.
func Load(title string, wikiPath string, r *git.Repository) (*Page, error) {
	filePath := path.Join(wikiPath, title+".wiki")
	body, err := ioutil.ReadFile(filePath)
	page := Page{Title: title, Document: string(body)}
	page.LoadLog(r, wikiPath)
	if err != nil {
		return nil, errors.Wrap(err, "Can't load wiki page")
	}
	return &page, nil
}

// Removes Page Document file with given title.
func Remove(title string, wikiPath string, r *git.Repository) error {
	err := os.Remove(path.Join(wikiPath, title+".wiki"))
	if err != nil {
		return errors.Wrap(err, "Can't remove wiki page")
	}
	_, err = CommitFile(
		r,
		title,
		fmt.Sprintf("Wiki page %s removed.", title),
		os.Getenv("GIT_USERNAME"),
		os.Getenv("GIT_EMAIL"),
	)
	return errors.Wrap(err, "Can't commit wiki page change")
}

// Gets log for given file.
func FileLog(r *git.Repository, s string) (page.Log, error) {
	ref, err := r.Head()
	if err != nil {
		return nil, err
	}
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash(), FileName: &s})
	if err != nil {
		return nil, err
	}

	var wikiLog page.Log
	err = cIter.ForEach(func(c *gitObject.Commit) error {
		wikiLog = append(wikiLog, page.LogEntry{
			ID:      c.Hash.String(),
			Message: c.Message,
			Date:    c.Author.When,
		})
		return nil
	})
	return wikiLog, nil
}

//TODO: consider other name
// Commits file with given fileName (without extension) to given repository
// with commit message as userName with email
func CommitFile(r *git.Repository, fileName, message, userName, email string) (*page.LogEntry, error) {
	w, err := r.Worktree()
	if err != nil {
		return nil, errors.Wrap(err, "Can't prepare wiki worktree")
	}
	le := page.LogEntry{Message: message}
	w.Add(fileName + ".wiki")
	le.Date = time.Now()
	commit, err := w.Commit(message, &git.CommitOptions{
		Author: &gitObject.Signature{
			Name:  userName,
			Email: email,
			When:  le.Date,
		},
	})
	le.ID = commit.String()
	if err != nil {
		return nil, errors.Wrap(err, "Can't commit wiki page change")
	}
	return &le, nil
}

// Init git repo for wiki if doesn't exists.
func Init(wikiPath string) (repo *git.Repository, err error) {
	repo, err = git.PlainInit(wikiPath, false)
	if err != nil {
		err = errors.Wrap(err, "Can't init wiki repository")
	}
	return
}

// Get git repository for wiki.
func Wiki(wikiPath string) (repo *git.Repository, err error) {
	repo, err = git.PlainOpen(wikiPath)
	if err != nil {
		return nil, errors.Wrap(err, "Can't get wiki repository")
	}
	return repo, nil
}

// Get log for given repository
func WikiLog(r *git.Repository) (page.Log, error) {
	ref, err := r.Head()
	if err != nil {
		return nil, errors.Wrap(err, "cant get wiki repository head")
	}
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, errors.Wrap(err, "Can't get wiki log")
	}
	var wikiLog page.Log
	err = cIter.ForEach(func(c *gitObject.Commit) error {
		wikiLog = append(wikiLog, page.LogEntry{
			ID:      c.Hash.String(),
			Message: c.Message,
			Date:    c.Author.When,
		})
		return nil
	})
	return wikiLog, nil
}
