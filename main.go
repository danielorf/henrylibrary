package main

import "github.com/danielorf/henrylibrary/pkg/api"

func main() {
	var a api.App
	a.Initialize("test.db")
	a.FillSampleData()
	a.Run(":3000")
}
