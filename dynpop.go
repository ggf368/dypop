package main

import (
	//"os"
	//"log"
	//"bitbucket.org/SeheonKim/albatros4/tools"
	//"io"
	//"bitbucket.org/SeheonKim/albatros4/synth"
	"fmt"
	"math/rand"

)
type IndVar []int

type HhVar  []int

type extraIndVar struct{
	varIndex	int
	name		string
	value		[]int
}

type extraHhVar struct{
	varIndex	int
	name		string
	value		[]int
}

type IndEvent struct {
	IndEventName string
	IndEventPrb  []float64
	IndEventIns  []int
	IndEventChg  func(*DynInd)
}

type HhEvent struct {
	HhEventName	string
	HhEventPrb	[]float64
	HhEventIns	[]int
	HhEventChg	func(*DynHh)
}

type IndVars   []*extraIndVar

type HhVars    []*extraHhVar

type IndEvents []*IndEvent

type HhEvents  []*HhEvent

type DynHh struct{
	HhId		int
	Cars		HhVar
	Prov		HhVar
	Sted		HhVar
	Subzone		HhVar
	Pc4		HhVar
	Sec		HhVar
	Drivers		HhVar
	DivorceEvent	HhEvent
	MarryEvent	HhEvent
	BirthEvent	HhEvent

	HhVars		HhVars
	HhEvents 	HhEvents
	Members		[]*DynInd
}

type DynInd struct {
	IndId		int
	Gender		IndVar
	Age		IndVar
	RAge		IndVar
	Work		IndVar
	Driver		IndVar
	IndVars		IndVars
	DeathEvent	IndEvent
	ChildLeaveEvent	IndEvent
	WorkChangeEvent IndEvent
	LicenseEvent	IndEvent
	IndEvents	IndEvents

}

type StHh struct {
	Hhid            int
	Gender1         int
	Work1           int
	Age1            int
	Driver1         int
	Gender2         int
	Work2           int
	Age2            int
	Driver2         int
	Age             int
	Comp            int
	Prov            int
	Sted            int
	Subzone         int
	Pc4             int
	Child           int
	Sec             int
	Num_cars        int
	Drivers         int
}

//a sample of static household
var exphh = StHh{
	Hhid:            619107,
	Gender1:         0,
	Work1:           2,
	Age1  :          1,
	Driver1:         0,
	Gender2 :        1,
	Work2 :	         0,
	Age2 :           0,
	Driver2 :        0,
	Age   :          1,
	Comp  :          3,
	Prov  :          3,
	Sted   :         4,
	Subzone  :       124,
	Pc4     :        7823,
	Child  :         2,
	Sec     :        3,
	Num_cars  :      1,
	Drivers  :       1}

var divorceHhs []*DynHh
var numHh int

//add and initialize other household attributes
func (hh *DynHh) NewHhVars(names []string,initValue []int) {

	for i,v :=range names{
		c:=new(extraHhVar)
		c.varIndex=i
		c.name=v
		c.value=[]int{initValue[i]}
		hh.HhVars=append(hh.HhVars,c)
	}
	return
}

//add and initialize other individual attributes
func (mem *DynInd) NewIndVars(names []string,initValue []int) {

	for i,v :=range names{
		c:=new(extraIndVar)
		c.varIndex=i
		c.name=v
		c.value=[]int{initValue[i]}
		mem.IndVars=append(mem.IndVars,c)
	}

	return
}

//TODO only the youngest child now
//transfer a static household to a dynamic household
func StHhToHh( shh StHh) *DynHh{
	c:=new(DynHh)
	//set value for household
	c.HhId=shh.Hhid
	c.Cars=[]int{shh.Num_cars}
	c.Prov=[]int{shh.Prov}
	c.Sted=[]int{shh.Sted}
	c.Subzone=[]int{shh.Subzone}
	c.Pc4=[]int{shh.Pc4}
	c.Sec=[]int{shh.Sec}
	c.Drivers=[]int{shh.Drivers}
	//set value for individual
	ind:=setValueInd(shh.Age1,shh.Gender1,shh.Work1,shh.Driver1)

	c.Members=append(c.Members,ind)

	if secondAdult(shh){
		secondInd:=setValueInd(shh.Age2,shh.Gender2,shh.Work2,shh.Driver2)
		c.Members=append(c.Members,secondInd)
	}

	if shh.Child>0 {
		child := setValueChild(shh.Child)
		c.Members=append(c.Members,child)
	}
	return c
}

