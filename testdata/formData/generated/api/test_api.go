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
}{

	{
		Name: "UploadFile",
		Args: Args{Method: "Put", URL: "/file"},
		Want: Want{},
	},
}
