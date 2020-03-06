package synth

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	"bitbucket.org/SeheonKim/albatros4/model"
)

func s(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

func d(v interface{}) string {
	return fmt.Sprintf("%d", v)
}

func WriteCsv(out io.Writer, hhs <-chan *model.Household) {
	csv := csv.NewWriter(out)
	defer csv.Flush()

	csv.Write([]string{
		"Hhid",
		"Home",
		"Gem",
		"Urb",
		"Comp",
		"Child",
		"Day",
		"SEC",
		"Ncar",
		"Driver",
		"FEV",
		"PHEV",
		"Age1",
		"Gender1",
		"Work1",
		"Driver1",
		"Age2",
		"Gender2",
		"Work2",
		"Driver2",
	})

	for hh := range hhs {
		var Age1 = 999999
		var Gender1 = 999999
		var Work1 = 999999
		var Driver1 = 999999
		var Age2 = 999999
		var Gender2 = 999999
		var Work2 = 999999
		var Driver2 = 999999

		for i := range hh.Member {
			if i == 0 {
				Age1 = int(hh.Member[i].Age)
				Gender1 = int(hh.Member[i].Gender)
				Work1 = int(hh.Member[i].Work)
				if hh.Member[i].IsDriver {
					Driver1 = 1
					hh.Driver = 1
				}
			} else {
				Age2 = int(hh.Member[i].Age)
				Gender2 = int(hh.Member[i].Gender)
				Work2 = int(hh.Member[i].Work)
				if hh.Member[i].IsDriver {
					Driver2 = 1
				}
			}
		}
		csv.Write([]string{
			d(hh.ID),
			d(hh.Home),
			d(hh.WoGem),
			d(hh.Urb),
			d(hh.Comp),
			d(hh.Child),
			d(hh.Day),
			d(hh.Sec),
			d(hh.NumCars),
			d(hh.Driver),
			d(hh.FEV),
			d(hh.PHEV),
			d(Age1),
			d(Gender1),
			d(Work1),
			d(Driver1),
			d(Age2),
			d(Gender2),
			d(Work2),
			d(Driver2),
		})

	}
}

func WriteCsvFile(filename string, hhs <-chan *model.Household) {
	f, err := os.Create(filename)
	if err != nil {
		log.Panic(err)
	}
	defer f.Close()
	WriteCsv(f, hhs)
}
