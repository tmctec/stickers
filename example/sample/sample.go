// Package sample holds the data the table demos show.
package sample

import _ "embed"

// CSV is a list of people in CSV form. Its first line names the columns: id, First Name,
// Last Name, Age and Occupation.
//
//go:embed sample.csv
var CSV []byte
