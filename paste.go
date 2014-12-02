package gopaste

import (
	"database/sql"
	"fmt"
	"github.com/kisielk/sqlstruct"
	"math"
	"math/rand"
	"net/url"
	"strings"
	"time"
)

const (
	InvalidPasteId = -1
	TimeFormat     = "2006-01-02 15:04:05 -07:00"
)

// Paste represents an individual paste.
type Paste struct {
	Id            int64          `sql:"id"`
	Title         sql.NullString `sql:"title"`
	Content       string         `sql:"content"`
	Author        sql.NullString `sql:"author"`
	Language      sql.NullString `sql:"language"`
	Channel       sql.NullString `sql:"channel"`
	Annotates     sql.NullInt64  `sql:"annotates"`
	Private       bool           `sql:"private"`
	Created       int64          `sql:"created"`
	AnnotationNum int            `sql:"-"`
}

// NewPaste creates a new paste from a submitted web form.
func NewPaste(v url.Values) *Paste {
	paste := &Paste{
		Content: v.Get("Content"),
		Private: (v.Get("Private") == "on"),
		Created: time.Now().Unix(),
	}

	if s := strings.TrimSpace(v.Get("Title")); s != "" {
		paste.Title.Valid = true
		paste.Title.String = s
	}

	if s := strings.TrimSpace(v.Get("Author")); s != "" {
		paste.Author.Valid = true
		paste.Author.String = s
	}

	if s := v.Get("Language"); s != "" {
		paste.Language.Valid = true
		paste.Language.String = s
	}

	if s := v.Get("Channel"); s != "" {
		paste.Channel.Valid = true
		paste.Channel.String = s
	}

	return paste
}

// TitleDef returns the paste title if set, or "untitled" otherwise.
func (p Paste) TitleDef() string {
	if p.Title.Valid {
		return p.Title.String
	}
	return "untitled"
}

// ReplyTitle returns the paste title prefixed with "Re:" if it doesn't already
// begin that way.
func (p Paste) ReplyTitle() string {
	title := p.TitleDef()
	if !strings.HasPrefix(title, "Re: ") {
		title = "Re: " + title
	}
	return title
}

// AuthorDef returns the paste author if set, or "anonymous" otherwise.
func (p Paste) AuthorDef() string {
	if p.Author.Valid {
		return p.Author.String
	}
	return "anonymous"
}

// LanguageDef returns the full name of the paste language if set, or "plain text" otherwise.
func (p Paste) LanguageDef() string {
	if p.Language.Valid {
		code := p.Language.String
		if name, ok := LanguageNames[code]; ok {
			return name
		}
		return code
	}
	return "plain text"
}

// ChannelDef returns the paste channel if set, or the empty string otherwise.
func (p Paste) ChannelDef() string {
	if p.Channel.Valid {
		return p.Channel.String
	}
	return ""
}

// RootId returns the ID of the paste this one annotates, or this paste's ID if
// this is a top-level paste.
func (p Paste) RootId() int64 {
	if p.Annotates.Valid {
		return p.Annotates.Int64
	} else {
		return p.Id
	}
}

const (
	Minute = 60
	Hour   = 60 * Minute
	Day    = 24 * Hour
	Week   = 7 * Day
	Month  = 30 * Day
	Year   = 365 * Day
)

type timeThreshold struct {
	seconds float64
	unit    string
}

var timeThresholds = []timeThreshold{
	{0, "just now"},
	{Minute, "minute"},
	{Hour, "hour"},
	{Day, "day"},
	{Week, "week"},
	{Month, "month"},
	{Year, "year"},
}

// CreatedTime returns the paste's creation time as a time.Time object.
func (p Paste) CreatedTime() time.Time {
	return time.Unix(p.Created, 0)
}

// CreatedDisplay returns the paste creation date in a human-readable format.
func (p Paste) CreatedDisplay() string {
	return p.CreatedTime().Format(TimeFormat)
}

