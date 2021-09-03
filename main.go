package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"nuchal-api/model"
	"nuchal-api/util"
	"nuchal-api/view"
	"os"
	"strconv"
	"strings"
	"time"
)

var productsInitialized = false

func init() {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}

	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("***%s****", i)
	}
	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%s:", i)
	}
	output.FormatFieldValue = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("%s", i))
	}
}

func main() {

	router := gin.Default()

	router.Handlers = append(router.Handlers, func(c *gin.Context) {

		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	router.GET("/chart/:userID/:productID/:alpha/:omega", getChart)

	router.GET("/productArr/:userID", getProductArr)
	router.GET("/productIDs/:userID", getProductIDs)

	router.GET("/quotes/:userID", getQuotes)

	router.PUT("/user", saveUser)
	router.GET("/users", getUsers)
	router.GET("/user/:userID", getUserByID)
	router.DELETE("/user/:userID", deleteUser)

	router.PUT("/pattern")
	router.GET("/pattern/:patternID")
	router.DELETE("/pattern/:patternID", deleteUser)

	router.Run("localhost:9080")
}

func getChart(c *gin.Context) {
	userID := userID(c)
	productID := c.Param("productID")
	alpha := util.StringToInt64(c.Param("alpha"))
	omega := util.StringToInt64(c.Param("omega"))
	c.IndentedJSON(http.StatusOK, view.NewChartData(userID, productID, alpha, omega))
}

func deleteUser(c *gin.Context) {
	model.DeleteUser(userID(c))
}

func saveUser(c *gin.Context) {

	data, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Err(err).Send()
		return
	}

	var u model.User
	if err := json.Unmarshal(data, &u); err != nil {
		log.Err(err).Send()
		return
	}

	model.SaveUser(u)
}

func getUsers(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, model.FindUsers())
}

func getUserByID(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, user(c))
}

func getProductIDs(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, model.ProductIDs)
}

func getProductArr(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, model.ProductArr)
}

func getQuotes(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, view.GetQuotes(userID(c)))
}

func userID(c *gin.Context) uint {

	stringID := c.Param("userID")

	intID, err := strconv.Atoi(stringID)
	if err != nil {
		log.Err(err).Send()
	}

	uintID := uint(intID)

	if !productsInitialized {

		if err := model.InitProducts(uintID); err != nil {
			log.Err(err).Send()
		} else {
			productsInitialized = true
		}
	}

	return uintID
}

func user(c *gin.Context) *model.User {

	uintID := userID(c)

	if !productsInitialized {

		if err := model.InitProducts(uintID); err != nil {
			log.Err(err).Send()
		} else {
			productsInitialized = true
		}
	}

	user := model.FindUserByID(uintID)

	return &user
}
