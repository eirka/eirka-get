package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/eirka/eirka-libs/db"
	"github.com/eirka/eirka-libs/user"
	"github.com/eirka/eirka-libs/validate"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func testCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("cached", true)
	}
}

func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	if req != nil {
		req.Header.Set("X-Real-Ip", "123.0.0.1")
	} else {
		panic("Could not set header")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAnalytics(t *testing.T) {

	gin.SetMode(gin.ReleaseMode)

	user.Secret = "secret"

	router := gin.New()

	// params need to be verified
	router.Use(validate.ValidateParams())
	// middleware requires auth
	router.Use(user.Auth(false))
	// actual middleware ;D
	router.Use(Analytics())
	// fake cache middleware
	router.Use(testCache())

	router.GET("/index/:ib/:page", func(c *gin.Context) {
		c.String(200, "OK")
	})

	router.GET("/nocache/:id", func(c *gin.Context) {
		c.String(200, "OK")
	})

	router.GET("/thread/:id", func(c *gin.Context) {
		c.Set("controllerError", true)
		c.String(500, "YA BLEW IT")
	})

	// requires db connection
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")

	cached := performRequest(router, "GET", "/index/1/2")

	assert.Equal(t, cached.Code, 200, "HTTP request code should match")

	nocached := performRequest(router, "GET", "/nocache/1")

	assert.Equal(t, nocached.Code, 200, "HTTP request code should match")

	bad := performRequest(router, "GET", "/thread/1")

	assert.Equal(t, bad.Code, 500, "HTTP request code should match")

	assert.NoError(t, mock.ExpectationsWereMet(), "An error was not expected")

}

func TestInsertRecord(t *testing.T) {

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")

	mock.ExpectExec(`INSERT INTO analytics`).
		WithArgs("1", 1, "123.0.0.1", "/index/1/2", 200, 500, "index", "2", false).
		WillReturnResult(sqlmock.NewResult(1, 1))

	request := requestType{
		Ib:        "1",
		IP:        "123.0.0.1",
		User:      1,
		Path:      "/index/1/2",
		ItemKey:   "index",
		ItemValue: "2",
		Status:    200,
		Latency:   500,
		Cached:    false,
	}

	err = insertRecord(request)
	assert.NoError(t, err, "An error was not expected")

	assert.NoError(t, mock.ExpectationsWereMet(), "An error was not expected")

}

func TestGenerateKey(t *testing.T) {

	key := generateKey("/index/1/2")

	assert.Equal(t, key.Key, "index", "Key should match")
	assert.Equal(t, key.Value, "2", "Value should match")
}
