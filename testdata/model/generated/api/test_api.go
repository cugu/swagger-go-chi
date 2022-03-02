package api

import "time"

type Args struct {
	Method string
	URL    string
	Data   interface{}
}
type Want struct {
	Status int
	Body   interface{}
}

var Tests = []struct {
	Name string
	Args Args
	Want Want
}{}
