// model.go
package main

type Image struct {
	ID       string `bson:"id"`
	Metadata string `bson:"metadata"`
}
