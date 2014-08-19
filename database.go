package gopaste

import (
	"database/sql"
)

var LanguageNames = map[string]string{
	"1c":             "1C",
	"actionscript":   "ActionScript",
	"apache":         "Apache",
	"applescript":    "AppleScript",
	"asciidoc":       "AsciiDoc",
	"autohotkey":     "AutoHotkey",
	"avrasm":         "AVR assembler",
	"axapta":         "Axapta",
	"bash":           "Bash",
	"brainfuck":      "Brainfuck",
	"c":              "C",
	"clojure":        "Clojure",
	"cmake":          "CMake",
	"coffeescript":   "CoffeeScript",
	"cpp":            "C++",
	"cs":             "C#",
	"css":            "CSS",
	"d":              "D",
	"delphi":         "Delphi",
	"diff":           "Diff",
	"django":         "Django",
	"dos":            "DOS .bat",
	"elixir":         "Elixir",
	"erlang":         "Erlang",
	"fix":            "FIX",
	"fsharp":         "F#",
	"gherkin":        "Gherkin",
	"glsl":           "OpenGL Shading Language",
	"go":             "Go",
	"haml":           "Haml",
	"handlebars":     "Handlebars",
	"haskell":        "Haskell",
	"html":           "HTML",
	"http":           "HTTP",
	"ini":            "INI",
	"java":           "Java",
	"javascript":     "JavaScript",
	"json":           "JSON",
	"lasso":          "Lasso",
	"lisp":           "Lisp",
	"livecodeserver": "LiveCode Server",
	"lua":            "Lua",
	"makefile":       "Makefile",
	"markdown":       "Markdown",
	"mathematica":    "Mathematica",
	"matlab":         "Matlab",
	"mel":            "Maya Embedded Language",
	"mizar":          "Mizar",
	"monkey":         "Monkey",
	"nginx":          "Nginx",
	"nix":            "Nix",
	"nsis":           "NSIS",
	"objectivec":     "Objective C",
	"ocaml":          "OCaml",
	"oxygene":        "Oxygene",
	"parser3":        "Parser3",
	"perl":           "Perl",
	"php":            "PHP",
	"profile":        "Python profiler",
	"protobuf":       "Protocol Buffers",
	"python":         "Python",
	"r":              "R",
	"rib":            "RenderMan RIB",
	"rsl":            "RenderMan RSL",
	"ruby":           "Ruby",
	"ruleslanguage":  "Oracle Rules Language",
	"rust":           "Rust",
	"scala":          "Scala",
	"scilab":         "Scilab",
	"scss":           "SCSS",
	"smalltalk":      "Smalltalk",
	"sql":            "SQL",
	"tex":            "TeX",
	"vala":           "Vala",
	"vbnet":          "VB.Net",
	"vbscript":       "VBScript",
	"vhdl":           "VHDL",
	"vim":            "Vim Script",
	"x86asm":         "x86 Assembly",
	"xml":            "XML",
}

func initDb(dbh *sql.DB) error {
	sql := `
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

	if _, err := dbh.Exec(sql); err != nil {
		return err
	}

	return nil
}
