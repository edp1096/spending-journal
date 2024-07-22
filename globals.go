package server

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/dgraph-io/badger/v3"
)

var listenIP = "127.0.0.1"
var listenPORT = "8080"
var listenADDR = listenIP + ":" + listenPORT

var db *badger.DB
var bleveIndex bleve.Index

var PayTypesDirect = []string{"현금", "체크카드", "직불카드"}
var PayTypesCredit = []string{"신용카드"}
