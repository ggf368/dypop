// Package synth contains all code and routines for the generation of the synthetic
// population.
package synth

import (
	"log"
	"math"
	"math/rand"
	"runtime"
	"strconv"
	"time"

	"bitbucket.org/SeheonKim/albatros4/mat"
	"bitbucket.org/SeheonKim/albatros4/model"
	"bitbucket.org/SeheonKim/albatros4/tools"
)

// SynthesizePopulationParams contains the parameters needed by the SynthesizePopulation function
type SynthesizePopulationParams struct {
	IndependentVars  []string
	MonDataFilename  string
	SubZonesFilename string
	ZipCodesFilename string
	LocsNLFilename   string //        *model.LocsNL
	IpfParams        tools.IpfParams
}

// countTable is used to count the different categories for each spatial zone
type countTable struct {
	spatialSegment        int
	ageHouseholdTable     *mat.Mat
	workHouseholdTable    *mat.Mat
	ageWorkHouseholdTable *mat.Mat
	multiwayTable         *mat.Mat
	count                 int
}

// subzoneResult adds the fitted multiwaytable to a subzone
type subzoneResult struct {
	subzone             *Subzone
	fittedMultiwayTable *mat.Mat
}

// max returns the maximum of two floats
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// sum returns the sum of a slice of floats
func sum(vs []float64) (r float64) {
	for _, v := range vs {
		r += v
	}
	return
}

// newCountTable creates an empty collection of counttables for a subzone
// These countables will later be used for the ipf calculations
func newCountTable(spatialSegment int, indepVars []string) *countTable {
	var t countTable
	t.spatialSegment = spatialSegment
	t.ageHouseholdTable = mat.Zeroes(7, 7)
	t.workHouseholdTable = mat.Zeroes(5, 5)
	t.ageWorkHouseholdTable = mat.Zeroes(15, 23)
	dims := []int{15, 23}
	dims = append(dims, VarLevels(indepVars)...)
	t.multiwayTable = mat.Zeroes(dims...)
	t.count = 0
	return &t
}

// countMonData fills all countTables for all spatial segments by reading in the mon data.
func countMonData(filename string, countTables []*countTable, indepVars []string) {
	log.Println("Counting mon data")
	start := time.Now()
	c := ReadMonData(filename)
	for r := range c {
		t := countTables[r.SpatialSegment]
		t.ageHouseholdTable.Inc(r.AgeHousehold.Index()...)
		t.workHouseholdTable.Inc(r.WorkHousehold.Index()...)
		if r.AgeWorkHousehold.U != -1 {
			t.ageWorkHouseholdTable.Inc(r.AgeWorkHousehold.Index()...)
			t.multiwayTable.Inc(r.Index(indepVars)...)
		}
	}
	log.Println("Done counting in", time.Since(start))
}

// Create a fitted multiway table for a subzone.
// The countTable should match the countable for the same spatial segment as the supplied subzone
func createFittedMultiwayTable(subzone *Subzone, countTable *countTable, ipfParams tools.IpfParams) *mat.Mat {
	ageHouseholdTable := countTable.ageHouseholdTable.Clone()
	fagm, _, convergence1 := tools.Ipf(ageHouseholdTable, subzone.AgeHouseholdColTotals(), subzone.AgeHouseholdRowTotals(), ipfParams)
	if convergence1 > ipfParams.ConvLevel {
		log.Fatalf("Error: ipf didn't converge (%f%%) for ageHouseholdTable in subzone %d", convergence1*100, subzone.Id)
	}

	workHouseholdTable := countTable.workHouseholdTable.Clone()
	fwsm, _, convergence2 := tools.Ipf(workHouseholdTable, subzone.WorkHouseholdColTotals(), subzone.WorkHouseholdRowTotals(), ipfParams)
	if convergence2 > ipfParams.ConvLevel {
		log.Fatalf("Error: ipf didn't converge (%f%%) for workHouseholdTable in subzone %d", convergence2*100, subzone.Id)
	}

	rowTotals := make([]float64, len(AgeHouseholdToXY))
	sumRowTotals := 0.0
	for i := range rowTotals {
		rowTotals[i] = fagm.At(AgeHouseholdToXY[i]...)
		sumRowTotals += rowTotals[i]
	}

	colTotals := make([]float64, len(WorkHouseholdToXY))
	sumColTotals := 0.0
	for i := range colTotals {
		colTotals[i] = fwsm.At(WorkHouseholdToXY[i]...)
		sumColTotals += colTotals[i]
	}

	differenceTotalsPercentage := math.Abs(sumColTotals-sumRowTotals) / max(sumColTotals, sumRowTotals)
	if sumColTotals != 0 && sumRowTotals != 0 { // Relaxation criteria for zero Totals (column total or row total)
		if differenceTotalsPercentage > 0.001 {
			log.Fatalf("Difference between totals percentage it too big (%f%%)", differenceTotalsPercentage*100)
		}
	}

	ratio := sumColTotals / sumRowTotals
	for i := range rowTotals {
		rowTotals[i] *= ratio
	}

	ageWorkHouseholdTable := countTable.ageWorkHouseholdTable.Clone()
	fawht, _, convergence3 := tools.Ipf(ageWorkHouseholdTable, colTotals, rowTotals, ipfParams)

	// TODO choose the better convttp" and "log" to your ergence level maybe the column totals or the row totals answer of ipf is better. Add this to the ipf function one day
	if convergence3 > 0.2 {
		log.Printf("The ageWorkHousehold Table in subzone %d didn't converge, there is a %.1f%% difference", subzone.Id, convergence3*100.0)
	}

	// Create the fitted multiway table
	fmwt := countTable.multiwayTable.Clone()
	index := make(mat.Index, len(fmwt.Dims))
	for x := 0; x < fmwt.Dims[0]; x++ {
		for y := 0; y < fmwt.Dims[1]; y++ {
			var ratio float64
			if countTable.ageWorkHouseholdTable.At(x, y) == 0 {
				ratio = 0
			} else {
				ratio = fawht.At(x, y) / countTable.ageWorkHouseholdTable.At(x, y)
			}

			index[0], index[1] = x, y
			for {
				fmwt.Mul(ratio, index...)
				if index.IncFrom(fmwt.Dims, 2) {
					break
				}
			}
		}
	}
	fmwt.InvariantRound()

	return fmwt
}

