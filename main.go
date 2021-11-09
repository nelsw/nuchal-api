package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"nuchal-api/model"
	"nuchal-api/util"
	"os"
	"strconv"
	"strings"
	"time"
)

func init() {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	zerolog.SetGlobalLevel(zerolog.TraceLevel)

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

	if err := model.PerformAllJobs(uint(1)); err != nil {
		log.Err(err).Stack().Send()
	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {

	router := gin.Default()

	router.Use(CORS())

	/*
		sim
	*/
	router.GET("/sim/pattern/:patternID/:alpha/:omega", getPatternSim)

	/*
		product
	*/
	router.GET("/product/:id", findProductByID)
	router.GET("/products", getAllProducts)
	router.GET("/products/:quote", getAllProductsByQuote)

	/*
		portfolio
	*/
	router.GET("/portfolio/:userID", getPortfolio)

	/*
		user
	*/
	router.PUT("/user", saveUser)
	router.GET("/users", getUsers)
	router.GET("/user/:userID", getUserByID)
	router.DELETE("/user/:userID", deleteUser)

	/*
		pattern
	*/
	router.PUT("/pattern", savePattern)
	router.GET("/patterns/:userID", getPatterns)
	router.GET("/pattern/:patternID", getPattern)
	router.DELETE("/pattern/:patternID", deletePattern)

	/*
		trade
	*/
	router.POST("/trade/:patternID", startTrading)

	/*
		order
	*/
	router.DELETE("/order/:userID/:orderID", deleteOrder)
	router.POST("/order/:userID", postOrder)

	/*
		chart
	*/
	router.GET("/chart/product/:userID/:productID/:alpha/:omega", getProductChart)

	router.Run("localhost:9080")
}

func getProductChart(c *gin.Context) {

	alpha := util.StringToInt64(c.Param("alpha"))
	omega := util.StringToInt64(c.Param("omega"))

	var chart model.Chart
	var err error

	if chart, err = model.NewProductChart(userID(c), c.Param("productID"), alpha, omega); err != nil {
		c.Status(400)
		return
	}

	c.IndentedJSON(http.StatusOK, chart)
}

func deleteOrder(c *gin.Context) {
	if err := model.DeleteOrder(userID(c), c.Param("orderID")); err != nil {
		log.Err(err).Stack().Send()
		c.Status(http.StatusBadRequest)
		return
	}
	c.Status(200)
}

func postOrder(c *gin.Context) {

	data, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Err(err).Stack().Send()
		c.Status(http.StatusBadRequest)
		return
	}

	var o cb.Order
	if err = json.Unmarshal(data, &o); err != nil {
		log.Err(err).Stack().Send()
		c.Status(http.StatusBadRequest)
		return
	}

	if err = model.PostOrder(userID(c), o); err != nil {
		log.Err(err).Stack().Send()
		c.Status(http.StatusBadRequest)
		return
	}

	c.Status(http.StatusOK)
}

func getPattern(c *gin.Context) {
	patternID, err := strconv.Atoi(c.Param("patternID"))
	if err != nil {
		log.Err(err).Stack().Send()
		c.Status(http.StatusBadRequest)
		return
	}
	c.IndentedJSON(http.StatusOK, model.FindPattern(uint(patternID)))
}

func startTrading(c *gin.Context) {
	patternID, err := strconv.Atoi(c.Param("patternID"))
	if err != nil {
		log.Err(err).Stack().Send()
		c.Status(http.StatusBadRequest)
		return
	}
	model.NewTrade(uint(patternID))
	c.Status(200)
}

func getPortfolio(c *gin.Context) {
	portfolio, err := model.GetPortfolio(userID(c))
	if err != nil {
		log.Err(err).Stack().Send()
		c.Status(http.StatusBadRequest)
		return
	}
	c.IndentedJSON(http.StatusOK, portfolio)
}

func getPatternSim(c *gin.Context) {
	alpha := util.StringToInt64(c.Param("alpha"))
	omega := util.StringToInt64(c.Param("omega"))
	patternID, err := strconv.Atoi(c.Param("patternID"))
	if err != nil {
		log.Err(err).Stack().Send()
		c.Status(http.StatusBadRequest)
		return
	}

	var sim model.Sim
	sim, err = model.NewSim(uint(patternID), alpha, omega)
	if err != nil {
		log.Err(err).Stack().Send()
		c.Status(http.StatusBadRequest)
		return
	}

	c.IndentedJSON(http.StatusOK, sim)
}

func deletePattern(c *gin.Context) {
	patternID, err := strconv.Atoi(c.Param("patternID"))
	if err != nil {
		log.Err(err).Stack().Send()
		c.Status(http.StatusBadRequest)
		return
	}
	model.DeletePattern(uint(patternID))
	c.Status(http.StatusOK)
}

func getPatterns(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, model.GetPatterns(userID(c)))
}

func savePattern(c *gin.Context) {
	data, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Err(err).Stack().Send()
		c.Status(http.StatusBadRequest)
	}
	var p model.Pattern
	if err := json.Unmarshal(data, &p); err != nil {
		log.Err(err).Stack().Send()
		c.Status(http.StatusBadRequest)
	}
	p.Save()
	c.IndentedJSON(http.StatusOK, model.FindPatternByID(p.ID))
}

func deleteUser(c *gin.Context) {
	model.DeleteUser(userID(c))
	c.Status(http.StatusOK)
}

func saveUser(c *gin.Context) {

	data, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Err(err).Stack().Send()
		c.Status(http.StatusBadRequest)
		return
	}

	var u model.User
	if err := json.Unmarshal(data, &u); err != nil {
		log.Err(err).Stack().Send()
		c.Status(http.StatusBadRequest)
		return
	}

	c.IndentedJSON(http.StatusOK, model.SaveUser(u))
}

func getUsers(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, model.FindUsers())
}

func getUserByID(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, user(c))
}

/*
	products
*/

func findProductByID(c *gin.Context) {
	p, err := model.FindProductByID(c.Param("id"))
	if err != nil {
		log.Err(err).Stack().Send()
		c.Status(http.StatusInternalServerError)
		return
	}
	c.IndentedJSON(http.StatusOK, p)
}

func getAllProducts(c *gin.Context) {
	p, err := model.FindAllProducts()
	if err != nil {
		log.Err(err).Stack().Send()
		c.Status(http.StatusInternalServerError)
		return
	}
	c.IndentedJSON(http.StatusOK, p)
}

func getAllProductsByQuote(c *gin.Context) {
	p, err := model.FindAllProductsByQuote(c.Param("quote"))
	if err != nil {
		log.Err(err).Stack().Send()
		c.Status(http.StatusInternalServerError)
		return
	}
	c.IndentedJSON(http.StatusOK, p)
}

func userID(c *gin.Context) uint {
	userID, err := strconv.Atoi(c.Param("userID"))
	if err != nil {
		log.Err(err).Stack().Send()
	}
	return uint(userID)
}

func user(c *gin.Context) *model.User {
	uintID := userID(c)
	user := model.FindUserByID(uintID)
	return &user
}
