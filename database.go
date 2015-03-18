package gopaste

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"sort"
	"strings"
)

const createTableSql = `
	CREATE TABLE IF NOT EXISTS pastes (
		id         INTEGER NOT NULL PRIMARY KEY,
		title      TEXT,
		content    TEXT NOT NULL,
		author     TEXT,
		language   TEXT,
		channel    TEXT,
		annotates  INTEGER,
		private    INTEGER NOT NULL,
		created    INTEGER NOT NULL
	);
`

// LanguageNames maps language identifers to the human-readable names of the
// languages supported by highlightjs.
var LanguageNames = map[string]string{
	"1c":             "1C",
	"actionscript":   "ActionScript",
	"apache":         "Apache",
	"applescript":    "AppleScript",
	"asciidoc":       "AsciiDoc",
	"aspectj":        "AspectJ",
	"autohotkey":     "AutoHotkey",
	"avrasm":         "AVR assembler",
	"axapta":         "Axapta",
	"bash":           "Bash",
	"brainfuck":      "Brainfuck",
	"c":              "C",
	"capnproto":      "Cap'n Proto",
	"clojure":        "Clojure",
	"cmake":          "CMake",
	"coffeescript":   "CoffeeScript",
	"cpp":            "C++",
	"cs":             "C#",
	"css":            "CSS",
	"d":              "D",
	"dart":           "Dart",
	"delphi":         "Delphi",
	"diff":           "Diff",
	"django":         "Django",
	"dos":            "DOS .bat",
	"dust":           "Dust",
	"elixir":         "Elixir",
	"erb":            "ERB (Embedded Ruby)",
	"erlang":         "Erlang",
	"fix":            "FIX",
	"fsharp":         "F#",
	"gcode":          "G-code (ISO 6983)",
	"gherkin":        "Gherkin",
	"glsl":           "OpenGL Shading Language",
	"go":             "Go",
	"gradle":         "Gradle",
	"groovy":         "Groovy",
	"haml":           "Haml",
	"handlebars":     "Handlebars",
	"haskell":        "Haskell",
	"haxe":           "Haxe",
	"html":           "HTML",
	"http":           "HTTP",
	"ini":            "INI",
	"java":           "Java",
	"javascript":     "JavaScript",
	"json":           "JSON",
	"lasso":          "Lasso",
	"less":           "Less",
	"lisp":           "Lisp",
	"livecodeserver": "LiveCode Server",
	"livescript":     "LiveScript",
	"lua":            "Lua",
	"makefile":       "Makefile",
	"markdown":       "Markdown",
	"mathematica":    "Mathematica",
	"matlab":         "Matlab",
	"mel":            "Maya Embedded Language",
	"mercury":        "Mercury",
	"mizar":          "Mizar",
	"monkey":         "Monkey",
	"nginx":          "Nginx",
	"nimrod":         "Nimrod",
	"nix":            "Nix",
	"nsis":           "NSIS",
	"objectivec":     "Objective C",
	"ocaml":          "OCaml",
	"oxygene":        "Oxygene",
	"p21":            "STEP Part 21 (ISO 10303-21)",
	"parser3":        "Parser3",
	"perl":           "Perl",
	"php":            "PHP",
	"powershell":     "PowerShell",
	"processing":     "Processing",
	"profile":        "Python profiler",
	"protobuf":       "Protocol Buffers",
	"puppet":         "Puppet",
	"python":         "Python",
	"q":              "Q",
	"r":              "R",
	"rib":            "RenderMan RIB",
	"rsl":            "RenderMan RSL",
	"ruby":           "Ruby",
	"ruleslanguage":  "Oracle Rules Language",
	"rust":           "Rust",
	"scala":          "Scala",
	"scheme":         "Scheme",
	"scilab":         "Scilab",
	"scss":           "SCSS",
	"smali":          "Smali",
	"smalltalk":      "Smalltalk",
	"sml":            "SML",
	"sql":            "SQL",
	"stata":          "Stata",
	"stylus":         "Stylus",
	"swift":          "Swift",
	"tcl":            "Tcl",
	"tex":            "TeX",
	"thrift":         "Thrift",
	"twig":           "Twig",
	"typescript":     "TypeScript",
	"vala":           "Vala",
	"vbnet":          "VB.Net",
	"vbscript":       "VBScript",
	"verilog":        "Verilog",
	"vhdl":           "VHDL",
	"vim":            "Vim Script",
	"x86asm":         "x86 Assembly",
	"xl":             "XL",
	"xml":            "XML",
}

type languageName struct {
	Name string
	Code string
}

// LanguageNamesSorted is a list of known languages, sorted by name.
var LanguageNamesSorted = []languageName{}

type byName []languageName

func (a byName) Len() int      { return len(a) }
func (a byName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool {
	return strings.ToUpper(a[i].Name) < strings.ToUpper(a[j].Name)
}

func init() {
	for c, n := range LanguageNames {
		LanguageNamesSorted = append(LanguageNamesSorted, languageName{Code: c, Name: n})
	}
	sort.Sort(byName(LanguageNamesSorted))
}

// initDb establishes Gopaste's database connection and creates the pastes table
// if necessary.
func (s *Server) initDb() error {
	dbh, err := sql.Open(s.Config.DbDriver, s.Config.DbSource)
	if err != nil {
		return fmt.Errorf("Error opening %s %s: %v\n", s.Config.DbDriver, s.Config.DbSource, err)
	}

	if _, err := dbh.Exec(createTableSql); err != nil {
		return err
	}

	s.Database = dbh
	return nil
}