// createMultiwayTablePerSubzone reads the subzones data and creates a multiway table for each subzone and then
// sends the result on the returned output channel.
func createMultiwayTablePerSubzone(filename string, countTables []*countTable, ipfParams tools.IpfParams) <-chan *subzoneResult {
	input := ReadSubzones(filename)
	outputa := make(chan *subzoneResult, 10)
	wg := 0
	for i := 0; i < runtime.NumCPU()-1; i++ {
		wg++
		go func() {

			for subzone := range input {
				if subzone.Huishoudens > subzone.Bevolking || subzone.Bevolking == 0 || subzone.Huishoudens == 0 {
					log.Printf("Skipping subzone %d because: #houshold > #population or #households = 0 or #population = 0", subzone.Id)
					continue
				} else {
					log.Printf("Processing subzone %d", subzone.Id)
				}

				fmwt := createFittedMultiwayTable(subzone, countTables[subzone.SpatialSegment()], ipfParams)

				outputa <- &subzoneResult{subzone, fmwt}
			}
			outputa <- nil
		}()
	}

	outputb := make(chan *subzoneResult, 10)
	go func() {
		for r := range outputa {
			if r == nil {
				wg--
				if wg == 0 {
					break
				} else {
					continue
				}
			}

			outputb <- r
		}
		close(outputb)
	}()

	return outputb
}