//rand real age according to the category of age
func randomAge(age int)int{

	switch {
	case age == 0 :
		return 18 + rand.Intn(18)
	case age == 1 :
		return 36 + rand.Intn(19)
	case age == 2 :
		return 55 + rand.Intn(10)
	case age == 3 :
		return 64 + rand.Intn(10)
	default:
		return 75 + rand.Intn(20)
	}
}

func ageLevel(rage int)int{
	switch{
	case rage>=18 && rage<36:
		return 0
	case rage>=36 && rage<55:
		return 1
	case rage>=55 && rage<64:
		return 2
	case rage>=64 && rage<75:
		return 3
	default:
		return 4
	}

}

//rand real age of child according to the category of child age
func randomChildAge(childAge int)(rcage int){
	switch {
	case childAge == 1 :
		rcage = rand.Intn(6)
	case childAge == 2 :
		rcage =  6 + rand.Intn(6)
	case childAge == 3 :
		rcage =  12 + rand.Intn(6)
	default:
		rcage=0
	}
	return rcage
}

//secondAdult judge if there is second adult
func secondAdult(shh StHh)bool{
	if shh.Age2 < 999999{
		return true
	}else{
		return false
	}
}

//set attribute of adult
func setValueInd(age,gender,work,driver int )*DynInd{
	ind:=new(DynInd)
	ind.Age=[]int{age}
	ind.Gender=[]int{gender}
	ind.Work=[]int{work}
	ind.Driver=[]int{driver}
	ind.RAge=[]int{ randomAge(age) }
	return ind
}

//set attributes of child
func setValueChild(age int )*DynInd{
	ind:=new(DynInd)
	ind.Age=[]int{age}
	ind.Gender=[]int{rand.Intn(1)}
	ind.Work=[]int{0}
	ind.Driver=[]int{0}
	ind.RAge=[]int{randomChildAge(age)}
	return ind
}



//set value of individual events
func(hh *DynHh)eventDeathHh(){
	for i,v:=range hh.Members{
		v.eventDeath(prbDeath(v))
		//todo need to check
		if v.DeathEvent.IndEventIns[len(v.DeathEvent.IndEventIns)-1]==1{
			hh.Members = append(hh.Members[:i],hh.Members[i+1:]...)
		}
	}
}

func (mem *DynInd) eventDeath(prb float64){
	mem.DeathEvent.IndEventName="death"
	mem.DeathEvent.IndEventPrb=append(mem.DeathEvent.IndEventPrb,prb)
	ins:=MonteCarlo([]float64{1-prb,prb})
	mem.DeathEvent.IndEventIns=append(mem.DeathEvent.IndEventIns,ins)
	return
}
func prbDeath(mem *DynInd)float64{
		switch mem.Age[len(mem.Age)-1] {
		case 0:
			return 0.1
		case 1:
			return 0.2
		case 2:
			return 0.3
		case 3:
			return 0.4
		default:
			return 0.4
	}

}

//set value of individual events
func(hh *DynHh)eventLeaveHh(){
	for i,v:=range hh.Members{
		v.eventLeave(prbLeave(v))
		//todo need to check
		if v.ChildLeaveEvent.IndEventIns[len(v.ChildLeaveEvent.IndEventIns)-1]==1{
			hh.Members=append(hh.Members[:i],hh.Members[i+1:]...)
		}
	}
}

