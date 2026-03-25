package db

import "time"

type WordPosting struct {
	Word      string
	URLHash   string
	Domain    string
	Frequency int
}

type WordIndex struct {
	ID          int
	Word        string
	Domain      string
	Frequency   int
	LastUpdated time.Time
}
