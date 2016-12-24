package mongoconnector

import (
	"reflect"
	"fmt"
	"regexp"
)

const findRegexString = "FindBy(.*)From(.*)"
const saveRegexString = "SaveTo(.*)"


func resolveAndCreateFunc(fieldFunctionValue reflect.Value, name string) reflect.Value {
	r := regexp.MustCompile(findRegexString)
	found := r.FindAllStringSubmatch(name, 1)
	//Find found
	if len(found) > 0 {

	}

	r = regexp.MustCompile(saveRegexString)
	found := r.FindAllStringSubmatch(name, 1)
	//Save found
	if len(found) > 0 {

	}

	fmt.Println(found[0][1])

	fn := func (args []reflect.Value) []reflect.Value {
		theString := args[0].String()
		fmt.Println(theString)

		return args
	}
	resultFunctionValue := reflect.MakeFunc(fieldFunctionValue.Type(), fn)
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