func synthesizePopulationToHouseholds(args SynthesizePopulationParams, c chan *model.Household) {
	defer close(c)

	locsnl := model.ReadLocsNLFile(args.LocsNLFilename)
	zipcode := ReadZipcode(args.ZipCodesFilename)

	// Create count tables
	countTables := make([]*countTable, N_spatial_segment)
	for i := range countTables {
		countTables[i] = newCountTable(i, args.IndependentVars)
	}

	// Do the counting
	countMonData(args.MonDataFilename, countTables, args.IndependentVars)

	subzoneResults := createMultiwayTablePerSubzone(args.SubZonesFilename, countTables, args.IpfParams)
	zipcodePerSubzone := ReadZipcodesPerSubzone(args.ZipCodesFilename)

	// Instead of writeOutput in SynthesizePopulation, the housedhold is constructing from here
	// This loop is similar to func constructHousehold in readmon.go
	hhid := 1

	// Go over all the results from the channel
	for result := range subzoneResults {
		var hh model.Household

		zipcodeSubzone := zipcodePerSubzone[result.subzone.Id]

		if zipcodeSubzone == nil {
			log.Printf("Skipping %d subzone because it does not exist in zipcode file", result.subzone.Id)
			continue
		}

		zipcodeSubzone.SetTotal(int(sum(result.fittedMultiwayTable.Vals)))
		hh.Prov = result.subzone.Prov
		hh.Urb = model.Urb(result.subzone.Sted - 1)

		index := make(mat.Index, len(result.fittedMultiwayTable.Dims))
		x, y := -0xDEADBEEF, -0xDEADBEEF
		for {
			if index[0] != x {
				x = index[0]
				household, work1, gender1, work2, gender2 := RWorkHousehold(x)

				if household == 0 {
					if work1 == 0 {
						hh.Comp = model.CompSingleNoWork
					} else {
						hh.Comp = model.CompSingleWork
					}
				} else if work1 == 0 && work2 == 0 {
					hh.Comp = model.CompDualNoWork
				} else if work1 != 0 && work2 != 0 {
					hh.Comp = model.CompDualTwoWork
				} else {
					hh.Comp = model.CompDualOneWork
				}

				hh.Member = hh.Member[:0] // Remove members from previous iteration
				hh.Member = append(hh.Member, &model.Person{})
				hh.Member[0].ID = 1
				hh.Member[0].Head = true
				hh.Member[0].Gender = model.Gender(gender1)
				hh.Member[0].Work = model.Work(work1)
				if household > 0 {
					hh.Member = append(hh.Member, &model.Person{})
					hh.Member[1].ID = 2
					hh.Member[1].Head = true
					hh.Member[1].Gender = model.Gender(gender2)
					hh.Member[1].Work = model.Work(work2)
				}
			}
			if index[1] != y {
				y = index[1]
				_, age1, _, age2, _ := RAgeHousehold(y)
				hh.Member[0].Age = model.Age(age1)
				hh.MaxAge = hh.Member[0].Age // MaxAge is a maximum age class between household heads
				if len(hh.Member) > 1 {
					hh.Member[1].Age = model.Age(age2)
					hh.MaxAge = model.Age(max(float64(hh.Member[0].Age), float64(hh.Member[1].Age))) // from mon.cpp at line 1101
				}
			}

			// Associate indep var name with value
			for i, name := range args.IndependentVars {
				switch name {
				case "Child":
					hh.Child = model.Child(index[i+2])
				case "NumCars":
					hh.NumCars = int8(min(int(index[i+2]), 2))
				case "Sec":
					hh.Sec = model.Sec(index[i+2])
				case "Drivers":
					num := index[i+2]
					for j := range hh.Member {
						hh.Member[j].IsDriver = j < num
					}
				}
			}
			if hh.Member[0].IsDriver {
				hh.Driver = 1
			}

			count := int(result.fittedMultiwayTable.At(index...))

			for i := 0; i < count; i++ {
				hh.ID = hhid

				if zc, err := strconv.Atoi(zipcodeSubzone.GetRandomZipcode()); err != nil {
					log.Panicln(err)
				} else {
					hh.Home = model.Location(zc)
					if locsnl.Ppc[hh.Home] == nil {
						// hh.WoGem = 999999
						log.Panicln("Home Ppc (", hh.Home, ") is not found in locsnl file! Compare zipcode file with locsnl file!")
					} else {
						hh.WoGem = locsnl.Ppc[hh.Home].Gem
					}
				}

				// Distribute FEV/PHEV by postcodes (Location-based vars...)
				if Binomial(1, float64(zipcode.Ppc[hh.Home].Fev)/float64(zipcode.Ppc[hh.Home].HH)) == 1 {
					hh.FEV = true
				} else {
					hh.FEV = false
				}
				if Binomial(1, float64(zipcode.Ppc[hh.Home].Phev)/float64(zipcode.Ppc[hh.Home].HH)) == 1 {
					hh.PHEV = true
				} else {
					hh.PHEV = false
				}
				c <- hh.Clone()
				hhid++
			}
			if index.Inc(result.fittedMultiwayTable.Dims) {
				break
			}
		}
	}
}

func Binomial(n float64, p float64) int {
	var s float64
	var d float64
	var x int
	var v float64

	if p <= 0.5 {
		s = p / (1 - p)
		d = math.Pow(1-p, n)
		x = 0
	} else {
		s = (1 - p) / p
		d = math.Pow(p, n)
		x = int(n)
	}
	a := (n + 1) * s
	theta := d
	u := rand.Float64()
	for true {
		v = u - theta
		if v <= 0 {
			break
		} else if p <= 0.5 {
			x += 1
			theta *= a/float64(x) - s
		} else {
			x -= 1
			theta *= a/(n-float64(x)) - s
		}
		u = v
	}
	return x
}

func SynthesizePopulationToHouseholds(args SynthesizePopulationParams) <-chan *model.Household {
	// Test correct names for independent variables
	for _, name := range []string{"Child", "Sec", "NumCars", "Drivers"} {
		if _, exists := IndepVarLevels[name]; !exists {
			log.Fatalf("Independent var with name %s is not defined", name)
		}
	}

	c := make(chan *model.Household)
	go synthesizePopulationToHouseholds(args, c)
	return c
}
