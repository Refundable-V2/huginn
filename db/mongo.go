package db

import (
	"context"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"log"
	"os"
	"time"
)

// The name of the collection in which the Teacher data is stored in
const TeacherCollection = "Teacher"

// The name of the collection in which the Application data is stored in
const ApplicationCollection = "Collection"

// Containing data used for the mongo db connection
type MongoDatabaseConnector struct {
	// the name of the database in the mongo db server
	database string
	// the client used in this connection
	client *mongo.Client
	// the created context of the client
	context context.Context
	// CancelFunc of the context
	closer context.CancelFunc
}

// Connects the MongoDatabaseConnector with the given MongoDB server
// returns whether this was successful
func (m *MongoDatabaseConnector) Connect() bool {
	uri, db, ok := resolveURI()
	if ok {
		client, err := mongo.NewClient(options.Client().ApplyURI(uri))
		if err != nil {
			log.Println(err)
			return false
		}
		ctx, cf := context.WithTimeout(context.Background(), 10*time.Minute)
		err = client.Connect(ctx)
		if err != nil {
			log.Println(err)
			cf()
			return false
		}
		m.client = client
		m.database = db
		m.context = ctx
		m.closer = cf
		return true
	}
	return false
}

// Closes the Connection to the MongoDB
// returns whether this operation was successful
func (m MongoDatabaseConnector) Close() (ok bool) {
	err := m.client.Disconnect(m.context)
	m.closer()
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

// Creates a new application in the system
func (m MongoDatabaseConnector) CreateApplication(application Application) bool {
	application.uuid = uuid.New().String()
	collection := m.client.Database(m.database).Collection(ApplicationCollection)
	insert, err := collection.InsertOne(m.context, application)
	if err != nil {
		log.Println(err)
		return false
	}
	log.Println("Inserted a new application with the UUID: ", application.uuid,
		"; the Title: ", application.name, "; under the ID: ", insert.InsertedID)
	return true
}

// Returns a specific application described by its uuid
func (m MongoDatabaseConnector) GetApplication(uuid string) (application Application) {
	collection := m.client.Database(m.database).Collection(ApplicationCollection)
	if err := collection.FindOne(m.context, bson.M{"uuid": uuid}).Decode(&application); err != nil {
		log.Println(err)
		return
	}
	return application
}

// Returns all active applications in the system
func (m MongoDatabaseConnector) GetActiveApplications() (applications []Application) {
	filter := bson.M{
		"$or": []bson.M{
			{"$and": []bson.M{
				{"kind": Training},
				{"progress": bson.M{"$in": []int{TRejected, TInProcess, TConfirmed, TRunning, TCostsPending, TCostsInProcess}}},
			}},
			{"$and": []bson.M{
				{"kind": SchoolEvent},
				{"progress": bson.M{"$in": []int{SERejected, SEInSubmission, SEInProcess, SEConfirmed, SERunning, SECostsPending, SECostsInProcess}}},
			}},
		},
	}
	collection :=  m.client.Database(m.database).Collection(ApplicationCollection)
	cursor, err := collection.Find(m.context, filter)
	if err != nil {
		log.Println(err)
		return
	}
	err = cursor.All(m.context, &applications)
	if err != nil {
		log.Println(err)
		return
	}
	return applications
}

func (m MongoDatabaseConnector) UpdateApplication() (ok bool) {
	return false
}

func (m MongoDatabaseConnector) DeleteApplication(uuid string) (ok bool) {
	return false
}

func (m MongoDatabaseConnector) CreateTeacher() (ok bool) {
	return false
}

func (m MongoDatabaseConnector) GetTeacher() (teacher Teacher) {
	return Teacher{}
}

func (m MongoDatabaseConnector) UpdateTeacher() (ok bool) {
	return false
}

func (m MongoDatabaseConnector) DeleteTeacher() (ok bool) {
	return false
}

// Constructs the URI out of the given information of the docker secrets
// returns the constructed URI, the database name, and whether the operation was successful
// if it was not successful the URI and the database name are empty strings
func resolveURI() (URI string, database string, ok bool) {
	database = os.Getenv("MONGO_DATABASE")
	usernameFilePath := os.Getenv("MONGO_USERNAME_FILE")
	passwordFilePath := os.Getenv("MONGO_PASSWORD_FILE")
	username, err := ioutil.ReadFile(usernameFilePath)
	if err != nil {
		log.Println(err)
		return "", "", false
	}
	password, err := ioutil.ReadFile(passwordFilePath)
	if err != nil {
		log.Println(err)
		return "", "", false
	}
	return "mongodb://" + string(username) + ":" + string(password) + "@" + "mongo:27017", database, true
}
