package main

import "apigw/route"

func main() {
	r := route.Router()
	r.Run(":8080")
}
