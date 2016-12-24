package mongoconnector

import (
	"gopkg.in/mgo.v2"
	"crypto/tls"
	"strings"
	"github.com/pkg/errors"
	"net"
)


func NewMongoSession(addrs string) (*mgo.Session, error) {
	tlsConfig := &tls.Config{}
	tlsConfig.InsecureSkipVerify = true

	dialInfo, err := mgo.ParseURL(addrs)
	if !strings.Contains(dialInfo.Addrs[0], "localhost") && !strings.Contains(dialInfo.Addrs[0], "127.0.0.1") {
		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
			return conn, err
		}
	}
	session, err := mgo.DialWithInfo(dialInfo)
	return session, errors.Wrap(err, "Could not dial mongo")
}

var mongoConnectorInstance *MongoConnector

type MongoConnector struct {
	mongoSession *mgo.Session

	mainDatabase string
	userCollection string
	messageCollection string
}

func NewMongoConnector(addrs string, database string, userCollection string, messageCollection string) (*MongoConnector, error) {
	session, err := NewMongoSession(addrs)
	if err != nil {
		return nil, err
	}
	session.SetMode(mgo.Monotonic, true)
	return &MongoConnector{
		mongoSession: session,
		mainDatabase: database,
		userCollection: userCollection,
		messageCollection: messageCollection,
	}, nil
}

func (selfPtr *MongoConnector) GetSession() *mgo.Session {
	return selfPtr.mongoSession.Copy()
}

func GetMongoConnectorSingleton() *MongoConnector {
	return mongoConnectorInstance
}

func InitializeMongoConnectorSingleton(addrs string, database string, userCollection string, messageCollection string) {
	if mongoConnectorInstance == nil {
		newMongoConnectorInstance, err := NewMongoConnector(addrs, database, userCollection, messageCollection)
		if err != nil {
			panic(err)
		}
		mongoConnectorInstance = newMongoConnectorInstance
	}
}
