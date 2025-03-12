package topic

import "github.com/IOTechSystems/onvif/xsd"

type MessageDescription struct {
	IsProperty xsd.Boolean `xml:"IsProperty,attr"`
	Source     Source      `json:",omitempty" xml:",omitempty"`
	Data       Data        `json:",omitempty" xml:",omitempty"`
}

type Source struct {
	SimpleItemDescription []SimpleItemDescription `json:",omitempty" xml:",omitempty"`
}

type Data struct {
	SimpleItemDescription []SimpleItemDescription `json:",omitempty" xml:",omitempty"`
}

type SimpleItemDescription struct {
	Name xsd.AnyType `xml:"Name,attr"`
	Type xsd.AnyType `xml:"Type,attr"`
}
