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
	taskReports map[string]model.AirdropSubmission
	tasks       []string

	// maps user id to wallet address
	userIds map[string]string

	drugs map[model.Drug]model.ValidationOption

	// maps short code to drug
	shortCodes map[string]*model.Drug

	// maps rfidText to drug
	rfidText map[string]*model.Drug

	// maps qrCode string to drug
	qrCodes map[string]*model.Drug

	incidenceReports *[]model.IncidenceReport

	// maps userId to slice of notifications
	notifications map[string]*[]model.Notification
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
	c.taskReports = make(map[string]model.AirdropSubmission)
	c.rfidText = make(map[string]*model.Drug)

	// initialize short codes
	c.shortCodes = make(map[string]*model.Drug)
	c.shortCodes["12345678"] = &model.SampleDrug2

	c.incidenceReports = &[]model.IncidenceReport{}
	c.userIds = make(map[string]string)
	c.notifications = make(map[string]*[]model.Notification)

	// initialize qr codes
	c.qrCodes = make(map[string]*model.Drug)
	c.qrCodes[model.SampleDrug2.String()] = &model.SampleDrug2
	c.qrCodes[model.SampleDrug1.String()] = &model.SampleDrug1
	return &c
}

func (c *Cache) Disconnect() error {
	return nil
}

func (c *Cache) SaveNotification(notification *model.Notification) error {
	if _, ok := c.notifications[notification.UserID]; !ok {
		c.notifications[notification.UserID] = &[]model.Notification{*notification}
		return nil
	}
	list := c.notifications[notification.UserID]
	*list = append(*list, *notification)
	return nil
}

func (c *Cache) ReadNotification(userId, notificationId string) error {
	ptr, _ := c.FetchAllUnreadNotifications(userId)
	list := *ptr

	for idx, notification := range list {
		if notification.ID == notificationId {

			// slice out the notification
			list = append(list[:idx], list[idx+1:]...)
		}
	}
	c.notifications[userId] = &list
	return nil
}

func (c *Cache) FetchAllUnreadNotifications(forUserId string) (*[]model.Notification, error) {
	return c.notifications[forUserId], nil
}

func (c *Cache) FetchAllTasks() ([]string, error) {
	return c.tasks, nil
}

func (c *Cache) FetchAllAirdropSubmissions() (*[]model.AirdropSubmission, error) {
	var tr []model.AirdropSubmission
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

	// check that user id isn't already existing
	_, ok := c.userIds[userId]
	if ok {
		c.FetchUserID()
	} else {

		// save new user id into wallet
		c.userIds[userId] = "0xc0ffee254729296a45a3885639AC7E10F9d54979"
		return userId, nil
	}
	return c.FetchUserID()
}

func (c *Cache) IsValidUser(id string) error {
	_, ok := c.userIds[id]
	if !ok {
		return ErrUserNotFound
	}
	return nil
}

func (c *Cache) FetchRandomQRCode() (string, error) {
	rand.Seed(time.Now().UnixNano())
	drugs := []model.Drug{model.SampleDrug1, model.SampleDrug2, model.SampleDrug3}
	random := rand.Intn(len(drugs) - 1)
	return drugs[random].String(), nil
}

func (c *Cache) FetchAirdropSubmissionByUserID(userId string) (*model.AirdropSubmission, error) {
	report, ok := c.taskReports[userId]
	if ok {
		return &report, nil
	} else {
		return nil, errors.New("invalid user id supplied")
	}
}

func (c *Cache) InsertAirdropSubmission(report *model.AirdropSubmission) error {
	c.taskReports[report.UserID] = *report
	return nil
}

func (c *Cache) ValidateQrText(value string) (*model.Drug, error) {
	drug, ok := c.qrCodes[value]
	if !ok {
		return nil, ErrDrugNotFound
	}
	return drug, nil
}

func (c *Cache) ValidateShortCode(value string) (*model.Drug, error) {
	drug, ok := c.shortCodes[value]
	if !ok {
		return nil, ErrDrugNotFound
	}
	return drug, nil
}

func (c *Cache) ValidateRFIDText(value string) (*model.Drug, error) {
	drug, ok := c.shortCodes[value]
	if !ok {
		return nil, ErrDrugNotFound
	}
	return drug, nil
}

func (c *Cache) SubmitIncidenceReport(report *model.IncidenceReport) error {
	*c.incidenceReports = append(*c.incidenceReports, *report)
	return nil
}

func (c *Cache) FetchWalletAddress(forUserId string) (string, error) {
	addr, ok := c.userIds[forUserId]
	if !ok {
		return "", ErrUserNotFound
	}
	return addr, nil
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
