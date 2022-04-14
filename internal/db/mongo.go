package db

import (
	"context"
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
	InternalError ErrorType = iota
	ValidationError
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrDrugNotFound      = errors.New("drug data not found")
	ErrNoSubmissionFound = errors.New("no airdrop submission found")
)

// collection names
const (
	drugs              string = "drugs"
	airdropSubmissions string = "airdropSubmissions"
	incidenceReports   string = "incidenceReports"
	users              string = "users"
	notifications      string = "notifications"
)

type Mongo struct {

	// a cursor to hrtnet database
	db *mongo.Database

	client *mongo.Client
}

// ConnectMongo connects to a single instance of Mongo server, with
// and embeds the database cursor in Mongo
func ConnectMongo(dsn string) (*Mongo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	// For a replica set, include the replica set name and a seedlist of the members in the URI string; e.g.
	// uri := "mongodb://mongodb0.example.com:27017,mongodb1.example.com:27017/?replicaSet=myRepl"
	// For a sharded cluster, connect to the mongos instances; e.g.
	// uri := "mongodb://mongos0.example.com:27017,mongos1.example.com:27017/"
	opt := options.Client().SetConnectTimeout(time.Second * 10).
		ApplyURI(dsn).SetAppName("hrtnet")
	client, err := mongo.Connect(ctx, opt)
	if err != nil {
		return nil, errors.Wrap(err, "error connecting to database")
	}

	// ping the connection to be sure the connection is properly configured
	if err = client.Ping(ctx, nil); err != nil {
		return nil, errors.Wrap(err, "error pinging mongo server")
	}

	database := client.Database("heartNet", &options.DatabaseOptions{})

	mongo := &Mongo{
		database,
		client,
	}

	mongo.runMigrations(ctx)
	return mongo, nil
}

// runMigrations creates necessary collections in a transaction,
// rollback and panics if any error is encountered.
// Note that creating collections within transactions is not
// available in Mongo 4.2 and earlier.
func (m *Mongo) runMigrations(ctx context.Context) {

	m.createDrugsCollection()
	m.createAirdropSubmissionCollection()
	m.createUsersCollection()
	m.createIncidenceReportCollection()
	m.createNotificationsCollection()

}

func (m *Mongo) createNotificationsCollection() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	jsonSchema := bson.M{
		"bsonType": "object",
		"required": []string{"message", "title", "uid"},
		"properties": bson.M{
			"uid": bson.M{
				"bsonType": "string",
			},
			"message": bson.M{
				"bsonType": "string",
			},
			"title": bson.M{
				"bsonType": "string",
			},
		},
	}
	validator := bson.M{
		"$jsonSchema": jsonSchema,
	}
	opts := options.CreateCollection().SetValidator(validator)

	// Potential error ignored because CreateCollection can only return
	// mongo.CommandError, which indicates that the collection is already existing
	if err := m.db.CreateCollection(ctx, notifications, opts); err != nil {

		logger.Logger.LogError("failed to create incidence report collection",
			"create incidence reports", err)
	}
}

func (m *Mongo) createIncidenceReportCollection() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	jsonSchema := bson.M{
		"bsonType": "object",
		"required": []string{"uid", "pharmacyName", "description", "pharmacyLocation", "evidenceImagesUrl", "receiptImageUrl"},
		"properties": bson.M{
			"uid": bson.M{
				"bsonType": "string",
			},
			"pharmacyName": bson.M{
				"bsonType": "string",
			},
			"description": bson.M{
				"bsonType": "string",
			},
			"pharmacyLocation": bson.M{
				"bsonType": "string",
			},
			"evidenceImagesUrl": bson.M{
				"bsonType": "array",
			},
			"receiptImageUrl": bson.M{
				"bsonType": "string",
			},
		},
	}

	validator := bson.M{
		"$jsonSchema": jsonSchema,
	}
	opts := options.CreateCollection().SetValidator(validator)

	// Potential error ignored because CreateCollection can only return
	// mongo.CommandError, which indicates that the collection is already existing
	if err := m.db.CreateCollection(ctx, incidenceReports, opts); err != nil {
		logger.Logger.LogError("failed to create incidence report collection",
			"create incidence reports", err)
	}
}

func (m *Mongo) createAirdropSubmissionCollection() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// user info is embedded rather than referenced because the airdrop submission doesn't make
	// sense without the user info.
	jsonSchema := bson.M{
		"bsonType": "object",
		"required": []string{"uid", "telegramUsername", "twitterUsername", "tweetLink", "submittedOn"},
		"properties": bson.M{
			"uid": bson.M{
				"bsonType": "string",
			},
			"telegramUsername": bson.M{
				"bsonType": "string",
			},
			"twitterUsername": bson.M{
				"bsonType": "string",
			},
			"tweetLink": bson.M{
				"bsonType": "string",
			},
			"submittedOn": bson.M{
				"bsonType": "date",
			},
		},
	}
	validator := bson.M{
		"$jsonSchema": jsonSchema,
	}
	opts := options.CreateCollection().SetValidator(validator)

	// Potential error ignored because CreateCollection can only return
	// mongo.CommandError, which indicates that the collection is already existing
	if err := m.db.CreateCollection(ctx, airdropSubmissions, opts); err != nil {
		logger.Logger.LogError("failed to create airdrop submission collection",
			"create airdrop submission collection", err)
	}
}