func (mem *DynInd) eventLeave(prb float64){
	mem.ChildLeaveEvent.IndEventName="childleave"
	mem.ChildLeaveEvent.IndEventPrb=append(mem.ChildLeaveEvent.IndEventPrb,prb)
	ins:=MonteCarlo([]float64{1-prb,prb})
	mem.ChildLeaveEvent.IndEventIns=append(mem.ChildLeaveEvent.IndEventIns,ins)
	return
}
func prbLeave(mem *DynInd)float64{
		switch mem.RAge[len(mem.RAge)-1] {
		case 18:
			return 0.1
		case 19:
			return 0.2
		case 20:
			return 0.3
		case 21:
			return 0.4
		case 22:
			return 0.4
		default:
			return 0
	}
}


func(hh *DynHh)eventJobHh(){
	for _,v:=range hh.Members{
		v.eventJob(prbJob(v),jobChange)
	}
}

//todo jobtype were not considered
func (mem *DynInd)eventJob(prb float64,fw func(*DynInd)){
	mem.WorkChangeEvent.IndEventName="job"
	mem.WorkChangeEvent.IndEventPrb=append(mem.WorkChangeEvent.IndEventPrb,prb)
	ins:=MonteCarlo([]float64{1-prb,prb})
	mem.WorkChangeEvent.IndEventIns=append(mem.WorkChangeEvent.IndEventIns,ins)
	mem.WorkChangeEvent.IndEventChg=fw
		if ins==1{fw(mem)}

	return
}

func jobChange(mem *DynInd){

	switch mem.Work[len(mem.Work)-1] {
	case 0:
		mem.Work[len(mem.Work)-1]=1
	case 1:
		mem.Work[len(mem.Work)-1]=0
	case 2:
		mem.Work[len(mem.Work)-1]=0
	}
	return
}

func prbJob(mem *DynInd)float64{
	switch mem.Work[len(mem.Work)-1] {
		case 0:
			return 0.8
		case 1:
			return 0.1
		case 2:
			return 0.2
		default:
			return 0
	}

}

func(hh *DynHh)eventLicenseHh(){
	for _,v:=range hh.Members{
		v.eventLicense(prbLicense(v),licenseChange)
	}
}

//change of license
func (mem *DynInd)eventLicense(prb float64,fl func(*DynInd)){
	mem.LicenseEvent.IndEventName="driver"
	mem.LicenseEvent.IndEventPrb=append(mem.LicenseEvent.IndEventPrb,prb)
	ins:=MonteCarlo([]float64{1-prb,prb})
	mem.LicenseEvent.IndEventIns=append(mem.LicenseEvent.IndEventIns,ins)
	mem.LicenseEvent.IndEventChg=fl
		if ins==1{fl(mem)}

	return
}

func licenseChange(mem *DynInd){

	if mem.Driver[len(mem.Driver)-1]==0 {
		mem.Driver[len(mem.Driver)-1]=1
	}
	return
}

func prbLicense(mem *DynInd)float64{
	if mem.Driver[len(mem.Driver)-1]==0 {
		return 0.8
	}else {
		return 0
	}

}



//set value of household birth events
func (hh *DynHh) eventBirth(prb float64, fb func(*DynHh)) {
	hh.BirthEvent.HhEventName="birth"
	hh.BirthEvent.HhEventPrb=append(hh.BirthEvent.HhEventPrb,prb)
	ins:=MonteCarlo([]float64{1-prb,prb})
	hh.BirthEvent.HhEventIns=append(hh.BirthEvent.HhEventIns,ins)
	hh.BirthEvent.HhEventChg=fb

	if ins==1{fb(hh)}

	return
}

func birthChange(hh *DynHh){
	bb:=setValueChild(0)
	hh.Members = append(hh.Members,bb)
}

//todo the first person and the second person are always couple
func prbBirth(hh *DynHh)float64{
	switch {
		case len(hh.Members)>1:
		if hh.Members[0].RAge[len(hh.Members[0].RAge)-1]>18 && hh.Members[1].RAge[len(hh.Members[1].RAge)-1]>18 &&
			hh.Members[0].Gender[len(hh.Members[0].Gender)-1]!=hh.Members[1].Gender[len(hh.Members[0].Gender)-1]{
			return 0.8
		}else{
			return 0.0}
		case len(hh.Members)<=1:
			return 0.0
		default:
			return 0.0
	}

}

