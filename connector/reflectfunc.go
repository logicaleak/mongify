package mongoconnector

import (
	"reflect"

	"regexp"
	"strings"
	"gopkg.in/mgo.v2/bson"
	"fmt"
)

const findRegexString = "FindBy(.*)From(.*)"
const saveRegexString = "SaveTo(.*)"

type Error struct {
	err error
}

func (self Error) Error() string {
	return self.err.Error()
}

func generateFindFunction(fieldFunctionValue reflect.Value, name string) reflect.Value {
	r := regexp.MustCompile(findRegexString)
	found := r.FindAllStringSubmatch(name, 1)
	field := found[0][1]
	collection := found[0][2]
	fn := func (args []reflect.Value) []reflect.Value {
		database := mongoConnectorInstance.GetDatabase()
		defer database.Session.Close()

		selector := bson.M{
			field : args[0].Interface(),
		}

		newValue := reflect.New(fieldFunctionValue.Type().Out(0))
		newValueInterface := newValue.Interface()
		fmt.Println(reflect.TypeOf(newValueInterface))
		//todo Can separate this portion to make it possible to be usable by literal string function calling?
		err := database.C(collection).Find(selector).One(newValueInterface)

		secondValue := reflect.ValueOf(Error{err: err})

		return []reflect.Value {
			newValue.Elem(),
			secondValue,
		}
	}

	resultFunctionValue := reflect.MakeFunc(fieldFunctionValue.Type(), fn)

	return resultFunctionValue
}


func resolveAndCreateFunc(fieldFunctionValue reflect.Value, name string) reflect.Value {
	var resultFunctionValue reflect.Value
	if strings.Contains(name, "Find") {
		resultFunctionValue = generateFindFunction(fieldFunctionValue, name)
	}
	return resultFunctionValue
}

func Implement(toImplement interface{}) interface{} {
	value := reflect.ValueOf(toImplement).Elem()

	for i := 0 ; i < value.Type().NumField(); i ++ {
		fieldValue := value.Field(i)
		resultFunction := resolveAndCreateFunc(fieldValue, value.Type().Field(0).Name)
		fieldValue.Set(resultFunction)
	}

	return value.Interface()
}