func (m *Mongo) createUsersCollection() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	jsonSchema := bson.M{
		"bsonType": "object",
		"required": []string{"uid"},
		"properties": bson.M{
			"uid": bson.M{
				"bsonType": "string",
			},
			"walletAddr": bson.M{
				"bsonType": "string",
			},
			"email": bson.M{
				"bsonType": "string",
				"pattern":  "@mongodb\\.com$",
			},
		},
	}
	validator := bson.M{
		"$jsonSchema": jsonSchema,
	}
	opts := options.CreateCollection().SetValidator(validator)

	if err := m.db.CreateCollection(ctx, users, opts); err != nil {
		logger.Logger.LogError("failed to create users collection",
			"create users collection", err)
	}
}

func (m *Mongo) createDrugsCollection() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	jsonSchema := bson.M{
		"bsonType": "object",
		"required": []string{"data", "drug", "validationOption"},
		"properties": bson.M{
			"data": bson.M{
				"bsonType":    "string",
				"description": "the validation data embedded on the drug container",
			},
			"drug": bson.M{
				"bsonType": "object",
				"required": []string{"name", "expiry", "batchNumber", "manufacturer", "manufactureDate"},
				"properties": bson.M{
					"expiry": bson.M{
						"bsonType": "date",
					},
					"manufacturer": bson.M{
						"bsonType": "string",
					},
					"manufactureDate": bson.M{
						"bsonType": "date",
					},
					"name": bson.M{
						"bsonType":    "string",
						"description": "name of the drug",
					},
				},
			},
			"validationOption": bson.M{
				"enum": []string{model.QrCode, model.RFID, model.ShortCode},
			},
		},
	}
	validator := bson.M{
		"$jsonSchema": jsonSchema,
	}
	opts := options.CreateCollection().SetValidator(validator)

	if err := m.db.CreateCollection(ctx, drugs, opts); err != nil {
		logger.Logger.LogError("failed to create drugs collection",
			"create drugs collection", err)
	}
	if err := m.seedDrugs(); err != nil {
		logger.Logger.LogError("failed to seed drugs collection",
			"seed drugs", err)
	}
}

func (m *Mongo) seedDrugs() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	docs := []interface{}{
		bson.D{
			{"drug", model.SampleDrug1},
			{"data", model.SampleDrug1.String()},
			{"validationOption", model.QrCode},
		},
		bson.D{
			{"drug", model.SampleDrug2},
			{"data", model.SampleDrug2.String()},
			{"validationOption", model.QrCode},
		},
		bson.D{
			{"drug", model.SampleDrug3},
			{"data", model.SampleDrug3.String()},
			{"validationOption", model.QrCode},
		},
		bson.D{
			{"drug", model.SampleDrug4},
			{"data", "12345678"},
			{"validationOption", model.ShortCode},
		},
		bson.D{
			{"drug", model.SampleDrug5},
			{"data", "12QWERTY"},
			{"validationOption", model.ShortCode},
		},
	}
	opts := options.InsertMany().SetOrdered(false)
	_, err := m.db.Collection(drugs).InsertMany(ctx, docs, opts)
	if err != nil {
		return errors.Wrap(err, "failed to seed drugs")
	}
	return nil
}

func (m *Mongo) ValidateQrText(value string) (*model.Drug, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var drug model.DBDrug
	err := m.db.Collection(drugs).FindOne(ctx, bson.D{{"data", value}, {"validationOption", model.QrCode}}).Decode(&drug)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrDrugNotFound
		}
		return nil, errors.Wrap(err, "validate qr text: failed to query drug")
	}
	return &drug.Drug, nil
}

func (m *Mongo) ValidateShortCode(value string) (*model.Drug, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var drug model.DBDrug
	err := m.db.Collection(drugs).FindOne(ctx, bson.D{{"data", value}, {"validationOption", model.ShortCode}}).Decode(&drug)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrDrugNotFound
		}
		return nil, errors.Wrap(err, "validate short code: failed to query drug")
	}
	return &drug.Drug, nil
}

func (m *Mongo) ValidateRFIDText(value string) (*model.Drug, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var drug model.DBDrug
	err := m.db.Collection(drugs).FindOne(ctx, bson.D{{"data", value}, {"validationOption", model.RFID}}).Decode(&drug)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrDrugNotFound
		}
		return nil, errors.Wrap(err, "validate rfid: failed to query drug")
	}
	return &drug.Drug, nil
}

func (m *Mongo) FetchAllAirdropSubmissions() (*[]model.AirdropSubmission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	curs, err := m.db.Collection(airdropSubmissions).Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	var submissions []model.AirdropSubmission

	if err := curs.All(ctx, &submissions); err != nil {
		return nil, errors.Wrap(err, "fetch all airdrop submissions: failed to decode find result into slice")
	}

	return &submissions, err

}

