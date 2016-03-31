package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"gopkg.in/olivere/elastic.v3"
)

const defaultQuerySize = 100000

func main() {
	app := cli.NewApp()
	app.Name = "eless"
	app.Usage = "cli for elasticsearch"

	var url string
	var prefix string
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "url, u",
			Value:       "http://127.0.0.1:9200",
			Usage:       "elasticsearch server url",
			Destination: &url,
		},
		cli.StringFlag{
			Name:        "prefix, p",
			Value:       "logstash-",
			Usage:       "indices prefix",
			Destination: &prefix,
		},
		cli.StringFlag{
			Name:  "separator, s",
			Value: " ",
			Usage: "output separator",
		},
		cli.StringSliceFlag{
			Name:  "field, f",
			Value: &cli.StringSlice{"@timestamp", "message"},
			Usage: "fields to return",
		},
		cli.StringSliceFlag{
			Name:  "date, d",
			Usage: "dates to return",
		},
		cli.StringSliceFlag{
			Name:  "term, t",
			Usage: "define term query, example: FIELD:TERM",
		},
	}

	app.Action = func(c *cli.Context) {
		client, err := elastic.NewSimpleClient(elastic.SetURL(url))
		if err != nil {
			panic(err)
		}

		dates := c.StringSlice("date")
		if dates == nil || len(dates) == 0 {
			dates = []string{yesterdayDate(), currentDate()}
		}

		indices := make([]string, len(dates))
		for i, date := range dates {
			indices[i] = prefix + date
		}

		fields := c.StringSlice("field")

		queriesArray := c.StringSlice("term")
		globalQuery := elastic.NewBoolQuery()
		for _, queryString := range queriesArray {
			queryStringArray := strings.Split(queryString, ":")
			field, value := queryStringArray[0], queryStringArray[1]
			query := elastic.NewMatchPhraseQuery(field, value)
			globalQuery.Must(query)
		}

		from := 0
		for processPortion(client, indices, globalQuery, fields, c.String("separator"), from) {
			from += defaultQuerySize
		}
	}

	app.Run(os.Args)
}

func processPortion(client *elastic.Client, indices []string, globalQuery *elastic.BoolQuery, fields []string, separator string, from int) bool {
	searchResult, err := client.Search(indices...).
		Query(globalQuery).
		Sort("@timestamp", true).
		Sort("offset", true).
		Fields(fields...).
		From(from).
		Size(defaultQuerySize).
		Do()

	if err != nil {
		panic(err)
	}

	if searchResult.Hits != nil {
		if int64(from) > searchResult.TotalHits() {
			return false
		}
		for _, hit := range searchResult.Hits.Hits {
			for _, field := range fields {
				for _, fieldValue := range hit.Fields[field].([]interface{}) {
					fmt.Print(fieldValue)
				}
				fmt.Print(separator)
			}
			fmt.Print("\n")
		}
		return true
	}

	return false
}

func currentDate() string {
	return time.Now().Format("2006.01.02")
}

func yesterdayDate() string {
	return time.Now().AddDate(0, 0, -1).Format("2006.01.02")
}
