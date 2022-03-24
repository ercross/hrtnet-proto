package db

import (
	"context"
	"fmt"
	"github.com/Hrtnet/social-activities/internal/logger"
	"github.com/Hrtnet/social-activities/internal/model"
	"github.com/jakoubek/onetimecode"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type ErrorType int

const (
	DBError ErrorType = iota
	ValidationError
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrDrugNotFound = errors.New("drug data not found")
)

const (
	tasksCollectionName       string = "tasks"
	qrCodeDataCollectionName  string = "qrCodes"
	taskReportsCollectionName string = "taskReport"
	userIdCollectionName      string = "userId"
)

type Mongo struct {
	db *mongo.Database
}

func Connect(dsn string) (*Mongo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	fmt.Println("this is the dsn as gotten: ", dsn)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dsn))
	if err != nil {
		return nil, errors.Wrap(err, "error connecting to database")
	}

	if err = client.Ping(context.TODO(), nil); err != nil {
		return nil, errors.Wrap(err, "error pinging mongo server")
	}

	database := client.Database("alpha", &options.DatabaseOptions{
		ReadConcern:    nil,
		WriteConcern:   nil,
		ReadPreference: nil,
		Registry:       nil,
	})

	mongo := &Mongo{
		database,
	}

	curs, err := mongo.db.ListCollections(ctx, bson.D{})
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving all collections")
	}

	// *
	for curs.Next(ctx) {

	}

	mongo.initializeQrCodeDataCollection()
	mongo.initializeTasksCollection()
	mongo.initializeTaskReportCollection()
	mongo.initializeUserIDCollection()

	return mongo, nil
}

func (m *Mongo) initializeTaskReportCollection() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	jsonSchema := bson.M{
		"bsonType": "object",
		"required": []string{"user_id", "submitted_on",
			"telegram_username", "twitter_username", "youtube_username"},
		"properties": bson.M{
			"user_id": bson.M{
				"bsonType": "string",
			},
			"submitted_on": bson.M{
				"bsonType": "date",
			},
			"telegram_username": bson.M{
				"bsonType": "string",
			},
			"twitter_username": bson.M{
				"bsonType": "string",
			},
			"youtube_username": bson.M{
				"bsonType": "string",
			},
		},
	}
	validator := bson.M{
		"$jsonSchema": jsonSchema,
	}
	opts := options.CreateCollection().SetValidator(validator)

	err := m.db.CreateCollection(ctx, taskReportsCollectionName, opts)
	if err != nil {
		logger.Logger.LogFatal("error creating tasks report collection",
			"creating tasks report collection", err)
	}
}

func (m *Mongo) initializeUserIDCollection() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	jsonSchema := bson.M{
		"bsonType": "object",
		"required": []string{"userId", "createdAt"},
		"properties": bson.M{
			"userId": bson.M{
				"bsonType": "string",
			},
			"createdAt": bson.M{
				"bsonType": "date",
			},
		},
	}
	validator := bson.M{
		"$jsonSchema": jsonSchema,
	}
	opts := options.CreateCollection().SetValidator(validator)

	err := m.db.CreateCollection(ctx, tasksCollectionName, opts)
	if err != nil {
		logger.Logger.LogFatal("error creating user id collection",
			"creating user id collection", err)
	}
}

// initializeTasksCollection saves all tasks into a tasks collection on the Mongo server.
// initializeTasksCollection should be called only if tasks collection is empty
func (m *Mongo) initializeTasksCollection() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	jsonSchema := bson.M{
		"bsonType": "object",
		"required": []string{"task"},
		"properties": bson.M{
			"task": bson.M{
				"bsonType":    "string",
				"description": "actual task value",
			},
		},
	}
	validator := bson.M{
		"$jsonSchema": jsonSchema,
	}
	opts := options.CreateCollection().SetValidator(validator)

	err := m.db.CreateCollection(ctx, tasksCollectionName, opts)
	if err != nil {
		logger.Logger.LogFatal("error creating tasks collection",
			"creating tasks collection", err)
	}
}

// initializeQrCodeDataCollection initializes the qr-codes collection with sample qr-codes.
// initializeQrCodeDataCollection should be called only if qr-codes collection is empty.
func (m *Mongo) initializeQrCodeDataCollection() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	jsonSchema := bson.M{
		"bsonType": "object",
		"required": []string{"data"},
		"properties": bson.M{
			"data": bson.M{
				"bsonType": "binData",
			},
		},
	}
	validator := bson.M{
		"$jsonSchema": jsonSchema,
	}
	opts := options.CreateCollection().SetValidator(validator)

	err := m.db.CreateCollection(ctx, qrCodeDataCollectionName, opts)
	if err != nil {
		logger.Logger.LogFatal("error creating qr code collection",
			"creating qr code collection", err)
	}
}