func (m *Mongo) InsertMultipleDrugs(values *[]model.DBDrug, option model.ValidationOption) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	var docs []interface{}

	for _, qr := range *values {
		doc, err := bson.Marshal(qr)
		if err != nil {
			logger.Logger.LogError(
				"error marshalling dbDrug to Bson", "insert multiple drug", err)
			continue
		}
		docs = append(docs, doc)
	}
	opts := options.InsertMany().SetOrdered(false)
	_, err := m.db.Collection(drugs).InsertMany(ctx, docs, opts)
	if err != nil {
		return errors.Wrap(err, "failed to insert multiple drugs")
	}

	return nil
}

func (m *Mongo) FetchRandomQRCode() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// query db
	curs, err := m.db.Collection(drugs).
		Aggregate(ctx, mongo.Pipeline{
			bson.D{{"$match", bson.D{{"validationOption", model.QrCode}}}},
			bson.D{{"$sample", bson.D{{"size", 1}}}},
		})
	if err != nil {
		return "", errors.Wrap(err, "failed to fetch random qr")
	}

	// scan result into dbDrugs
	var dbDrugs []model.DBDrug
	if err := curs.All(ctx, &dbDrugs); err != nil {
		return "", errors.Wrap(err, "failed to decode random qr code result")
	}

	if len(dbDrugs) < 1 {
		return "", errors.New("failed to fetch random qr code; no data returned")
	}
	return dbDrugs[0].ValidationData, nil
}

func (m *Mongo) FetchAirdropSubmissionByUserID(userId string) (*model.AirdropSubmission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	var submission model.AirdropSubmission
	err := m.db.Collection(airdropSubmissions).FindOne(ctx, bson.D{{"uid", userId}}).Decode(&submission)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNoSubmissionFound
		}
		return nil, errors.Wrap(err, "failed to fetch user airdrop submission details")
	}

	return &submission, nil
}

func (m *Mongo) GenerateNewUserID() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userId := onetimecode.NewAlphanumericalCode(
		onetimecode.WithAlphaNumericCode(),
		onetimecode.WithMax(6),
		onetimecode.WithoutDashes(),
	).Code()

	var user model.User
	if err := m.db.Collection(users).
		FindOne(ctx, bson.D{{"uid", userId}}).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			// that implies the uid is unassigned and can then be assigned
		} else {
			return "", errors.Wrap(err, "fetch user id: failed to decode user")
		}
	}

	// if user is found, regenerate a new user id
	if len(user.UID) != 0 {
		logger.Logger.LogWarn("UID collision",
			"generate new user id", errors.New("already existing user id returned for new user"))
		m.GenerateNewUserID()
	}

	return userId, nil
}

func (m *Mongo) InsertAirdropSubmission(submission *model.AirdropSubmission) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := m.db.Collection(airdropSubmissions).InsertOne(ctx, submission)
	if err != nil {
		return errors.Wrap(err, "failed to insert airdrop submission into db")
	}
	return nil
}

func (m *Mongo) Disconnect() error {
	return m.db.Client().Disconnect(context.TODO())
}

func (m *Mongo) SaveNotification(notification *model.Notification) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := m.db.Collection(notifications).InsertOne(ctx, notification)
	if err != nil {
		return errors.Wrap(err, "failed to insert notification into db")
	}
	return nil
}

func (m *Mongo) ReadNotification(notificationId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := m.db.Collection(notifications).DeleteOne(ctx, bson.D{{"id", notificationId}})
	if err != nil {
		return errors.Wrap(err, "failed to read notification")
	}
	return nil
}

func (m *Mongo) FetchAllUnreadNotifications(forUserId string) (*[]model.Notification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	curs, err := m.db.Collection(notifications).Find(ctx, bson.D{{"uid", forUserId}})
	if err != nil {
		return nil, err
	}

	var notifications []model.Notification

	if err := curs.All(ctx, &notifications); err != nil {
		return nil, errors.Wrap(err, "fetch all unread notifications: failed to decode find result into slice")
	}

	return &notifications, err
}

func (m *Mongo) IsValidUser(uid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var user model.User
	if err := m.db.Collection(users).
		FindOne(ctx, bson.D{{"uid", uid}}).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrUserNotFound
		}
		return errors.Wrap(err, "is valid user: failed to decode result")
	}
	if user.UID == "" {
		return ErrUserNotFound
	}
	return nil
}

func (m *Mongo) FetchUserInfo(uid string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var user model.User
	if err := m.db.Collection(users).
		FindOne(ctx, bson.D{{"uid", uid}}).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrUserNotFound
		}
		return nil, errors.Wrap(err, "is valid user: failed to decode result")
	}
	return &user, nil
}

func (m *Mongo) SubmitIncidenceReport(report *model.IncidenceReport) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := m.db.Collection(airdropSubmissions).InsertOne(ctx, report)
	if err != nil {
		return errors.Wrap(err, "failed to insert incidence report into db")
	}
	return nil
}
