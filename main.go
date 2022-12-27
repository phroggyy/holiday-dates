package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/phroggyy/holiday-dates/pkg/dates"
	"github.com/phroggyy/holiday-dates/pkg/dates/logger"
	"io"
	"net/http"
	"os"
	"strconv"
)

var (
	startYear = flag.Int("start-year", 2023, "Specify the first year you want to get public holidays for")
	years     = flag.Int("years", 1, "Specify the number of consecutive years you want to retrieve public holidays for")
)

func init() {
	if err := logger.Initialize(); err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	if len(os.Args) > 1 {
		arg := os.Args[1]
		fmt.Println(arg)
		if arg != "serve" {
			fmt.Println("The only valid arg is `serve`")
			os.Exit(1)
		}

		s := gin.Default()
		s.GET("/:startYear/:endYear", func(context *gin.Context) {
			format := ""
			if format = context.Query("format"); format == "" {
				format = "csv"
			}

			sstartYear := context.Param("startYear")
			sendYear := context.Param("endYear")
			startYear, err := strconv.Atoi(sstartYear)
			if err != nil {
				context.AbortWithStatus(http.StatusNotFound)
				return
			}
			endYear, err := strconv.Atoi(sendYear)
			if err != nil {
				context.AbortWithStatus(http.StatusNotFound)
				return
			}

			holidays, err := dates.GetHolidaysBetween(startYear, endYear)
			if err != nil {
				context.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			switch format {
			case "json":
				context.JSON(http.StatusOK, holidays)
			case "csv":
				context.Header("Content-Type", "text/csv")
				if err := writeCsvOutput(holidays, context.Writer); err != nil {
					context.AbortWithStatus(http.StatusInternalServerError)
					return
				}

			default:
				context.AbortWithStatusJSON(http.StatusUnprocessableEntity, map[string]string{"error": "format invalid"})
			}

		})

		if err := s.Run(":8080"); err != nil {
			panic(err)
		}
	}

	holidays, err := dates.GetHolidaysBetween(*startYear, *startYear+*years)
	if err != nil {
		panic(err)
	}

	if err := writeCsvOutput(holidays, os.Stdout); err != nil {
		panic(err)
	}
}

func writeCsvOutput(holidays []*dates.CountryDateItem, out io.Writer) error {
	records := [][]string{
		{"date", "name", "country_code", "country_name"},
	}

	for _, d := range holidays {
		records = append(records, []string{
			d.Date, d.Name, d.CountryCode, d.CountryName,
		})
	}

	w := csv.NewWriter(out)

	return w.WriteAll(records)
}