func (m *Mongo) ValidateQrText(value string) (*model.Drug, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	drug := model.DrugFromString(value)
	if drug == nil {
		return nil, errors.New("invalid qr data")
	}

	options.FindOne().SetSort(bson.D{{"drug_name", drug.Name}})
	res := m.db.Collection(qrCodeDataCollectionName).FindOne(ctx, bson.D{{"_id", drug.Name}})
	if res.Err() != nil {
		return nil, res.Err()
	}
	var data2 model.Drug
	return nil, res.Decode(&data2)
}

func (m *Mongo) FetchAllTaskReports() (*[]model.TasksReport, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	curs, err := m.db.Collection(taskReportsCollectionName).Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	var reports []model.TasksReport

	for curs.Next(ctx) {
		var report model.TasksReport
		if err := curs.Decode(&report); err != nil {

			continue
		}
		reports = append(reports, report)
	}
	return &reports, err

}

func (m *Mongo) InsertMultipleDrugs(drugs *[]model.Drug, option model.ValidationOption) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	var docs []interface{}
	for _, qr := range *drugs {
		doc, err := bson.Marshal(qr)
		if err != nil {
			logger.Logger.LogError(
				"error marshalling QrCodeData to Bson", "insert multiple qr data", err)
			continue
		}
		docs = append(docs, doc)
	}
	opts := options.InsertMany().SetOrdered(false)
	_, err := m.db.Collection(taskReportsCollectionName).InsertMany(ctx, docs, opts)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mongo) FetchRandomQRCode() (*model.Drug, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	curs, err := m.db.Collection(qrCodeDataCollectionName).
		Aggregate(ctx, mongo.Pipeline{bson.D{{"$sample", bson.D{{"size", 5}}}}})

	if err != nil {
		return nil, errors.Wrap(err, "error extracting running aggregate query")
	}

	var qrs []model.Drug
	for curs.Next(ctx) {
		var qr model.Drug
		if err := curs.Decode(&qr); err != nil {
			return nil, err
		}
		qrs = append(qrs, qr)
	}

	if len(qrs) > 1 {
		return &qrs[0], nil
	}
	return nil, errors.New("unable to retrieve random Qr code")
}

func (m *Mongo) FetchTaskReportByUserID(userId string) (*model.TasksReport, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	res := m.db.Collection(taskReportsCollectionName).FindOne(ctx, bson.D{{"user_id", userId}})
	if res.Err() != nil {
		return nil, errors.Wrap(res.Err(), "error querying database")
	}
	var taskReport model.TasksReport
	if err := res.Decode(taskReport); err != nil {
		return nil, errors.Wrap(err, "fetch task by user id: error decoding bson")
	}
	return &taskReport, nil
}

func (m *Mongo) FetchAllTasks() ([]string, error) {
	return nil, nil
}

func (m *Mongo) FetchUserID() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userId := onetimecode.NewAlphanumericalCode(
		onetimecode.WithAlphaNumericCode(),
		onetimecode.WithMax(6),
		onetimecode.WithoutDashes(),
	).Code()

	opts := options.FindOne().SetSort(bson.D{{"user_id", 1}})
	var user struct {
		Id string `bson:"id"`
	}
	if err := m.db.Collection(userIdCollectionName).
		FindOne(ctx, bson.D{{"user_id", userId}}, opts).Decode(&user); err != nil {
		return "", errors.Wrap(err, "unable to decode db data")
	}
	if len(user.Id) != 0 {
		m.FetchUserID()
	} else {
		return userId, nil
	}
	return userId, nil
}

func (m *Mongo) InsertDrug(drug model.Drug, option model.ValidationOption) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	raw, err := bson.Marshal(drug)
	if err != nil {
		return errors.Wrap(err, "insert qr data: error encoding data")
	}

	_, err = m.db.Collection(qrCodeDataCollectionName).InsertOne(ctx, raw)
	if err != nil {
		return errors.Wrap(err, "error inserting qr data into database")
	}

	return nil
}

func (m *Mongo) CreateTaskReport(report *model.TasksReport) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	data, err := bson.Marshal(report)
	if err != nil {
		return errors.Wrap(err, "insertTaskReport: error marshaling task report")
	}
	_, err = m.db.Collection(taskReportsCollectionName).InsertOne(ctx, data)
	if err != nil {
		return errors.Wrap(err, "error inserting task report into db")
	}
	return nil
}

func (m *Mongo) Disconnect() error {
	return m.db.Client().Disconnect(context.TODO())
}

func (m *Mongo) GetDrugByBatchNumber(batchNumber, manufacturer string) (*model.Drug, error) {
	return nil, nil
}
