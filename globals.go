package server

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/dgraph-io/badger/v3"
)

// var listenIP = "0.0.0.0"
var listenIP = "127.0.0.1"
var listenPORT = "12480"
var listenADDR = listenIP + ":" + listenPORT

var db *badger.DB
var bleveIndex bleve.Index

var ExchangeRate float64 = 1390
