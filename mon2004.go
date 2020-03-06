package synth

import (
	"log"
	"reflect"

	"bitbucket.org/SeheonKim/albatros4/model"
)

// IndepVarLevels is a map that defines all allowed independent variables with
// their name and the number of levels.
var IndepVarLevels = map[string]int{
	"NumCars": 4,
	"Drivers": 4,
	"Sec":     4,
	"Child":   4,
	"Day":     7,
}

// MonMember is a classified household member from the mon data
type MonMember struct {
	Hhid             int
	SpatialSegment   int
	Work             int
	NumCars          int
	Age              int
	Household        int
	Child            int
	Drivers          int
	Sec              int
	Day              int
	AgeHousehold     UV
	WorkHousehold    UV
	AgeWorkHousehold UV
}

// UV is a struct that returns x,y coordinates in count tables
type UV struct {
	U int
	V int
}

// UVToUV defines a mapping from UV to UV coordinates
type UVToUV map[UV]UV

// UVToIndex defines a mapping from UV coordinates to an index
type UVToIndex map[UV]int

// IndexToXY defines a mapping from an index to xy (UV like) coordinates
type IndexToXY map[int][]int

// AgeHouseholdRemap remaps age household counts
var AgeHouseholdRemap = UVToUV{
	UV{2, 0}: UV{1, 0},
	UV{3, 0}: UV{2, 1},
	UV{4, 0}: UV{3, 2},
	UV{3, 1}: UV{3, 2},
	UV{4, 1}: UV{3, 2},
	UV{4, 2}: UV{4, 3},
	UV{0, 2}: UV{1, 2},
	UV{0, 3}: UV{1, 2},
	UV{1, 3}: UV{1, 2},
	UV{0, 4}: UV{1, 2},
	UV{1, 4}: UV{2, 3},
	UV{2, 4}: UV{2, 3}}

// AgeHouseholdToIndex convert an age houshold UV coordinate to an index
var AgeHouseholdToIndex = encodeUVToIndex(7,
	10, 11, -1, -1, -1, 0, -1,
	12, 13, 14, -1, -1, 1, -1,
	-1, 15, 16, 17, -1, 2, -1,
	-1, -1, 18, 19, 20, 3, -1,
	-1, -1, -1, 21, 22, 4, -1,
	5, 6, 7, 8, 9, -1, -1,
	-1, -1, -1, -1, -1, -1, -1)

// AgeHouseholdToXY us the inverse of AgeHouseholdToIndex
var AgeHouseholdToXY = encodeIndexToXY(AgeHouseholdToIndex)

// WorkHouseholdToIndex converts an work household UV coordinate to an index
var WorkHouseholdToIndex = encodeUVToIndex(5,
	6, 7, 8, 0, -1,
	9, 10, 11, 1, -1,
	12, 13, 14, 2, -1,
	3, 4, 5, -1, -1,
	-1, -1, -1, -1, -1)

// WorkHouseholdToXY is the inverse of WorkHouseholdToXY
var WorkHouseholdToXY = encodeIndexToXY(WorkHouseholdToIndex)

// encodeUVToIndex is a convinence function to create the map from UVCoordinate
// to an index
func encodeUVToIndex(width int, values ...int) (r UVToIndex) {
	r = make(UVToIndex)
	for i, v := range values {
		r[UV{i % width, i / width}] = v
	}
	return
}

// encodeIndexToXY is a convinence function to create the inverse of the
// encodeUVToIndex result.
func encodeIndexToXY(m UVToIndex) IndexToXY {
	r := make(map[int][]int)
	for k, v := range m {
		if v != -1 {
			r[v] = append(r[v], k.U, k.V)
		}
	}
	return r
}

// Map returns the index for a given UV coordinate
func (uv2i *UVToIndex) Map(uv UV) int {
	i, exists := (*uv2i)[uv]
	if !exists {
		log.Panicln("Key not found in UVToIndex")
	}
	return i
}

// Index returns a int slice for the UV
func (uv *UV) Index() []int {
	return []int{uv.U, uv.V}
}

// RWorkHousehold returns household and work and gender values for given index
func RWorkHousehold(index int) (household, work1, gender1, work2, gender2 int) {
	l := WorkHouseholdToXY[index]
	x, y := l[0], l[1]
	if x < 3 && y < 3 { // Two adult household (Male : 1, Female : 0)
		return 1, x, 1, y, 0
	}
	if x == 3 { // Independent female (Female : 0)
		return 0, y, 0, 999999, 999999
	}
	if y == 3 { // Independent male (Male : 1)
		return 0, x, 1, 999999, 999999
	}

	log.Panicln("Unused reversed classification")
	return
}

// RAgeHousehold returns household age and gender values for given index
func RAgeHousehold(index int) (household, age1, gender1, age2, gender2 int) {
	l := AgeHouseholdToXY[index]
	x, y := l[0], l[1]
	if x < 5 && y < 5 {
		return 1, x, 0, y, 1
	}
	if x == 5 {
		return 0, y, 1, 999999, 999999
	}
	if y == 5 {
		return 0, x, 0, 999999, 999999
	}

	log.Panicln("Unused reversed classification")
	return
}