//set value of household divorce events
func (hh *DynHh) eventDivorce(prb float64, fd func(*DynHh)) {
	hh.DivorceEvent.HhEventName="divorce"
	hh.DivorceEvent.HhEventPrb=append(hh.DivorceEvent.HhEventPrb,prb)
	ins:=MonteCarlo([]float64{1-prb,prb})
	hh.DivorceEvent.HhEventIns=append(hh.DivorceEvent.HhEventIns,ins)
	hh.DivorceEvent.HhEventChg=fd

	if ins==1{
		fd(hh)
	}

	return
}
//todo always the first person leave

func divorceChange(hh *DynHh){
	temp:=hh
	copy(hh.Members[0:],hh.Members[1:])
	hh.Members[len(hh.Members)-1]=nil
	hh.Members=hh.Members[:len(hh.Members)-1]
	hh.HhId=numHh
	numHh++
	nhh:=*temp
	nhh.HhId=numHh
	numHh++
	nhh.Members=temp.Members[:1]
	divorceHhs=append(divorceHhs,&nhh)
}


func prbDivorce(hh *DynHh)float64{
	switch {
		case len(hh.Members)>1:
			return 0.8
		case len(hh.Members)<=1:
			return 0.0
		default:
			return 0.0
	}

}


//yearUpdate append the attributes of new year to the slice
func (hh *DynHh)yearUpdate()*DynHh{
	hh.Cars=append(hh.Cars,hh.Cars[len(hh.Cars)-1])
	hh.Prov=append(hh.Prov,hh.Prov[len(hh.Prov)-1])
	hh.Sted=append(hh.Sted,hh.Sted[len(hh.Sted)-1])
	hh.Subzone=append(hh.Subzone,hh.Subzone[len(hh.Subzone)-1])
	hh.Pc4=append(hh.Pc4,hh.Pc4[len(hh.Pc4)-1])
	hh.Sec=append(hh.Sec,hh.Sec[len(hh.Sec)-1])
	hh.Drivers=append(hh.Drivers,hh.Drivers[len(hh.Drivers)-1])
	for _,v:=range hh.Members{
		v.RAge=append(v.RAge,v.RAge[len(v.RAge)-1]+1)
		v.Age=append(v.Age,ageLevel(v.RAge[len(v.Age)-1]))
		v.Gender=append(v.Gender,v.Gender[len(v.Gender)-1])
		v.Work=append(v.Work,v.Work[len(v.Work)-1])
		v.Driver=append(v.Driver,v.Driver[len(v.Driver)-1])
	}

	hh.eventDeathHh()

	hh.eventJobHh()

	hh.eventLeaveHh()

	hh.eventLicenseHh()

	hh.eventBirth(prbBirth(hh),birthChange)

	hh.eventDivorce(prbDivorce(hh),divorceChange)

	return hh

}

func MonteCarlo(probs []float64) int {
	sum := 0.0
	for _, p := range probs {
		sum += p
	}

	r := rand.Float64() * sum

	v := 0.0
	for i, p := range probs[:len(probs)-1] {
		v += p
		if r < v {
			return i
		}
	}

	return len(probs) - 1
}


func main() {
  	dhh:=new(DynHh)

	dhh=StHhToHh(exphh)

	dhh.eventDeathHh()

	dhh.eventJobHh()

	dhh.eventLeaveHh()

	dhh.eventLicenseHh()

	dhh.eventBirth(prbBirth(dhh),birthChange)

	dhh.eventDivorce(prbDivorce(dhh),divorceChange)
	dhh.NewHhVars([]string{"housetype","cartype"},[]int{3,5})

	for i:=0;i<6;i++{
		dhh=dhh.yearUpdate()
	}


     	for _,v:=range dhh.HhVars{
	fmt.Println(*v)
	}

     	for _,c:=range dhh.Members{
	fmt.Println(*c)
	}

fmt.Println(*dhh)
	fmt.Println(dhh.BirthEvent.HhEventIns)
}