// CreatedRel returns a string describing how long ago the paste was created, in
// a human-friendly format (e.g. "3 days ago").
func (p Paste) CreatedRel() string {
	relString := "ago"
	secondsAgo := time.Since(p.CreatedTime()).Seconds()
	if secondsAgo < 0 {
		secondsAgo *= -1
		relString = "from now"
	}

	i := 0
	for ; i+1 < len(timeThresholds); i++ {
		if timeThresholds[i+1].seconds > secondsAgo {
			break
		}
	}

	unit := timeThresholds[i].unit
	if i == 0 {
		return unit
	}

	unitsAgo := secondsAgo / timeThresholds[i].seconds
	if unitsAgo >= 2 {
		return fmt.Sprintf("%d %ss %s", int(unitsAgo), unit, relString)
	}

	var article string
	if unit[0] == 'h' {
		article = "an"
	} else {
		article = "a"
	}

	return fmt.Sprintf("%s %s %s", article, unit, relString)
}

//
type LineNumber struct {
	Num    int
	Anchor string
}

// LineNumbers returns a list of LineNumber objects for a paste.
func (p Paste) LineNumbers() (ns []LineNumber) {
	for i := 0; i <= strings.Count(strings.TrimSuffix(p.Content, "\n"), "\n"); i++ {
		n := i + 1
		var s string
		if p.AnnotationNum > 0 {
			s = fmt.Sprintf("%d.%d", p.AnnotationNum, n)
		} else {
			s = fmt.Sprint(n)
		}
		ns = append(ns, LineNumber{Num: n, Anchor: s})
	}
	return ns
}

// publicId returns the next available public paste ID.
func publicId(dbh *sql.DB) (int64, error) {
	var num int64
	err := dbh.QueryRow("SELECT COALESCE(MAX(id), 0) FROM pastes WHERE NOT private").Scan(&num)
	if err != nil {
		return InvalidPasteId, err
	}
	return num + 1, nil
}

// privateIdBase is the smallest private paste ID.
const privateIdBase = 1 << 62

var privateIdRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// privateId returns a random number in the range [1<<62, 1<<63).
func privateId() int64 {
	return privateIdBase + privateIdRand.Int63n(privateIdBase)
}

func InsertPaste(dbh *sql.DB, paste *Paste) (int64, error) {
	tx, err := dbh.Begin()
	if err != nil {
		return InvalidPasteId, err
	}

	if paste.Id == 0 {
		if paste.Private {
			paste.Id = privateId()
		} else {
			id, err := publicId(dbh)
			if err != nil {
				return InvalidPasteId, err
			}
			paste.Id = id
		}
	}

	query := `
		INSERT INTO pastes (id, title, content, author, language,
		                    channel, annotates, private, created)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
	_, err = tx.Exec(query,
		paste.Id, paste.Title, paste.Content, paste.Author, paste.Language,
		paste.Channel, paste.Annotates, paste.Private, paste.Created,
	)

	if err != nil {
		tx.Rollback()
		return InvalidPasteId, err
	}

	err = tx.Commit()
	if err != nil {
		return InvalidPasteId, err
	}

	return paste.Id, nil
}

// GetPaste fetches a single paste from its ID.
func GetPaste(dbh *sql.DB, pasteId int64) (*Paste, error) {
	query := fmt.Sprintf("SELECT %s FROM pastes WHERE id = ?", sqlstruct.Columns(Paste{}))
	rows, err := dbh.Query(query, pasteId)
	if err != nil || !rows.Next() {
		return nil, err
	}

	defer rows.Close()
	paste := &Paste{}
	if err = sqlstruct.Scan(paste, rows); err != nil {
		return nil, err
	}

	annotation, err := AnnotationOrdinal(dbh, pasteId)
	if err != nil {
		return nil, err
	}

	paste.AnnotationNum = annotation
	return paste, nil
}

// GetAnnotations fetches all annotations of the paste with the given ID.
func GetAnnotations(dbh *sql.DB, pasteId int64) ([]*Paste, error) {
	query := fmt.Sprintf("SELECT %s FROM pastes WHERE annotates = ? ORDER BY id", sqlstruct.Columns(Paste{}))
	rows, err := dbh.Query(query, pasteId)
	if err != nil {
		return nil, err
	}

	annotations := []*Paste{}
	for rows.Next() {
		paste := &Paste{}
		if err = sqlstruct.Scan(paste, rows); err != nil {
			return nil, err
		}
		annotations = append(annotations, paste)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	for i := range annotations {
		annotations[i].AnnotationNum = i + 1
	}

	return annotations, nil
}

// AnnotationOrdinal returns N such that the paste is the Nth annotation of its parent.
func AnnotationOrdinal(dbh *sql.DB, pasteId int64) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM pastes p
		    LEFT JOIN pastes o ON o.annotates = p.annotates
		                      AND o.id <= p.id
		WHERE p.id = ? AND p.annotates IS NOT NULL
	`
	var num int
	err := dbh.QueryRow(query, pasteId).Scan(&num)
	return num, err
}

