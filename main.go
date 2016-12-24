package main

import (

)
import (
	"mongoconnector/connector"
)

type A struct {
	FindByThing func(string) string
}


func main() {
	a := A{}
	returnedInterface := mongoconnector.Implement(&a)
	newA := returnedInterface.(A)
	newA.FindByThing("fuck")
}