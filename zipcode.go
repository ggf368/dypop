package synth

import (
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"

	"bitbucket.org/SeheonKim/albatros4/model"
	"bitbucket.org/SeheonKim/albatros4/tools"
)

// ZipCode is one zipcode including how many zipcodes there are left for this one
type ZipCodeCount struct {
	ZipCodeCount string
	Count        int
}

// ZipCodeGenerator is used to randomly generate zipcodes given the same
// distribution but with a different total number of zipcodes.
type ZipCodeGenerator struct {
	Subzone       int
	ZipCodeCounts []ZipCodeCount
	Total         int
}

// ZipCodePerSubzone contains a ZipCodeGenerator for each subzone based on subzone id.
type ZipCodePerSubzone map[int]*ZipCodeGenerator

// Add adds a zipcode and the number of zipcodes there are to a generator
func (z *ZipCodeGenerator) Add(zipcode string, count int) {
	z.ZipCodeCounts = append(z.ZipCodeCounts, ZipCodeCount{zipcode, count})
	z.Total += count
}

// SetTotal will set the new total for the generator. It will change the counts of the individual
// ZipCodes but keep the same distribution.
func (z *ZipCodeGenerator) SetTotal(total int) {
	if len(z.ZipCodeCounts) == 0 {
		log.Printf("Setting total on zipcode generator not containing any zipcodes for subzone %d", z.Subzone)
		z.Total = 0
		return
	}

	s := 0
	for _, z := range z.ZipCodeCounts {
		s += z.Count
	}

	t := 0
	for i := range z.ZipCodeCounts[:len(z.ZipCodeCounts)-1] {
		z.ZipCodeCounts[i].Count = (z.ZipCodeCounts[i].Count * total) / s
		t += z.ZipCodeCounts[i].Count
	}

	z.ZipCodeCounts[len(z.ZipCodeCounts)-1].Count = total - t
	z.Total = total
}

// GetRandomZipcode generate a random ZipCode. It will also
// decrease the count of that zipcode and the total count.
func (z *ZipCodeGenerator) GetRandomZipcode() string {
	if z.Total == 0 {
		return ""
	}

	r := rand.Int() % z.Total
	t := 0
	for i := range z.ZipCodeCounts {
		t += z.ZipCodeCounts[i].Count
		if r < t {
			z.ZipCodeCounts[i].Count--
			z.Total--
			return z.ZipCodeCounts[i].ZipCodeCount
		}
	}
	return ""
}

type ZipCodeRecord struct {
	Subzone int
	Ppc     string
	Hh      int
	Fev     int
	Phev    int
}

// ReadZipcodesPerZubzone loads a file with information about how many
// zipcodes there or per subzone and returns a zipcodegenerator per subzone.
func ReadZipcodesPerSubzone(filename string) (m ZipCodePerSubzone) {
	f, err := os.Open(filename)
	if err != nil {
		log.Panicln(err)
	}
	defer f.Close()

	m = make(ZipCodePerSubzone)

	csv := tools.NewCsvReader(f, '\t')
	for {
		r := new(ZipCodeRecord)
		err := csv.Read(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Panicln(err)
		}

		if _, exists := m[r.Subzone]; !exists {
			m[r.Subzone] = &ZipCodeGenerator{Subzone: r.Subzone}
		}
		m[r.Subzone].Add(r.Ppc, r.Hh)
	}

	return m
}

type ZipCode struct {
	Ppc map[model.Location]*ZipCodeInfo
}
type ZipCodeInfo struct {
	Ppc model.Location // The zipcode
	HH  int
	Fev  int
	Phev int
}

func ReadZipcode(filename string) *ZipCode {
	reader, err := os.Open(filename)
	if err != nil {
		log.Panicln("Error opening zipcode file: ", err)
	}
	defer reader.Close()

	zipcode := newZipcode()
	csv := tools.NewCsvReader(reader, '\t')
	for {
		r := new(ZipCodeRecord)
		err := csv.Read(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Panicln(err)
		}
		ppc, _ := strconv.Atoi(r.Ppc)

		if _, exists := zipcode.Ppc[model.Location(ppc)]; exists {
			log.Panicln("Double zipcode", r.Ppc, "found")
		}

		zipcodeInfo := new(ZipCodeInfo)
		zipcodeInfo.Ppc = model.Location(ppc)
		zipcodeInfo.HH = r.Hh
		zipcodeInfo.Fev = r.Fev
		zipcodeInfo.Phev = r.Phev

		zipcode.Ppc[model.Location(ppc)] = zipcodeInfo
	}
	return zipcode
}

func newZipcode() *ZipCode {
	z := new(ZipCode)
	z.Ppc = make(map[model.Location]*ZipCodeInfo)
	return z
}