type PastePage struct {
	Total  int
	Start  int
	End    int
	Pastes []*PasteData
}

func cdiv(n, d int) int {
	return int(math.Ceil(float64(n) / float64(d)))
}

func (p *PastePage) PageCount(pageSize int) int {
	return cdiv(p.Total, pageSize)
}

// TopLevelPastes fetches the paste IDs for all pastes which are not private or
// annotations
func TopLevelPastes(dbh *sql.DB, opts *BrowseOpts) (*PastePage, error) {
	commonSql := "FROM pastes WHERE NOT private AND annotates IS NULL"

	var parameters []interface{}
	if author, ok := opts.Search["author"]; ok {
		commonSql += " AND author = ?"
		parameters = append(parameters, author)
	}

	if channel, ok := opts.Search["channel"]; ok {
		commonSql += " AND channel = ?"
		parameters = append(parameters, channel)
	}

	if language, ok := opts.Search["language"]; ok {
		commonSql += " AND language = ?"
		parameters = append(parameters, language)
	}

	page := &PastePage{}

	countSql := "SELECT COUNT(*) " + commonSql
	countRow := dbh.QueryRow(countSql, parameters...)
	err := countRow.Scan(&page.Total)
	if err != nil {
		return nil, err
	}

	offset := (opts.Page - 1) * opts.PageSize
	querySql := "SELECT id " + commonSql + " ORDER BY id DESC"
	querySql += fmt.Sprintf(" LIMIT %d OFFSET %d", opts.PageSize, offset)

	rows, err := dbh.Query(querySql, parameters...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var pasteId int64
		if err = rows.Scan(&pasteId); err != nil {
			return nil, err
		}

		data, err := GetPasteData(dbh, pasteId)
		if err != nil {
			return nil, err
		}

		page.Pastes = append(page.Pastes, data)
	}

	pageCount := len(page.Pastes)
	if pageCount > 0 {
		page.Start = offset + 1
		page.End = offset + pageCount
	}

	return page, nil
}

// PasteData represents a paste and all of its annotations.
type PasteData struct {
	Paste       *Paste
	Annotations []*Paste
}

// GetPasteData fetches a paste and its annotations from the given paste ID.
func GetPasteData(dbh *sql.DB, pasteId int64) (*PasteData, error) {
	paste, err := GetPaste(dbh, pasteId)
	if err != nil || paste == nil {
		return nil, err
	}

	data := &PasteData{Paste: paste}
	annotations, err := GetAnnotations(dbh, pasteId)
	if err != nil {
		return nil, err
	}

	data.Annotations = annotations
	return data, nil
}

type PasteView struct {
	*Paste
	Top  *Paste
	Prev *Paste
}

func (d PasteData) PasteView() *PasteView {
	return &PasteView{Paste: d.Paste}
}

func (d PasteData) AnnotationsView() (view []PasteView) {
	var prev *Paste
	for _, ann := range d.Annotations {
		view = append(view, PasteView{
			Paste: ann,
			Top:   d.Paste,
			Prev:  prev,
		})
		prev = ann
	}
	return
}
