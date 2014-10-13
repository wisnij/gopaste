package gopaste

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
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
		created    TEXT NOT NULL
	);
`

type languageName struct {
	Name string
	Code string
}

// LanguageNamesSorted is a list of known languages, sorted by name.
var LanguageNamesSorted = []languageName{
	{"1C", "1c"},
	{"ActionScript", "actionscript"},
	{"Apache", "apache"},
	{"AppleScript", "applescript"},
	{"AsciiDoc", "asciidoc"},
	{"AutoHotkey", "autohotkey"},
	{"AVR assembler", "avrasm"},
	{"Axapta", "axapta"},
	{"Bash", "bash"},
	{"Brainfuck", "brainfuck"},
	{"C", "c"},
	{"C#", "cs"},
	{"C++", "cpp"},
	{"Cap'n Proto", "capnproto"},
	{"Clojure", "clojure"},
	{"CMake", "cmake"},
	{"CoffeeScript", "coffeescript"},
	{"CSS", "css"},
	{"D", "d"},
	{"Dart", "dart"},
	{"Delphi", "delphi"},
	{"Diff", "diff"},
	{"Django", "django"},
	{"DOS .bat", "dos"},
	{"Dust", "dust"},
	{"Elixir", "elixir"},
	{"Erlang", "erlang"},
	{"F#", "fsharp"},
	{"FIX", "fix"},
	{"G-Code", "gcode"},
	{"Gherkin", "gherkin"},
	{"Go", "go"},
	{"Gradle", "gradle"},
	{"Groovy", "groovy"},
	{"Haml", "haml"},
	{"Handlebars", "handlebars"},
	{"Haskell", "haskell"},
	{"Haxe", "haxe"},
	{"HTML", "html"},
	{"HTTP", "http"},
	{"INI", "ini"},
	{"Java", "java"},
	{"JavaScript", "javascript"},
	{"JSON", "json"},
	{"Lasso", "lasso"},
	{"Lisp", "lisp"},
	{"LiveCode Server", "livecodeserver"},
	{"Lua", "lua"},
	{"Makefile", "makefile"},
	{"Markdown", "markdown"},
	{"Mathematica", "mathematica"},
	{"Matlab", "matlab"},
	{"Maya Embedded Language", "mel"},
	{"Mizar", "mizar"},
	{"Monkey", "monkey"},
	{"Nginx", "nginx"},
	{"Nimrod", "nimrod"},
	{"Nix", "nix"},
	{"NSIS", "nsis"},
	{"Objective C", "objectivec"},
	{"OCaml", "ocaml"},
	{"OpenGL Shading Language", "glsl"},
	{"Oracle Rules Language", "ruleslanguage"},
	{"Oxygene", "oxygene"},
	{"Parser3", "parser3"},
	{"Perl", "perl"},
	{"PHP", "php"},
	{"Protocol Buffers", "protobuf"},
	{"Python", "python"},
	{"Python profiler", "profile"},
	{"Q", "q"},
	{"R", "r"},
	{"RenderMan RIB", "rib"},
	{"RenderMan RSL", "rsl"},
	{"Ruby", "ruby"},
	{"Rust", "rust"},
	{"Scala", "scala"},
	{"Scheme", "scheme"},
	{"Scilab", "scilab"},
	{"SCSS", "scss"},
	{"Smalltalk", "smalltalk"},
	{"SQL", "sql"},
	{"Swift", "swift"},
	{"TeX", "tex"},
	{"Thrift", "thrift"},
	{"TypeScript", "typescript"},
	{"Vala", "vala"},
	{"VB.Net", "vbnet"},
	{"VBScript", "vbscript"},
	{"VHDL", "vhdl"},
	{"Vim Script", "vim"},
	{"x86 Assembly", "x86asm"},
	{"XML", "xml"},
}

// LanguageNames maps language identifers to the human-readable names of the
// languages supported by highlightjs.
var LanguageNames = map[string]string{}

func init() {
	for _, v := range LanguageNamesSorted {
		LanguageNames[v.Code] = v.Name
	}
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
