package main

import (
	"fmt"
	"os"
	"time"

	"github.com/codegangsta/cli"
	"gopkg.in/olivere/elastic.v3"
)

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
	}

	app.Action = func(c *cli.Context) {
		client, err := elastic.NewSimpleClient(elastic.SetURL(url))
		if err != nil {
			panic(err)
		}

		dates := c.StringSlice("date")
		if dates == nil || len(dates) == 0 {
			dates = []string{currentDate()}
		}

		indices := make([]string, len(dates))
		for i, date := range dates {
			indices[i] = prefix + date
		}

		fields := c.StringSlice("field")

		searchResult, err := client.Search().
			Index(indices...).
			Sort("@timestamp", false).
			Sort("offset", false).
			Fields(fields...).
			From(0).
			Size(10).
			Do()

		if err != nil {
			panic(err)
		}

		if searchResult.Hits != nil {
			for _, hit := range searchResult.Hits.Hits {
				for _, field := range fields {
					fmt.Print(hit.Fields[field].(string))
					fmt.Print(c.String("separator"))
				}
				fmt.Print("\n")
			}
		}
	}

	app.Run(os.Args)
}

func currentDate() string {
	return time.Now().Format("2006.01.02")
}
