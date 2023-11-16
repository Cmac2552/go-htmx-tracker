package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "modernc.org/sqlite"
)

var (
	lock = sync.Mutex{}
)

type item struct {
	id        int
	thingType string
	count     string
}

type CustomContext struct {
	echo.Context
}

func (c *CustomContext) putIntoDB(thingType string, count string) {
	db, err := sql.Open("sqlite", "./DB1.db")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	if err != nil {
		fmt.Println(err)
	}

	count_int, err := strconv.Atoi(count)

	db.Exec("INSERT INTO table1 (type, count) VALUES (?, ?)", thingType, count_int)
}

func (c *CustomContext) pullFromDB() []item {
	db, err := sql.Open("sqlite", "./DB1.db")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	result, err := db.Query("SELECT * FROM table1")
	if err != nil {
		fmt.Println(err)
	}

	var thingContainer []item
	for result.Next() {
		var thing item
		err := result.Scan(&thing.id, &thing.thingType, &thing.count)
		if err != nil {
			fmt.Println(err)
		}
		thingContainer = append(thingContainer, thing)
	}

	return thingContainer
}

func (c *CustomContext) clearDB() {
	db, err := sql.Open("sqlite", "./DB1.db")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	db.Exec("DELETE FROM table1")

}

func hand1(c echo.Context) error {
	cc := c.(*CustomContext)
	items := cc.pullFromDB()
	return indexPage(items).Render(context.Background(), c.Response().Writer)
}

func hand2(c echo.Context) error {
	cc := c.(*CustomContext)
	fmt.Println(cc.FormValue("date"))
	cc.putIntoDB(cc.FormValue("type"), cc.FormValue("count"))
	items := cc.pullFromDB()
	return forLoopTest(items).Render(context.Background(), c.Response().Writer)
}

func hand3(c echo.Context) error {
	cc := c.(*CustomContext)
	cc.clearDB()
	return forLoopTest(make([]item, 0)).Render(context.Background(), c.Response().Writer)
}

func hand4(c echo.Context) error {
	// cc := c.(*CustomContext)
	fmt.Println("here")
	return c.HTML(http.StatusOK, "")
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/assets", "assets")

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &CustomContext{c}
			return next(cc)
		}
	})

	e.GET("/", hand1)
	e.GET("/date", hand4)
	e.POST("/new-item", hand2)
	e.DELETE("/", hand3)
	e.Logger.Fatal(e.Start(":1323"))
}
