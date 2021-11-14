package model

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"nuchal-api/db"
	"nuchal-api/util"
	"time"
)

type JobType string

const (
	InitAllCBProducts JobType = "init all cb products"
	InitOneDayOfRates         = "init one day of rates"
	SellAllOfAProduct         = "sell all of a product"
)

type Job struct {
	gorm.Model
	UserID  uint    `gorm:"user_id"`
	JobType JobType `gorm:"job_type"`
	Message string  `json:"message"`
}

func init() {
	db.Migrate(&Job{})
}

func lookToSell() {
	go func() {
		for {

			log.Trace().Msg("broker looking for work")

			var jobs []Job

			tx := db.
				Resolve().
				Where("job_type = ?", SellAllOfAProduct).
				Find(&jobs)

			if tx.Error != nil {
				log.Err(tx.Error).Send()
				return
			}

			time.Sleep(time.Second * 15)
		}
	}()
	select {}
}

func NewSellRequest(userID uint, productID string, size int) {
	//job := Job{UserID:  userID, JobType: SellAllOfAProduct}

}

func PerformAllJobs(userID uint) error {

	productJob := &Job{
		Model:   gorm.Model{},
		UserID:  userID,
		JobType: InitAllCBProducts,
	}

	db.Resolve().Create(&productJob)

	if err := productJob.Perform(); err != nil {
		log.Error().Err(err).Stack().Send()
		return err
	}

	ratesJob := &Job{
		Model:   gorm.Model{},
		UserID:  userID,
		JobType: InitOneDayOfRates,
	}

	db.Resolve().Create(&ratesJob)

	if err := ratesJob.Perform(); err != nil {
		log.Error().Err(err).Stack().Send()
		return err
	}

	return nil
}

func (j *Job) Perform() error {
	switch j.JobType {
	case InitAllCBProducts:
		return j.initAllCBProducts()
	case InitOneDayOfRates:
		return j.initOneDayOfRates()
	}
	return nil
}

func (j *Job) initOneDayOfRates() error {
	if products, err := FindAllProducts(); err != nil {
		return err
	} else {
		for _, product := range products {
			omega := time.Now()
			alpha := omega.Add(time.Hour * -24)
			if _, err := GetRates(j.UserID, product.ID, alpha.Unix(), omega.Unix()); err != nil {
				log.Error().Err(err).Stack().Send()
			}
		}
		return nil
	}
}

func (j *Job) initAllCBProducts() error {

	if !j.isTimeToDoJob() {
		return nil
	}

	var cbCurrencies []cb.Currency
	var cbProducts []cb.Product
	var err error

	u := FindUserByID(j.UserID)

	if cbCurrencies, err = u.Client().GetCurrencies(); err != nil {
		return err
	}

	if cbProducts, err = u.Client().GetProducts(); err != nil {
		return err
	}

	currencyMap := map[string]cb.Currency{}
	for _, c := range cbCurrencies {
		currencyMap[c.ID] = c
	}

	var result []Product
	for _, p := range cbProducts {
		result = append(result, Product{
			StrModel: StrModel{ID: p.ID},
			Name:     currencyMap[p.BaseCurrency].Name,
			Base:     p.BaseCurrency,
			Quote:    p.QuoteCurrency,
			Min:      util.StringToFloat64(p.BaseMinSize),
			Max:      util.StringToFloat64(p.BaseMaxSize),
			Step:     util.StringToFloat64(p.QuoteIncrement),
		})
	}

	db.Resolve().Save(&result)
	db.Resolve().Save(j)

	return nil
}

func (j *Job) isTimeToDoJob() bool {

	var lastJob Job

	db.Resolve().
		Where("job_type = ?", string(j.JobType)).
		Order("omega desc").
		First(&lastJob)

	timeToDoJob := lastJob.ID == uint(0) || lastJob.UpdatedAt.After(time.Now().Add(time.Hour*24))

	log.Trace().
		Str("job type", string(j.JobType)).
		Bool("time to do job", timeToDoJob).
		Send()

	return timeToDoJob
}
