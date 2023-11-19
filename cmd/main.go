package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"strconv"
	"time"
	"traffic-shaping/manager"
)

func main() {
	bucketManager := manager.NewBucketManager(1<<20, 5*time.Millisecond)
	bucketManager.Start()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			priorityHeader, ok := c.Request().Header["X-Priority"]
			// default priority is 5
			if !ok {
				priorityHeader = []string{"5"}
			}

			// priority value must be between 1 and 5
			priorityValue, err := strconv.Atoi(priorityHeader[0])
			if err != nil || priorityValue < 1 || priorityValue > 5 {
				return c.String(500, "Invalid priority value. It must be between 1 and 5.")
			}

			select {
			case <-c.Request().Context().Done():
				return c.String(500, "Content canceled.")
			default:
				if err := bucketManager.Consume(uint(priorityValue)); err != nil {
					return c.String(429, "Too Many Requests. Try again later.")
				}

				return next(c)
			}
		}
	})
	e.GET("/", func(c echo.Context) error {
		return c.String(200, "Hello world!")
	})

	e.Logger.Fatal(e.Start(":8080"))
}
