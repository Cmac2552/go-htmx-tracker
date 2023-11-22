package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "modernc.org/sqlite"
)

type item struct {
	id        int
	thingType string
	count     string
	date      string
}

type items struct {
	date  string
	items []item
}

type CurrentDate struct {
	currentDate string
	lock        sync.Mutex
}

var (
	currDate = &CurrentDate{}
)

func (currDate *CurrentDate) setDate(date string) {
	currDate.lock.Lock()
	currDate.currentDate = date
	currDate.lock.Unlock()

}
func (currDate *CurrentDate) getDate() string {
	currDate.lock.Lock()
	defer currDate.lock.Unlock()
	return currDate.currentDate

}

func putIntoDB(thingType string, count string, date string) {
	db, err := sql.Open("sqlite", "./DB1.db")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	if err != nil {
		fmt.Println(err)
	}

	count_int, err := strconv.Atoi(count)

	db.Exec("INSERT INTO table1 (type, count, date) VALUES (?, ?, ?)", thingType, count_int, date)
}

func pullFromDB(dates []string) []items {
	db, err := sql.Open("sqlite", "./DB1.db")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	dayContainer := make([]items, len(dates))
	for i := 0; i < len(dates); i++ {

		result, err := db.Query("SELECT * FROM table1 WHERE date = ?", dates[i])
		if err != nil {
			fmt.Println(err)
		}

		for result.Next() {
			var thing item
			err := result.Scan(&thing.id, &thing.thingType, &thing.count, &thing.date)
			if err != nil {
				fmt.Println(err)
			}
			dayContainer[i].items = append(dayContainer[i].items, thing)
		}
		dayContainer[i].date = dates[i]
		result.Close()
	}

	return dayContainer
}

func clearDB() {
	db, err := sql.Open("sqlite", "./DB1.db")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	db.Exec("DELETE FROM table1")

}

func isTheDayRight() time.Time {
	db, err := sql.Open("sqlite", "./DB1.db")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	result, err := db.Query("SELECT * from dayTracker")
	var timeInDb string
	result.Next()
	result.Scan(&timeInDb)
	result.Close()
	if err != nil {
		fmt.Println(nil)
	}

	currentTime, err := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	if err != nil {
		fmt.Println(nil)
	}

	parsedDbTime, err := time.Parse("2006-01-02", timeInDb)
	if err != nil {
		fmt.Println(nil)
	}

	if parsedDbTime.Equal(currentTime) {

		return currentTime
	}
	fmt.Println(currentTime.Format("2006-01-02"))
	db.Exec("UPDATE dayTracker SET day=? WHERE day=?", currentTime.Format("2006-01-02"), parsedDbTime.Format("2006-01-02"))

	return currentTime

}

func dateHelper(dateToFormat string) string {
	convertedTime, err := time.Parse("2006-01-02", dateToFormat)
	if err != nil {
		fmt.Println(err)
	}

	return convertedTime.Format("January 2 2006")
}

func getLastDays(firstDay string, numDays int) []string {
	days := make([]string, numDays)
	parsedTime, err := time.Parse("2006-01-02", firstDay)
	if err != nil {
		fmt.Println(err)
	}
	for i := 0; i < numDays; i++ {
		days[i] = parsedTime.AddDate(0, 0, i*-1).Format("2006-01-02")
	}
	return days
}

func hand1(c echo.Context) error {

	if c.QueryParam("date") != "" {
		currDate.setDate(c.QueryParam("date"))
	}

	items := pullFromDB(getLastDays(currDate.getDate(), 5))
	fmt.Println(items)
	return indexPage(items, dateHelper(currDate.getDate())).Render(context.Background(), c.Response().Writer)
}

func hand2(c echo.Context) error {

	putIntoDB(c.FormValue("type"), c.FormValue("count"), currDate.getDate())

	items := pullFromDB([]string{currDate.getDate()})
	return forLoopTest(items[0].items).Render(context.Background(), c.Response().Writer)

}

func hand3(c echo.Context) error {

	clearDB()
	return forLoopTest(make([]item, 0)).Render(context.Background(), c.Response().Writer)
}

func hand5(c echo.Context) error {
	return datePicker().Render(context.Background(), c.Response().Writer)
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/assets", "assets")

	currDate.setDate(isTheDayRight().Format("2006-01-02"))

	e.GET("/", hand1)
	e.GET("/date-picker", hand5)
	e.POST("/new-item", hand2)
	e.DELETE("/", hand3)
	e.Logger.Fatal(e.Start(":1323"))
}
