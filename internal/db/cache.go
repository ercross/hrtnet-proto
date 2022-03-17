package db

import (
	"errors"
	"github.com/Hrtnet/social-activities/internal/model"
	"github.com/jakoubek/onetimecode"
	"math/rand"
	"time"
)

type Cache struct {
	// maps userId to task report
	taskReports map[string]model.TasksReport
	tasks       []string

	userIds map[string]string

	drugs map[model.Drug]model.ValidationOption

	// maps short code to drug
	shortCodes map[string]*model.Drug

	// maps rfidText to drug
	rfidText map[string]*model.Drug
}

var sampleDrug1 = model.Drug{
	Manufacturer:   "Heart Pharm",
	Name:           "Chloramphenicol",
	ManufacturedOn: time.Now(),
	Expiry:         time.Now().Add(time.Hour * 72),
	BatchNumber:    "1234BRQ",
}

var sampleDrug2 = model.Drug{
	Manufacturer:   "Heart Pharm",
	Name:           "Loxagyl",
	ManufacturedOn: time.Now(),
	Expiry:         time.Now().Add(time.Hour * 72),
	BatchNumber:    "1234BRQ",
}

func InitCache() *Cache {
	var c Cache
	c.tasks = []string{
		"Join our Telegram community",
		"Follow us on Twitter",
		"Retweet with the hashtag #heartnet, #hrt",
		"Join our Discord server",
		"Using your User ID as your referral ID, refer 5 of your friends to join our telegram community",
	}
	c.taskReports = make(map[string]model.TasksReport)

	c.userIds = make(map[string]string)
	return &c
}

func (c *Cache) Disconnect() error {
	return nil
}

func (c *Cache) FetchAllTasks() ([]string, error) {
	return c.tasks, nil
}

func (c *Cache) FetchAllTaskReports() (*[]model.TasksReport, error) {
	var tr []model.TasksReport
	for _, value := range c.taskReports {
		tr = append(tr, value)
	}
	return &tr, nil
}

func (c *Cache) FetchUserID() (string, error) {
	userId := onetimecode.NewAlphanumericalCode(
		onetimecode.WithAlphaNumericCode(),
		onetimecode.WithMax(6),
		onetimecode.WithoutDashes(),
	).Code()
	_, ok := c.userIds[userId]
	if ok {
		c.FetchUserID()
	} else {
		return userId, nil
	}
	return c.FetchUserID()
}

func (c *Cache) FetchRandomQRCode() (string, error) {
	rand.Seed(time.Now().UnixNano())
	drugs := []model.Drug{sampleDrug1, sampleDrug2, sampleDrug1}
	random := rand.Intn(len(drugs) - 1)
	return drugs[random].String(), nil
}

func (c *Cache) FetchTaskReportByUserID(userId string) (*model.TasksReport, error) {
	report, ok := c.taskReports[userId]
	if ok {
		return &report, nil
	} else {
		return nil, errors.New("invalid user id supplied")
	}
}

func (c *Cache) CreateTaskReport(report *model.TasksReport) error {
	c.taskReports[report.UserID] = *report
	return nil
}

func (c *Cache) ValidateQrText(value string) (*model.Drug, error) {
	drug := model.DrugFromString(value)
	if drug == nil {
		return nil, errors.New("invalid QR data")
	}
	vOption, ok := c.drugs[*drug]

	if ok && vOption == model.QrCode {
		return drug, nil
	}

	return nil, errors.New("drug data not found in HeartNet repository")
}

func (c *Cache) ValidateShortCode(value string) (*model.Drug, error) {
	drug, ok := c.shortCodes[value]
	if !ok {
		return nil, errors.New("invalid tracking code")
	}
	return drug, nil
}

func (c *Cache) ValidateRFIDText(value string) (*model.Drug, error) {
	drug, ok := c.shortCodes[value]
	if !ok {
		return nil, errors.New("invalid tracking code")
	}
	return drug, nil
}

func (c *Cache) FetchDrugByBatchNumber(batchNumber, manufacturer string) (*model.Drug, error) {
	return nil, nil
}

func (c *Cache) InsertDrug(model.Drug, model.ValidationOption) error {
	return nil
}

func (c *Cache) InsertMultipleDrugs(*[]model.Drug, model.ValidationOption) error {
	return nil
}
