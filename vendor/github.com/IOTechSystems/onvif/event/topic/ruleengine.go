package topic

import "github.com/IOTechSystems/onvif/xsd"

type RuleEngine struct {
	Topic                *xsd.Boolean          `xml:"topic,attr"`
	MotionRegionDetector *MotionRegionDetector `json:",omitempty" xml:",omitempty"`
	CellMotionDetector   *CellMotionDetector   `json:",omitempty" xml:",omitempty"`
	TamperDetector       *TamperDetector       `json:",omitempty" xml:",omitempty"`
	Recognition          *Recognition          `json:",omitempty" xml:",omitempty"`
	CountAggregation     *CountAggregation     `json:",omitempty" xml:",omitempty"`
}

type MotionRegionDetector struct {
	Topic  *xsd.Boolean `xml:"topic,attr"`
	Motion *Motion      `json:"Motion" xml:"Motion"`
}

type CellMotionDetector struct {
	Topic  *xsd.Boolean `xml:"topic,attr"`
	Motion *Motion
}

type Motion struct {
	Topic              *xsd.Boolean        `xml:"topic,attr"`
	MessageDescription *MessageDescription `json:",omitempty" xml:",omitempty"`
}

type TamperDetector struct {
	Topic  *xsd.Boolean `xml:"topic,attr"`
	Tamper *Tamper
}

type Tamper struct {
	Topic              *xsd.Boolean        `xml:"topic,attr"`
	MessageDescription *MessageDescription `json:",omitempty" xml:",omitempty"`
}

type Recognition struct {
	Topic        *xsd.Boolean `xml:"topic,attr"`
	Face         *Face        `json:",omitempty" xml:",omitempty"`
	LicensePlate *Face        `json:",omitempty" xml:",omitempty"`
}

type Face struct {
	Topic              *xsd.Boolean        `xml:"topic,attr"`
	MessageDescription *MessageDescription `json:",omitempty" xml:",omitempty"`
}

type LicensePlate struct {
	Topic              *xsd.Boolean        `xml:"topic,attr"`
	MessageDescription *MessageDescription `json:",omitempty" xml:",omitempty"`
}

type CountAggregation struct {
	Topic   *xsd.Boolean `xml:"topic,attr"`
	Counter *Counter     `json:",omitempty" xml:",omitempty"`
}

type Counter struct {
	Topic              *xsd.Boolean        `xml:"topic,attr"`
	MessageDescription *MessageDescription `json:",omitempty" xml:",omitempty"`
}
