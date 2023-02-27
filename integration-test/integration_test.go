package integration_test

import (
	"log"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	. "github.com/Eun/go-hit"
)

const (
	host            = "app:8080"
	healthPath      = "http://" + host + "/"
	attempts        = 10
	defaultLoginCap = 10
	basePath        = "http://" + host + "/v1"
)

func TestMain(m *testing.M) {
	if err := healthCheck(attempts); err != nil {
		log.Fatalf("integration tests: host %s is not available: %s", host, err)
	}

	log.Printf("integration tests: host %s is available", host)

	code := m.Run()
	os.Exit(code)
}

func healthCheck(attempts int) error {
	var err error
	for attempts > 0 {
		err = Do(Get(healthPath), Expect().Status().Equal(http.StatusOK))
		if err == nil {
			return nil
		}

		log.Printf("integration tests: url %s is not available, attempts left: %d", healthPath, attempts)
		time.Sleep(time.Second)
		attempts--
	}
	return err
}

func TestHTTPCheckBucket(t *testing.T) {
	var wg sync.WaitGroup
	body := `{
		"login": "Vasya",
		"password": "qwerty",
		"ip": "172.16.254.1"
	}`
	Test(t,
		Description("Check bucket Success"),
		Post(basePath+"/check"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().String(body),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".ok").Equal("true"),
	)

	body = `{
		"login": "Petya",
		"password": "qwerty",
		"ip": "172.1600.254.1"
	}`
	Test(t,
		Description("Check bucket Fail"),
		Post(basePath+"/check"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().String(body),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().String().Equal("invalid IP address\n"),
	)

	body = `{
		"login": "Ivan",
		"password": "qwerty",
		"ip": "172.16.254.1"
	}`

	for i := 0; i < defaultLoginCap; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			MustDo(
				Description("Check bucket Success"),
				Post(basePath+"/check"),
				Send().Headers("Content-Type").Add("application/json"),
				Send().Body().String(body),
				Expect().Status().Equal(http.StatusOK),
				Expect().Body().JSON().JQ(".ok").Equal("true"),
			)
		}()
	}
	wg.Wait()

	Test(t,
		Description("Check bucket Success but ok=false"),
		Post(basePath+"/check"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().String(body),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".ok").Equal("false"),
	)
}

func TestAddToBlacklist(t *testing.T) {
	body := `{
		"ip": "192.0.2.2/24"
	}`
	Test(t,
		Description("Add to blacklist Success"),
		Post(basePath+"/add/blacklist"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().String(body),
		Expect().Status().Equal(http.StatusOK),
	)

	body = `{
		"ip": "1902.0.2.2/24"
	}`
	Test(t,
		Description("Add to blacklist Fail"),
		Post(basePath+"/add/blacklist"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().String(body),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().String().Equal("invalid IP/net address\n"),
	)
}

func TestDeleteFromBlacklist(t *testing.T) {
	body := `{
		"ip": "192.0.2.2/24"
	}`
	Test(t,
		Description("Delete from blacklist Success"),
		Post(basePath+"/remove/blacklist"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().String(body),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".deleted").Equal(1),
	)

	body = `{
		"ip": "1902.0.2.2/24"
	}`
	Test(t,
		Description("Delete from blacklist Success width zero result"),
		Post(basePath+"/remove/blacklist"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().String(body),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".deleted").Equal(0),
	)
}

func TestAddToWhitelist(t *testing.T) {
	body := `{
		"ip": "193.0.2.2/22"
	}`
	Test(t,
		Description("Add to whitelist Success"),
		Post(basePath+"/add/whitelist"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().String(body),
		Expect().Status().Equal(http.StatusOK),
	)

	body = `{
		"ip": "1903.0.2.2/22"
	}`
	Test(t,
		Description("Add to whitelist Fail"),
		Post(basePath+"/add/whitelist"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().String(body),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().String().Equal("invalid IP/net address\n"),
	)
}

func TestDeleteFromWhitelist(t *testing.T) {
	body := `{
		"ip": "193.0.2.2/22"
	}`
	Test(t,
		Description("Delete from whitelist Success"),
		Post(basePath+"/remove/whitelist"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().String(body),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".deleted").Equal(1),
	)

	body = `{
		"ip": "1903.0.2.2/22"
	}`
	Test(t,
		Description("Delete from whitelist Success width zero result"),
		Post(basePath+"/remove/whitelist"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().String(body),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".deleted").Equal(0),
	)
}

func TestRemoveKeys(t *testing.T) {
	body := `{
		"login": "Vitya",
		"password": "qazwsx",
		"ip": "162.16.254.1"
	}`

	MustDo(
		Description("Check bucket Success"),
		Post(basePath+"/check"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().String(body),
		Expect().Status().Equal(http.StatusOK),
	)

	body = `{
		"keys": ["Vitya", "162.16.254.1", "qazwsx"]
	}`

	Test(t,
		Description("Delete keys from buckets"),
		Post(basePath+"/remove/keys"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().String(body),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".deleted").Equal([]any{string("Vitya"), string("162.16.254.1"), string("qazwsx")}),
	)

	Test(t,
		Description("Repeat width null result"),
		Post(basePath+"/remove/keys"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().String(body),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".deleted").Equal(nil),
	)

	body = `{
		"keys": ["Ivan", "Vasya", "qwerty", "172.16.254.1"]
	}`

	Test(t,
		Description("Teardown (clean keys)"),
		Post(basePath+"/remove/keys"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().String(body),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".deleted").Equal(
			[]any{string("Ivan"), string("Vasya"), string("qwerty"), string("172.16.254.1")}),
	)
}
