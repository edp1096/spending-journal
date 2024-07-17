package server

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/dgraph-io/badger/v3"
)

var db *badger.DB
var bleveIndex bleve.Index

var directPays = []string{"현금", "체크카드", "직불카드"}
var creditPays = []string{"신용카드"}
