package mongoconnector

import (
	"reflect"

	"regexp"
	"strings"
	"gopkg.in/mgo.v2/bson"
	"fmt"
	"unicode"
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
	regex := "FindBy" + wordMatch
	for i := 0; i < count; i++ {
		regex += logic + wordMatch
	}
	regex += "From(.*)"
	return regex
}


func CamelToSnake(s string) string {
	var result string
	var words []string
	var lastPos int
	rs := []rune(s)

	for i := 0; i < len(rs); i++ {
		if i > 0 && unicode.IsUpper(rs[i]) {
			words = append(words, s[lastPos:i])
			lastPos = i
		}
	}

	// append the last word
	if s[lastPos:] != "" {
		words = append(words, s[lastPos:])
	}

	for k, word := range words {
		if k > 0 {
			result += "_"
		}

		result += strings.ToLower(word)
	}

	return result
}



func generateFindFunction(fieldFunctionValue reflect.Value, name string) reflect.Value {

	var logic string
	if strings.Contains(name, "Or") {
		logic = "Or"
	}

	if strings.Contains(name, "And") {
		logic = "And"
	}

	var r *regexp.Regexp
	if logic == "" {
		r = regexp.MustCompile(findRegexString)
	} else {
		r = regexp.MustCompile(createLogicalRegex(name, logic))
	}


	found := r.FindAllStringSubmatch(name, 1)


	fields := found[0][1:len(found[0]) - 1]

	collection := CamelToSnake(found[0][len(found[0]) - 1])

	fn := func (args []reflect.Value) []reflect.Value {
		database := mongoConnectorInstance.GetDatabase()
		defer database.Session.Close()

		//todo We can read the field from the return type or save type
		//todo read the tag and decide save from there
		andOrList := make([]bson.M, 0)
		for i := 0; i < len(fields); i++ {
			andOrList  = append(andOrList , bson.M{
				CamelToSnake(fields[i]) : args[i].Interface(),
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

		firstReturnType := fieldFunctionValue.Type().Out(0)
		newValue := reflect.New(firstReturnType)
		newValueInterface := newValue.Interface()

		var err error
		if firstReturnType.Kind() == reflect.Slice {
			err = database.C(collection).Find(selector).All(newValueInterface)
		} else {
			err = database.C(collection).Find(selector).One(newValueInterface)
		}


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
		resultFunction := resolveAndCreateFunc(fieldValue, value.Type().Field(i).Name)
		fieldValue.Set(resultFunction)
	}

	return value.Interface()
}
