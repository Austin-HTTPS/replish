package server_test

import (
	"testing"

	"fmt"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/leon332157/replish/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"time"

	"github.com/valyala/fasthttp"
)

func TestServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Replish Server")
}

var _ = BeforeSuite(func() {
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
	log.SetReportCaller(false)
	log.SetLevel(log.InfoLevel)
	go startFiber()
	go server.StartForwardServer(7373)
	go server.StartReverseProxy()
	time.Sleep(3 * time.Second)
})

var client = &fasthttp.Client{}

func startFiber() {
	app := fiber.New(fiber.Config{DisableStartupMessage: true, DisableKeepalive: false})

	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendString("haha")
	})

	app.Post("/*", func(c *fiber.Ctx) error {
		return c.SendString("haha")
	})

	go app.Listen("127.0.0.1:7373")
	fmt.Println("fiber started")
}

var _ = Describe("Replish Server", func() {
	Describe("TCP Forwarder", func() {
		It("should serve 10000 requests (POST & GET)", func() {
			Expect(makeRequests(10000, 8383)).To(Succeed())
		})
	})
	Describe("Reverse Proxy", func() {
		It("should serve 10000 requests (POST & GET)", func() {
			Expect(makeRequests(10000, 8484)).To(Succeed())
		})
	})
})

func makeRequests(n int, port int) error {
	url := fmt.Sprintf("http://127.0.0.1:%v", port)
	var (
		req  fasthttp.Request
		resp fasthttp.Response
	)
	req.SetRequestURI(url)
	for x := 0; x < n; x++ {
		req.Header.SetMethod(fasthttp.MethodGet)
		err := client.DoTimeout(&req, &resp, 500*time.Millisecond)
		if err != nil {
			return fmt.Errorf("Failed on attempt %v err: %v", x, err)
		}
		if resp.StatusCode() != fasthttp.StatusOK {
			return fmt.Errorf("Unexpected status code: %d. Expecting %d", resp.StatusCode(), fasthttp.StatusOK)
		}
		// Assuming GET didn't fail, POST shouldn't fail either.
		req.Header.SetMethod(fasthttp.MethodPost)
		err = client.DoTimeout(&req, &resp, 500*time.Millisecond)
		if err != nil {
			return fmt.Errorf("Failed on attempt %v err: %v", x, err)
		}
		if resp.StatusCode() != fasthttp.StatusOK {
			return fmt.Errorf("Unexpected status code: %d. Expecting %d", resp.StatusCode(), fasthttp.StatusOK)
		}
	}
	return nil
}