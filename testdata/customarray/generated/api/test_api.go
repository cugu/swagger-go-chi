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
		Name: "CreateUserBatch",
		Args: Args{Method: "Post", URL: "/users", Data: []interface{}{map[string]interface{}{"name": "bob"}}},
		Want: Want{
			Status: 204,
			Body:   nil,
		},
	},
}
