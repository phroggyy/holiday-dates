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

type CountryListItem struct {
	CountryCode string `json:"countryCode"`
	Name        string `json:"name"`
}

type CountryDateItem struct {
	Date        string `json:"date"`
	LocalName   string `json:"localName"`
	Name        string `json:"name"`
	CountryCode string `json:"countryCode"`
	CountryName string `json:"countryName,omitempty"`
	Fixed       bool   `json:"fixed"`
	Global      bool   `json:"global"`
	//LaunchYear  string   `json:"launchYear,omitempty"`
	//Counties    []string `json:"counties"`
	//Types       []string `json:"types"`
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
			eg.Go(func() error {
				r, err := http.Get(requestUrl(fmt.Sprintf("publicholidays/%d/%s", currentYear, country.CountryCode)))

				if err != nil {
					logger.SLog.Errorw("failed to retrieve holidays", "country", country.CountryCode, "error", err)
					return err
				}

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
					o.CountryName = country.Name
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
			output = append(output, d...)
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
