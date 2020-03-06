package synth

import (
	"io"
	"log"
	// "math/rand"
	"os"

	"bitbucket.org/SeheonKim/albatros4/model"
	"bitbucket.org/SeheonKim/albatros4/tools"
)

type SynthData struct {
	// Mandatory attributes to construct a household
	HHID    int
	Home    int
	Gem     int
	Urb     int
	Comp    int
	Child   int
	Day     int
	SEC     int
	Ncar    int
	Driver  int
	Age1    int
	Gender1 int
	Work1   int
	Driver1 int
	Age2    int
	Gender2 int
	Work2   int
	Driver2 int

	// Optional attributes
	EV   int
	FEV  int
	PHEV int
}

func ReadSynthFile(filename string) <-chan *model.Household {
	file, err := os.Open(filename)
	if err != nil {
		log.Panicln("Error reading synth file:", err)
	}

	hhs := make(chan *model.Household)

	go func() {
		defer file.Close()
		defer close(hhs)

		csv := tools.NewCsvReader(file, ',')
		for {
			record := new(SynthData)
			err := csv.Read(record)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Panic(err)
			}

			hh := model.NewHousehold()
			hh.ID = record.HHID
			hh.Home = model.Location(record.Home)
			hh.WoGem = record.Gem
			hh.Urb = model.Urb(record.Urb)
			hh.Comp = model.Comp(record.Comp)
			hh.Child = model.Child(record.Child)
			hh.Day = model.Day(record.Day)
			hh.Sec = model.Sec(record.SEC)
			hh.NumCars = int8(record.Ncar)
			hh.Driver = record.Driver

			// Ownership Electric Vehicle
			// Extension
			hh.EV = record.EV == 1
			hh.FEV = record.FEV == 1
			hh.PHEV = record.PHEV == 1

			// pFEV := float64(0.16)
			// if float64(rand.Intn(100)/100) > pFEV {
			// 	hh.FEV = true
			// } else {
			// 	hh.PHEV = true
			// }
			
			if record.Age1 != 999999 {
				mem1 := model.NewPerson()
				mem1.ID = len(hh.Member) + 1
				mem1.Head = true
				mem1.Gender = model.Gender(record.Gender1)
				mem1.Age = model.Age(record.Age1)
				hh.MaxAge = mem1.Age
				if record.Driver1 == 1 {
					mem1.IsDriver = true
				}
				mem1.Work = model.Work(record.Work1)

				hh.Member = append(hh.Member, mem1)
			} else {
				log.Panicln("Household ", hh.ID, " does not have any household member")
			}

			if record.Age2 != 999999 {
				mem2 := model.NewPerson()
				mem2.ID = len(hh.Member) + 1
				mem2.Head = true
				mem2.Gender = model.Gender(record.Gender2)
				mem2.Age = model.Age(record.Age2)
				if mem2.Age > hh.MaxAge {
					hh.MaxAge = mem2.Age
				}
				if record.Driver2 == 1 {
					mem2.IsDriver = true
				}
				mem2.Work = model.Work(record.Work2)

				hh.Member = append(hh.Member, mem2)
			}

			// Fill up hh
			hhs <- hh
		}
	}()

	return hhs
}