// VarLevels will using the given independent variable names return an int slice of the
// variable levels
func VarLevels(names []string) (r []int) {
	for _, n := range names {
		r = append(r, IndepVarLevels[n])
	}
	return
}

// Index for a mon member given a index into the multiway table using the
// given independed variable names.
// The multiway table and the call to this function should use the same
// independed variables in the same order.
func (m *MonMember) Index(vars []string) (r []int) {
	r = make([]int, 2+len(vars))
	r[0] = m.AgeWorkHousehold.U
	r[1] = m.AgeWorkHousehold.V

	for i, v := range vars {
		r[i+2] = int(reflect.ValueOf(*m).FieldByName(v).Int())
	}
	return
}

func partners(hh *model.Household) (male, female *model.Person) {
	numHeads := 0
	for _, mem := range hh.Member {
		if mem.Head {
			numHeads++
		}
	}

	if numHeads < 2 {
		log.Panicln("Trying to find two members in a one member household")
	}

	male, female = hh.Member[0], hh.Member[1]
	if male.Gender == 0 && female.Gender == 1 {
		return female, male
	}
	return
}

// min returns the minumum value of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func parseHouseholds(in <-chan *model.Household, out chan *MonMember) {
	for hh := range in {
		// if err := model.CleanData(hh); err != nil { // Only drop household does not meet 7 cleaning criteria // see CleanData()
		// 	log.Printf("Dropping houshold %d: %s\n", hh.ID, err)
		// 	continue
		// }

		numDrivers := 0
		numHeads := 0
		for _, mem := range hh.Member {
			if mem.IsDriver {
				numDrivers++
			}
			if mem.Head {
				numHeads++
			}
		}

		for _, mem := range hh.Member {
			var m MonMember

			m.Hhid = hh.ID
			m.SpatialSegment = C_spatial_segment(int(hh.Urb), hh.Prov)
			m.Work = int(mem.Work)
			m.NumCars = min(int(hh.NumCars), 2)
			m.Age = int(hh.MaxAge)
			if mem.Head {
				m.Household = numHeads - 1
			} else {
				m.Household = 2
			}
			m.Child = int(hh.Child)
			m.Drivers = min(numDrivers, 3)
			m.Sec = int(hh.Sec)

			m.Day = int(hh.Day)

			// AgeHousehold
			if m.Household == 0 { // Independent
				if mem.Gender == model.Male {
					m.AgeHousehold.U, m.AgeHousehold.V = m.Age, 5
				} else {
					m.AgeHousehold.U, m.AgeHousehold.V = 5, m.Age
				}
			} else if m.Household == 1 { // two adult household
				male, female := partners(hh)
				m.AgeHousehold.U, m.AgeHousehold.V = int(male.Age), int(female.Age)
				if p, exists := AgeHouseholdRemap[m.AgeHousehold]; exists {
					m.AgeHousehold = p
				}
			} else {
				if mem.Gender == model.Male { // living in
					m.AgeHousehold.U, m.AgeHousehold.V = m.Age, 6
				} else {
					m.AgeHousehold.U, m.AgeHousehold.V = 6, m.Age
				}
			}

			// WorkHousehold
			if m.Household == 0 { // Independent
				if mem.Gender == model.Male {
					m.WorkHousehold.U, m.WorkHousehold.V = m.Work, 3
				} else {
					m.WorkHousehold.U, m.WorkHousehold.V = 3, m.Work
				}
			} else if m.Household == 1 { // two adult household
				male, female := partners(hh)
				m.WorkHousehold.U, m.WorkHousehold.V = int(male.Work), int(female.Work)
			} else { // living in
				if mem.Gender == model.Male {
					m.WorkHousehold.U, m.WorkHousehold.V = m.Work, 4
				} else {
					m.WorkHousehold.U, m.WorkHousehold.V = 4, m.Work
				}
			}

			// AgeWorkHouseold
			m.AgeWorkHousehold.U = WorkHouseholdToIndex.Map(m.WorkHousehold)
			m.AgeWorkHousehold.V = AgeHouseholdToIndex.Map(m.AgeHousehold)
			if m.AgeWorkHousehold.U == -1 || m.AgeWorkHousehold.V == -1 {
				m.AgeWorkHousehold = UV{-1, -1}
			}

			out <- &m

		}
	}
}

// ReadMonData returns a channel on with MonMembers will be returned.
// You must read all members until the channel is closed.
func ReadMonData(filename string) <-chan *MonMember {
	c := make(chan *MonMember, 10)

	go func() {
		defer close(c)

		hh := model.ReadMonFile(filename) // read Household data from MON data
		parseHouseholds(hh, c)            // convert Household to MonMember
	}()

	return c
}
