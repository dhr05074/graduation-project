package main

import (
	"github.com/labstack/echo/v4"
	"strconv"
	"traffic-shaping/lib"
)

func main() {
	pq := make(chan int, 1<<10)
	manager := lib.NewQueueManager(pq)
	inbound := manager.Start()

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			pri, ok := c.Request().Header["X-Priority"]
			if !ok {
				pri = []string{"5"}
			}

			pr, err := strconv.Atoi(pri[0])
			if err != nil {
				return c.String(500, "Error")
			}

			select {
			case <-c.Request().Context().Done():
				return c.String(500, "Error")
			case inbound <- pr:
				return next(c)
			default:
				return c.String(429, "TMR")
			}
		}
	})
	e.GET("/", func(c echo.Context) error {

		return c.String(200, "Hello world!")
	})

	e.Logger.Fatal(e.Start(":8080"))
}
