# Gopaste

## Synopsis

    go get github.com/wisnij/gopaste/gopasted
    cd $GOPATH/src/github.com/wisnij/gopaste
    $GOPATH/bin/gopasted [--source=gopaste.sqlite] [--port=80]

## Description

Gopaste is a simple pastebin written in Go.

### Features

- Syntax highlighting (courtesy of [highlight.js](http://highlightjs.org/))
- Paste annotation and diffs
- Private pastes

### Possible future features

- Full text search
- IRC bot integration
- Per-line comments
- RESTful API

## Author

Copyright (C) 2014 Jim Wisniewski <<wisnij@gmail.com>>.  Released under GNU
AGPLv3 (see [LICENSE](LICENSE) for full legalese).

Basic design inspired by/shamelessly stolen from
[lpaste](http://github.com/chrisdone/lpaste).
