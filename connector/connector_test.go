package mongoconnector

import (
	dockertest "gopkg.in/ory-am/dockertest.v2"
	"testing"
	//"fmt"
	"k8s.io/kops/_vendor/github.com/docker/docker/pkg/testutil/assert"
	"fmt"
	"strconv"
)

type TestDocument struct {
	Thing string `bson:"thing"`
	Thing2 string `bson:"thing2"`
}

type Dao struct {
	FindByThingFromTestC func (thing string) (TestDocument, error)
	FindByThingAndThing2FromTestC func (thing string, thing2 string) (TestDocument, error)
	FindByThingOrThing2FromTestC func (thing string, thing2 string) ([]TestDocument, error)
}

func TestImplement(t *testing.T) {
	c, ip, port, err := dockertest.SetupMongoContainer()
	fmt.Println(port)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	defer c.Kill()

	InitializeMongoConnectorSingleton(ip + ":" + strconv.Itoa(port), "test_db")
	mc := GetMongoConnectorSingleton()

	session := mc.GetSession()
	defer session.Close()

	//Load data
	err = session.DB("test_db").C("test_c").Insert(TestDocument{
		Thing: "thing",
		Thing2: "thing2",
	})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	err = session.DB("test_db").C("test_c").Insert(TestDocument{
		Thing: "thing0",
		Thing2: "thing20",
	})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	////


	implementedDao := Implement(&Dao{}).(Dao)

	testDocument, err := implementedDao.FindByThingFromTestC("thing")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assert.Equal(t, "thing", testDocument.Thing)

	testDocument, err = implementedDao.FindByThingAndThing2FromTestC("thing", "thing2")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assert.Equal(t, "thing", testDocument.Thing)
	assert.Equal(t, "thing2", testDocument.Thing2)



	testDocuments, err := implementedDao.FindByThingOrThing2FromTestC("thing", "thing20")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assert.Equal(t, 2, len(testDocuments))
	assert.Equal(t, "thing", testDocuments[0].Thing)
	assert.Equal(t, "thing2", testDocuments[0].Thing2)
	assert.Equal(t, "thing0", testDocuments[1].Thing)
	assert.Equal(t, "thing20", testDocuments[1].Thing2)
}