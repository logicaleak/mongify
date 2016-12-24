package main

import (

)
import (
	"mongoconnector/connector"
	"fmt"
)


type TheCollection struct {
	Thing string `bson:"Thing"`
}

type A struct {
	FindByThingFromTheCollection func(string) (TheCollection, mongoconnector.Error)
}


func main() {
	mongoconnector.InitializeMongoConnectorSingleton("localhost:27017", "rtest", "coll", "coll2")
	a := A{}
	returnedInterface := mongoconnector.Implement(&a)
	newA := returnedInterface.(A)
	tc, _ := newA.FindByThingFromTheCollection("thing")
	fmt.Println(tc)
}