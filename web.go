package gopaste

import (
	"bytes"
	"fmt"
	"github.com/aryann/difflib"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var tmpl *template.Template

func trunc(s string, max int) string {
	if len(s) < max {
		return s
	}

	last := max - 3
	return s[:last] + "..."
}

func init() {
	tmpl = template.New("web")
	tmpl.Funcs(template.FuncMap{
		"trunc": trunc,
	})
	if _, err := tmpl.ParseGlob("*.template"); err != nil {
		log.Fatalf("template parsing: %v\n", err)
	}

}

////////////////////////////////////////////////////////////////////////////////

type Query struct {
	Request  *http.Request
	Response http.ResponseWriter
	Action   string
	Args     []string
}

func NewQuery(w http.ResponseWriter, req *http.Request) *Query {
	data := &Query{
		Request:  req,
		Response: w,
	}

	path := req.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")

	data.Action = parts[0]
	if len(parts) > 1 {
		data.Args = parts[1:]
	}

	return data
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("[web] %s %s %s", req.RemoteAddr, req.Method, req.URL.Path)

	q := NewQuery(w, req)
	if err := s.handle(q); err != nil {
		if e, ok := err.(HttpError); ok {
			http.Error(w, e.Error(), e.Code)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

type ActionFunc func(*Server, *Query) error

var handlers = map[string]ActionFunc{
	"":         (*Server).doMain,
	"annotate": (*Server).doAnnotate,
	"browse":   (*Server).doBrowse,
	"diff":     (*Server).doDiff,
	"new":      (*Server).doNew,
	"raw":      (*Server).doRaw,
	"static":   (*Server).doStatic,
	"view":     (*Server).doView,
}

func (s *Server) handle(d *Query) error {
	handler := handlers[d.Action]
	if handler == nil {
		return HttpError{fmt.Sprintf("'%s' not found", d.Request.URL), http.StatusNotFound}
	}

	return handler(s, d)
}

////////////////////////////////////////////////////////////////////////////////

type AnyMap map[string]interface{}

type HttpError struct {
	Message string
	Code    int
}

func (e HttpError) Error() string {
	return fmt.Sprintf("ERROR %d: %s", e.Code, e.Message)
}

func parsePasteId(str string) (int64, error) {
	id, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return InvalidPasteId, fmt.Errorf("invalid paste id '%s'", str)
	}
	return id, nil
}

// runTemplate executes a template and writes the results as HTML if successful
func runTemplate(w http.ResponseWriter, name string, data interface{}) error {
	buf := new(bytes.Buffer)
	err := tmpl.ExecuteTemplate(buf, name, data)
	if err != nil {
		return HttpError{fmt.Sprintf("error processing template %s: %v", name, err), http.StatusInternalServerError}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func (s *Server) doMain(q *Query) error {
	opts := NewBrowseOpts()
	opts.PageSize = 10

	page, err := TopLevelPastes(s.Database, opts)
	if err != nil {
		return HttpError{err.Error(), http.StatusInternalServerError}
	}

	return runTemplate(q.Response, "main", AnyMap{
		"MainPage":  true,
		"Title":     "Home",
		"Page":      page,
		"Languages": LanguageNames,
		"Channels":  []string{}, // TODO
	})
}

////////////////////////////////////////////////////////////////////////////////

type BrowseOpts struct {
	Page     int
	PageSize int
	Search   map[string]string
}

func NewBrowseOpts() *BrowseOpts {
	return &BrowseOpts{
		Page:     1,
		PageSize: 50,
		Search:   make(map[string]string),
	}
}

func (o *BrowseOpts) Parse(args []string) error {
	for i := 0; i+1 < len(args); i += 2 {
		key, err := url.QueryUnescape(args[i])
		if err != nil {
			return err
		}
		val, err := url.QueryUnescape(args[i+1])
		if err != nil {
			return err
		}
		switch key {
		case "page":
			pagenum, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid page number: %s", val)
			}
			o.Page = int(pagenum)
		default:
			o.Search[key] = val
		}
	}

	return nil
}

func (o *BrowseOpts) String() string {
	var parts, keys []string

	for key := range o.Search {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		parts = append(parts,
			url.QueryEscape(key),
			url.QueryEscape(o.Search[key]),
		)
	}

	if o.Page > 1 {
		parts = append(parts, "page", fmt.Sprint(o.Page))
	}

	return strings.Join(parts, "/")
}

func (o *BrowseOpts) NewPage(page int) *BrowseOpts {
	newOpts := NewBrowseOpts()
	newOpts.Page = page
	newOpts.PageSize = o.PageSize
	for k, v := range o.Search {
		newOpts.Search[k] = v
	}
	return newOpts
}

func (o *BrowseOpts) Prev() *BrowseOpts {
	return o.NewPage(o.Page - 1)
}

func (o *BrowseOpts) Next() *BrowseOpts {
	return o.NewPage(o.Page + 1)
}

func (o *BrowseOpts) Nearby(window int, max int) (pages []int) {
	min := 1
	var low, high int

	width := 2*window + 1
	if width >= max {
		low = min
		high = max
	} else {
		low = o.Page - window
		lowExtra := 0
		if low < min {
			lowExtra = min - low
			low = min
		}

		high = o.Page + window + lowExtra
		if high > max {
			low -= high - max
			high = max
		}
	}

	for n := low; n <= high; n++ {
		pages = append(pages, n)
	}

	return
}

func (s *Server) doBrowse(q *Query) error {
	opts := NewBrowseOpts()
	err := opts.Parse(q.Args)
	if err != nil {
		return HttpError{err.Error(), http.StatusBadRequest}
	}

	page, err := TopLevelPastes(s.Database, opts)
	if err != nil {
		return HttpError{err.Error(), http.StatusInternalServerError}
	}

	return runTemplate(q.Response, "browse", AnyMap{
		"Title": "Browse pastes",
		"Page":  page,
		"Opts":  opts,
	})
}

////////////////////////////////////////////////////////////////////////////////

// doDiff displays the difference between two pastes.
func (s *Server) doDiff(q *Query) error {
	if len(q.Args) < 2 {
		return HttpError{"invalid request", http.StatusBadRequest}
	}

	fromStr, toStr := q.Args[0], q.Args[1]

	fromId, err := parsePasteId(fromStr)
	if err != nil {
		return HttpError{err.Error(), http.StatusBadRequest}
	}

	toId, err := parsePasteId(toStr)
	if err != nil {
		return HttpError{err.Error(), http.StatusBadRequest}
	}

	from, err := GetPaste(s.Database, fromId)
	if err != nil {
		return HttpError{err.Error(), http.StatusInternalServerError}
	}
	if from == nil {
		return HttpError{fmt.Sprintf("paste %d not found", fromId), http.StatusNotFound}
	}

	to, err := GetPaste(s.Database, toId)
	if err != nil {
		return HttpError{err.Error(), http.StatusInternalServerError}
	}
	if to == nil {
		return HttpError{fmt.Sprintf("paste %d not found", toId), http.StatusNotFound}
	}

	diff := difflib.Diff(
		strings.Split(from.Content, "\n"),
		strings.Split(to.Content, "\n"),
	)

	diffText := ""
	for _, line := range diff {
		var prefix string
		switch line.Delta {
		case difflib.LeftOnly:
			prefix = "-"
		case difflib.RightOnly:
			prefix = "+"
		default:
			prefix = " "
		}
		diffText += prefix + line.Payload + "\n"
	}

	return runTemplate(q.Response, "diff", AnyMap{
		"Title":    fmt.Sprintf("Diff #%d / #%d", from.Id, to.Id),
		"From":     from,
		"To":       to,
		"DiffText": diffText,
	})
}

////////////////////////////////////////////////////////////////////////////////

// doNew adds a new top-level paste.
func (s *Server) doNew(q *Query) error {
	return s.handleNew(q, nil)
}

// doAnnotation adds a new annotation to an existing paste.
func (s *Server) doAnnotate(q *Query) error {
	if len(q.Args) < 1 {
		return HttpError{"invalid request", http.StatusBadRequest}
	}

	idStr := q.Args[0]
	id, err := parsePasteId(idStr)
	if err != nil {
		return HttpError{err.Error(), http.StatusBadRequest}
	}

	paste, err := GetPaste(s.Database, id)
	if err != nil {
		return HttpError{err.Error(), http.StatusInternalServerError}
	}
	if paste == nil {
		return HttpError{fmt.Sprintf("paste %d not found", id), http.StatusNotFound}
	}

	return s.handleNew(q, paste)
}

func (s *Server) handleNew(q *Query, parent *Paste) error {
	method := q.Request.Method
	switch method {
	case "GET", "HEAD":
		return s.displayNewPage(q, parent)
	case "POST":
		return s.insertNewPaste(q, parent)
	default:
		return HttpError{fmt.Sprintf("unsupported request method: %s", method), http.StatusNotImplemented}
	}
}

func (s *Server) displayNewPage(q *Query, parent *Paste) error {
	var title string
	if parent != nil {
		title = fmt.Sprintf("Annotating #%d: %s", parent.Id, parent.TitleDef())
	} else {
		title = "New paste"
	}

	return runTemplate(q.Response, "new", AnyMap{
		"Title":     title,
		"Annotates": parent,
		"Languages": LanguageNames,
		"Channels":  []string{}, // TODO
	})
}

func (s *Server) insertNewPaste(q *Query, parent *Paste) error {
	err := q.Request.ParseForm()
	if err != nil {
		return HttpError{fmt.Sprintf("error parsing form: %s", err.Error()), http.StatusInternalServerError}
	}

	paste := NewPaste(q.Request.PostForm)
	if parent != nil {
		paste.Annotates.Int64 = parent.RootId()
		paste.Annotates.Valid = true
		paste.Private = parent.Private
	}

	pasteId, err := InsertPaste(s.Database, paste)
	if err != nil {
		return HttpError{fmt.Sprintf("error inserting new paste: %s", err.Error()), http.StatusInternalServerError}
	}

	var newPath string
	if parent != nil {
		annotation, err := AnnotationOrdinal(s.Database, pasteId)
		if err != nil {
			return HttpError{fmt.Sprintf("error fetching paste %d: %s", pasteId, err.Error()), http.StatusInternalServerError}
		}

		newPath = fmt.Sprintf("/view/%d#a%d", parent.Id, annotation)
	} else {
		newPath = fmt.Sprintf("/view/%d", pasteId)
	}

	http.Redirect(q.Response, q.Request, newPath, http.StatusSeeOther)
	return nil
}

////////////////////////////////////////////////////////////////////////////////

// doRaw returns the verbatim content of a paste as plain text.
func (s *Server) doRaw(q *Query) error {
	if len(q.Args) < 1 {
		return HttpError{"invalid request", http.StatusBadRequest}
	}

	idStr := q.Args[0]
	id, err := parsePasteId(idStr)
	if err != nil {
		return HttpError{err.Error(), http.StatusBadRequest}
	}

	paste, err := GetPaste(s.Database, id)
	if err != nil {
		return HttpError{err.Error(), http.StatusInternalServerError}
	}
	if paste == nil {
		return HttpError{fmt.Sprintf("paste %d not found", id), http.StatusNotFound}
	}

	q.Response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(q.Response, paste.Content)

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// doStatic serves a single static file.
func (s *Server) doStatic(q *Query) error {
	if len(q.Args) < 1 {
		return HttpError{"invalid request", http.StatusBadRequest}
	}

	path := []string{"static"}
	path = append(path, q.Args...)
	http.ServeFile(q.Response, q.Request, filepath.Join(path...))

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// doView displays a paste and any annotations with syntax highlighting.
func (s *Server) doView(q *Query) error {
	if len(q.Args) < 1 {
		return HttpError{"invalid request", http.StatusBadRequest}
	}

	idStr := q.Args[0]
	id, err := parsePasteId(idStr)
	if err != nil {
		return HttpError{err.Error(), http.StatusBadRequest}
	}

	pasteData, err := GetPasteData(s.Database, id)
	if err != nil {
		return HttpError{err.Error(), http.StatusInternalServerError}
	}
	if pasteData == nil {
		return HttpError{fmt.Sprintf("paste %d not found", id), http.StatusNotFound}
	}

	return runTemplate(q.Response, "view", AnyMap{
		"Title":   fmt.Sprintf("Paste #%d: %s", pasteData.Paste.Id, pasteData.Paste.TitleDef()),
		"Content": pasteData,
	})
}
