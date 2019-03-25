package main

func main() {
	var app App
	app.Initialize("test.db")
	app.FillSampleData()
	app.Run(":3000")
}
