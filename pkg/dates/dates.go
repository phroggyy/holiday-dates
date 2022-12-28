package dates

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/phroggyy/holiday-dates/pkg/dates/logger"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"net/url"
)

const APIBaseUrl = "https://date.nager.at/api/v3/"

// A map from country to county/region
var authoritativeCounties = map[string]string{
	"GB": "GB-ENG",
	"US": "US-NY",
	"CA": "CA-ON",
}

type CountryListItem struct {
	CountryCode string `json:"countryCode"`
	Name        string `json:"name"`
}

type CountryDateItem struct {
	Date        string   `json:"date"`
	LocalName   string   `json:"localName"`
	Name        string   `json:"name"`
	CountryCode string   `json:"countryCode"`
	CountryName string   `json:"countryName,omitempty"`
	Fixed       bool     `json:"fixed"`
	Global      bool     `json:"global"`
	Counties    []string `json:"counties"`
}

func GetHolidaysBetween(startYear, endYear int) ([]*CountryDateItem, error) {
	logger.SLog.Debugw("retrieving available countries")
	r, err := http.Get(requestUrl("AvailableCountries"))
	if err != nil {
		logger.SLog.Errorw("failed to retrieve countries", "error", err)
		return nil, err
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var countries []*CountryListItem
	err = json.Unmarshal(b, &countries)
	if err != nil {
		return nil, err
	}

	var output []*CountryDateItem
	currentYear := startYear
	for currentYear <= endYear {
		eg, _ := errgroup.WithContext(context.Background())
		dates := make(chan []*CountryDateItem, len(countries))
		for _, country := range countries {
			c := country
			eg.Go(func() error {
				r, err := http.Get(requestUrl(fmt.Sprintf("publicholidays/%d/%s", currentYear, c.CountryCode)))

				if err != nil {
					logger.SLog.Errorw("failed to retrieve holidays", "country", c.CountryCode, "error", err)
					return err
				}

				logger.SLog.Debugw("retrieved holidays for country", "year", currentYear, "country", c.CountryCode)

				b, err := io.ReadAll(r.Body)
				if err != nil {
					return err
				}

				var out []*CountryDateItem
				err = json.Unmarshal(b, &out)

				if err != nil {
					return err
				}

				for _, o := range out {
					o.CountryName = c.Name
				}

				dates <- out
				return nil
			})
		}

		err = eg.Wait()
		if err != nil {
			logger.SLog.Errorw("failed to retrieve holidays for year", "year", currentYear, "error", err)
			return nil, err
		}
		close(dates)
		for d := range dates {
			for _, date := range d {
				if len(date.Counties) == 0 || isAuthoritativeCounty(date) {
					output = append(output, date)
				}
			}
		}
		logger.SLog.Debugw("retrieved all holidays for year", "year", currentYear)

		currentYear = currentYear + 1
	}

	return output, nil
}

func requestUrl(p string) string {
	r, _ := url.JoinPath(APIBaseUrl, p)
	return r
}

func isAuthoritativeCounty(d *CountryDateItem) bool {
	for _, county := range d.Counties {
		if authoritativeCounties[d.CountryCode] == county {
			return true
		}
	}

	return false
}
