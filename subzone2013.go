package synth

import (
	"io"
	"log"
	"os"
	"runtime"

	"bitbucket.org/SeheonKim/albatros4/tools"
)

// Subzone defines the data needed for a subzone
type Subzone struct {
	Id                     int `csv:"subzone"`
	Prov                   int
	Sted                   int
	Bevolking              int
	Huishoudens            int
	AgeHouseholdColTotal1  float64 `csv:"ma 0-34"`
	AgeHouseholdColTotal2  float64 `csv:"ma 35-54"`
	AgeHouseholdColTotal3  float64 `csv:"ma 55-64"`
	AgeHouseholdColTotal4  float64 `csv:"ma 65-74"`
	AgeHouseholdColTotal5  float64 `csv:"ma 75+"`
	AgeHouseholdColTotal6  float64 `csv:"vr ind"`
	AgeHouseholdColTotal7  float64 `csv:"vr liv"`
	AgeHouseholdRowTotal1  float64 `csv:"vr 0-34"`
	AgeHouseholdRowTotal2  float64 `csv:"vr 35-54"`
	AgeHouseholdRowTotal3  float64 `csv:"vr 55-64"`
	AgeHouseholdRowTotal4  float64 `csv:"vr 65-74"`
	AgeHouseholdRowTotal5  float64 `csv:"vr 75+"`
	AgeHouseholdRowTotal6  float64 `csv:"ma ind"`
	AgeHouseholdRowTotal7  float64 `csv:"ma liv"`
	WorkHouseholdColTotal1 float64 `csv:"ma nt"`
	WorkHouseholdColTotal2 float64 `csv:"ma pt"`
	WorkHouseholdColTotal3 float64 `csv:"ma ft"`
	WorkHouseholdColTotal4 float64 `csv:"vr ind"`
	WorkHouseholdColTotal5 float64 `csv:"vr liv"`
	WorkHouseholdRowTotal1 float64 `csv:"vr nt"`
	WorkHouseholdRowTotal2 float64 `csv:"vr pt"`
	WorkHouseholdRowTotal3 float64 `csv:"vr ft"`
	WorkHouseholdRowTotal4 float64 `csv:"ma ind"`
	WorkHouseholdRowTotal5 float64 `csv:"ma liv"`
}

func (s *Subzone) SpatialSegment() int {
	return C_spatial_segment(s.Sted, s.Prov)
}

func (s *Subzone) AgeHouseholdColTotals() []float64 {
	return []float64{s.AgeHouseholdColTotal1, s.AgeHouseholdColTotal2, s.AgeHouseholdColTotal3, s.AgeHouseholdColTotal4, s.AgeHouseholdColTotal5, s.AgeHouseholdColTotal6, s.AgeHouseholdColTotal7}
}

func (s *Subzone) AgeHouseholdRowTotals() []float64 {
	return []float64{s.AgeHouseholdRowTotal1, s.AgeHouseholdRowTotal2, s.AgeHouseholdRowTotal3, s.AgeHouseholdRowTotal4, s.AgeHouseholdRowTotal5, s.AgeHouseholdRowTotal6, s.AgeHouseholdRowTotal7}
}

func (s *Subzone) WorkHouseholdColTotals() []float64 {
	return []float64{s.WorkHouseholdColTotal1, s.WorkHouseholdColTotal2, s.WorkHouseholdColTotal3, s.WorkHouseholdColTotal4, s.WorkHouseholdColTotal5}
}

func (s *Subzone) WorkHouseholdRowTotals() []float64 {
	return []float64{s.WorkHouseholdRowTotal1, s.WorkHouseholdRowTotal2, s.WorkHouseholdRowTotal3, s.WorkHouseholdRowTotal4, s.WorkHouseholdRowTotal5}
}

// N_spatial_segment defines the number of spatial segment levels
var N_spatial_segment = 5

// C_spatial_segment classifies a spatial segment using the given sted and prov values. // from Albatross book 2.0 at p.38
// Sted in MON 2004: 1 - 5 (changed to) -> // hh.Urb: 0 - 4
func C_spatial_segment(sted, prov int) int {
	switch {
	case sted == 0:
		return 0
	case sted == 1:
		return 1
	case sted == 2:
		return 2
	case prov == 4 || prov == 9 || prov == 11 || prov == 12:
		return 4
	default:
		return 3
	}
}

func readSubzones(filename string, c chan *Subzone) {
	file, err := os.Open(filename)
	if err != nil {
		log.Panicln(err)
	}
	defer file.Close()

	csv := tools.NewCsvReader(file, '\t')
	for {
		ss := new(Subzone)
		err := csv.Read(ss)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Panicln(err)
		}

		c <- ss
	}
}

// ReadSubzones returns a channel on witch the subzones will be returned.
// You must read all members until the channel is closed.
func ReadSubzones(filename string) <-chan *Subzone {
	c := make(chan *Subzone, runtime.NumCPU()*5)
	go func() {
		defer close(c)
		readSubzones(filename, c)
	}()
	return c
}
