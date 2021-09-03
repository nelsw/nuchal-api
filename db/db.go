package db

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	reggol "gorm.io/gorm/logger"
	gol "log"
	"os"
	"strconv"
	"time"
)

type config struct {
	Host string `envconfig:"POSTGRES_HOST"`
	User string `envconfig:"POSTGRES_USER"`
	Pass string `envconfig:"POSTGRES_PASSWORD"`
	Name string `envconfig:"POSTGRES_DB"`
	Port int    `envconfig:"POSTGRES_PORT"`
}

var cfg *config

var pg *gorm.DB

func init() {
	cfg = new(config)
	if envs, err := godotenv.Read(".env"); err == nil {
		if port, err := strconv.Atoi(envs["POSTGRES_PORT"]); err == nil {
			cfg.Host = envs["POSTGRES_HOST"]
			cfg.User = envs["POSTGRES_USER"]
			cfg.Name = envs["POSTGRES_DB"]
			cfg.Pass = envs["POSTGRES_PASSWORD"]
			cfg.Port = port
		}
	}
}

func (c *config) dsn() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d", c.Host, c.User, c.Pass, c.Name, c.Port)
}

func (c *config) validate() error {

	if pg, err := openDB(c.dsn()); err != nil {
		return err
	} else if sql, err := pg.DB(); err != nil {
		return err
	} else if err := sql.Ping(); err != nil {
		return err
	} else if err := sql.Close(); err != nil {
		return err
	}

	return nil
}

func openDB(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: reggol.New(
			gol.New(os.Stdout, "\r\n", gol.LstdFlags), // io writer
			reggol.Config{
				SlowThreshold:             time.Second,   // Slow SQL threshold
				LogLevel:                  reggol.Silent, // Log level
				IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
				Colorful:                  false,         // Disable color
			},
		),
	})
}

func NewDB(vv ...interface{}) *gorm.DB {

	db, err := openDB(cfg.dsn())
	if err == nil && vv != nil && len(vv) > 0 {
		for _, v := range vv {
			if err = db.AutoMigrate(v); err != nil {
				break
			}
		}
	}

	if err != nil {
		log.Debug().Err(err).Send()
	}

	return db
}

func Resolve() *gorm.DB {
	if pg == nil {
		pg = NewDB()
	}
	return pg
}

func Migrate(v interface{}) {
	err := Resolve().AutoMigrate(v)
	if err != nil {
		log.Debug().Err(err).Send()
	}
}
