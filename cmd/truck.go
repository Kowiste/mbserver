package main

import (
	"encoding/binary"
	"strconv"

	"github.com/siemens/go/src/math"
)

//Constant values
const (
	PercentActive byte = 3
)

var geoMap map[string][]float64

//GTruck Group of truck
var GTruck GroupTrucks

//GroupTrucks group truck
type GroupTrucks struct {
	Trucks map[string]*Truck
}

//Truck truck
type Truck struct {
	Max      int
	Next     int
	GPS      []float64
	ActPump  int
	pumps    map[byte]*Pump
	TimesAct int
	Moving   bool
}

//Pump pump struct
type Pump struct {
	StartStop    bool
	Voltage      Volts
	Current      Current
	Power        float32
	Flow         Flow
	TotalVol     float32
	PercentNoise uint16
}
type Volts struct {
	Voltage   uint16
	VoltNom   uint16
	VoltNoise int
}
type Current struct {
	Current      float32
	CurrentNom   float32
	CurrentNoise float32
	RaisePerc    float32
}
type Flow struct {
	FlowRate    float32
	FlowRateNom float32
	FlowNoise   float32
	RaisePerc   float32
}

func (t *GroupTrucks) updateGeo() {
	for truck := range t.Trucks {
		retMem := doubleToByte(geoMap[truck], t.Trucks[truck].Next) //Get the next 2 float for every truck
		if t.Trucks[truck].Next < t.Trucks[truck].Max-2 {           //Check for the limit
			t.Trucks[truck].Next += 2 //Pointer to the next element 2 position
		} else {
			t.Trucks[truck].Next = 0
		}
		num, _ := strconv.Atoi(truck)
		for element := range retMem {
			memory[num+element] = retMem[element]
		}
		println("")
	}
}
func (t *GroupTrucks) updatePump() {
	for truck := range t.Trucks {
		t.Trucks[truck].updatePump()
		num, _ := strconv.Atoi(truck)
		retMem := t.Pump2uint16()
		for element := range retMem {
			memory[num+element+8] = retMem[element] // 8 bc the geolocation
		}
	}
}

//NewPump create pump
func NewPump(NominalVolts uint16, NominalCurrent float32, NominalFlow float32, NoisePercent int, NoiseFlow float32) *Pump {
	return &Pump{
		Voltage:      Volts{VoltNom: NominalVolts, VoltNoise: int(NoisePercent)},
		Current:      Current{CurrentNom: NominalCurrent, CurrentNoise: (NoiseFlow), RaisePerc: 10},
		Flow:         Flow{FlowRateNom: NominalFlow, FlowNoise: NoiseFlow, RaisePerc: 10},
		PercentNoise: uint16(NoisePercent),
	}
}
func (t *Truck) updatePump() {
	if t.ActPump == 0 {
		n := byte(rd.Intn(100)) //random start
		if n < PercentActive {
			n = 100
			t.ActPump++
			for element := range t.pumps {
				t.pumps[element].StartStop = true
			}
		}
	} else if t.ActPump == NTimesActive { //Finish active
		t.ActPump = 0
		for element := range t.pumps {
			t.pumps[element].StartStop = false
		}
	} else {
		t.ActPump++
	}
	t.updateVoltage()
	t.updateCurrent()
	t.updateFlow()
	println("N: ", t.ActPump, ", Act: ", t.pumps[0].StartStop, ", Amps: ", int(t.pumps[0].Current.Current), ", Volts: ", int(t.pumps[0].Voltage.Voltage), ", Flow: ", int(t.pumps[0].Flow.FlowRate), ", Total: ", int(t.pumps[0].TotalVol))
}
func (t *Truck) updateVoltage() {
	for pump := range t.pumps {
		if t.ActPump != 0 {
			//Nominal value - Volts Nominal/2 + rand(VoltsNom)
			t.pumps[pump].Voltage.Voltage = t.pumps[pump].Voltage.VoltNom - uint16(randomInt(t.pumps[pump].Voltage.VoltNoise))
		} else {
			t.pumps[pump].Voltage.Voltage = 0
		}
	}
}
func (t *Truck) updateCurrent() {

	for pump := range t.pumps {
		//v := float32(t.TimesAct) * (t.pumps[pump].Current.RaisePerc / 100)
		if t.ActPump == 0 {
			t.pumps[pump].Current.Current = 0
			/*} else if t.ActPump > (t.TimesAct - int(v)) { //We are at the end of the signal
			//de := float32(math.Pow(float64(float32(t.ActPump)-v), 2))
			//t.pumps[pump].Current.Current = t.pumps[pump].Current.CurrentNom - 0.25*de
			t.pumps[pump].Current.Current = 0*/
		} else {
			anterior := t.pumps[pump].Current.Current
			nominal := t.pumps[pump].Current.CurrentNom
			t.pumps[pump].Current.Current = anterior + (nominal-anterior)/(float32(t.TimesAct)*(t.pumps[pump].Current.RaisePerc/100))
		}
	}
}
func (t *Truck) updateFlow() {
	for pump := range t.pumps {
		v := float32(t.TimesAct) * (t.pumps[pump].Current.RaisePerc / 100)
		if t.ActPump != 0 {
			t.pumps[pump].Flow.FlowRate = t.pumps[pump].Flow.FlowRate + (t.pumps[pump].Flow.FlowRateNom-t.pumps[pump].Flow.FlowRate)/(float32(t.TimesAct)*(t.pumps[pump].Flow.RaisePerc/100))
		} else {
			if t.pumps[pump].Flow.FlowRate > 0 {
				de := float32(math.Pow(float64(float32(t.ActPump)-v), 2))
				t.pumps[pump].Flow.FlowRate = t.pumps[pump].Flow.FlowRate - 0.5*de
			} else {
				t.pumps[pump].Flow.FlowRate = 0
			}
		}
		t.pumps[pump].TotalVol += t.pumps[pump].Flow.FlowRate
	}
}

func randomInt(level int) int {
	return rd.Intn(level) - level/2
}

//Pump2uint16 pump
func (t *GroupTrucks) Pump2uint16() []uint16 {
	out := make([]uint16, 0)
	for truck := range t.Trucks {
		for pump := range t.Trucks[truck].pumps {
			if t.Trucks[truck].pumps[pump].StartStop {
				out = append(out, 1)
			} else {
				out = append(out, 0)
			}
			out = append(out, float2uint16(t.Trucks[truck].pumps[pump].Current.Current)...)                                                      //current
			out = append(out, t.Trucks[truck].pumps[pump].Voltage.Voltage)                                                                       //voltage
			out = append(out, float2uint16(t.Trucks[truck].pumps[pump].Current.Current*float32(t.Trucks[truck].pumps[pump].Voltage.Voltage))...) //power consumption
			out = append(out, float2uint16(t.Trucks[truck].pumps[pump].Flow.FlowRate)...)
			out = append(out, float2uint16(t.Trucks[truck].pumps[pump].TotalVol)...)
		}
	}
	return out
}

func float2uint16(input float32) []uint16 {
	v := make([]byte, 8)
	binary.BigEndian.PutUint32(v, math.Float32bits(input))
	return []uint16{binary.BigEndian.Uint16(v[:2]), binary.BigEndian.Uint16(v[2:])}
}
