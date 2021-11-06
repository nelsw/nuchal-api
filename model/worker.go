package model

import (
	"gorm.io/gorm"
	"nuchal-api/db"
	"time"
)

type JobType int

const (
	InitAllCBProducts JobType = iota
	InitAllCBCurrencies
)

type Job struct {
	gorm.Model
	UserID  uint      `gorm:"user_id"`
	JobType JobType   `gorm:"job_type"`
	alpha   time.Time `gorm:"alpha"`
	omega   time.Time `gorm:"omega"`
}

func init() {
	db.Migrate(&Job{})
}

func (j *Job) Perform() error {
	switch j.JobType {
	case InitAllCBProducts:
		return j.initAllCBProducts()
	case InitAllCBCurrencies:
		return j.initAllCBCurrencies()
	}
	return nil
}

func (j *Job) initAllCBProducts() error {

	if !j.isTimeToDoJob() {
		return nil
	}

	u := FindUserByID(j.UserID)

	products, err := u.Client().GetProducts()
	if err != nil {
		return err
	}

	var result []Product
	for _, product := range products {
		result = append(result, Product{Product: product})
	}

	db.Resolve().Create(&result)
	j.omega = time.Now()
	db.Resolve().Create(j)

	return nil
}

func (j *Job) initAllCBCurrencies() error {

	if !j.isTimeToDoJob() {
		return nil
	}

	u := FindUserByID(j.UserID)

	j.alpha = time.Now()

	currencies, err := u.Client().GetCurrencies()
	if err != nil {
		return err
	}

	var result []Currency
	for _, c := range currencies {
		result = append(result, Currency{Symbol: c.ID, Name: c.Name})
	}

	db.Resolve().Create(&result)
	j.omega = time.Now()
	db.Resolve().Create(j)

	return nil
}

func (j *Job) isTimeToDoJob() bool {

	var lastJob Job

	db.Resolve().
		Where("job_type = ?", int(j.JobType)).
		Order("omega desc").
		First(&lastJob)

	if lastJob.ID == uint(0) {
		return true
	}

	return lastJob.omega.Add(time.Hour * 24).After(time.Now())
}
