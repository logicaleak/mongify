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

func createLogicalRegex(name string, logic string) string {
	count := strings.Count(name, logic)
	wordMatch := "([a-z|A-Z|0-9]*)"
	regex := "findBy" + wordMatch
	for i := 0; i < count; i++ {
		regex += logic + wordMatch
	}
	return regex
}



func generateFindFunction(fieldFunctionValue reflect.Value, name string, logic string) reflect.Value {
	var r *regexp.Regexp
	if logic == "" {
		r = regexp.MustCompile(findRegexString)
	} else {
		r = regexp.MustCompile(createLogicalRegex(name, logic))
	}

	found := r.FindAllStringSubmatch(name, 1)
	fields := found[0][1:]

	collection := found[0][2]
	fn := func (args []reflect.Value) []reflect.Value {
		database := mongoConnectorInstance.GetDatabase()
		defer database.Session.Close()

		andOrList := make([]bson.M, 0)
		for i := 0; i < len(fields); i++ {
			andOrList  = append(andOrList , bson.M{
				fields[i] : args[i].Interface(),
			})
		}

		var selector bson.M
		switch logic {
		case "":
			selector = andOrList[0]
			break
		case "Or":
			selector = bson.M{
				"$or" : andOrList,
			}
			break
		case "And":
			selector = bson.M{
				"$and" : andOrList,
			}
			break
		}


		newValue := reflect.New(fieldFunctionValue.Type().Out(0))
		newValueInterface := newValue.Interface()
		
		err := database.C(collection).Find(selector).One(newValueInterface)

		secondValue := reflect.ValueOf(&err).Elem()

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
	} else {
		panic(fmt.Errorf("Name of the function to be implemented does not make sense : '%s'", name))
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
