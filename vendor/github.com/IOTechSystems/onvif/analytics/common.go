package analytics

import "github.com/IOTechSystems/onvif/xsd"

type Parameters struct {
	SimpleItemDescription  []SimpleItemDescription  `json:",omitempty"`
	ElementItemDescription []ElementItemDescription `json:",omitempty"`
	Extension              *xsd.String              `json:",omitempty"`
}

type SimpleItemDescription struct {
	Name  string `json:",omitempty" xml:",attr"`
	Type  string `json:",omitempty" xml:",attr"`
	Value string `json:",omitempty" xml:",attr"`
}

type ElementItemDescription struct {
	Name  string `json:",omitempty" xml:",attr"`
	Value string `json:",omitempty" xml:",attr"`
}

type Messages struct {
	IsProperty  *xsd.Boolean `json:",omitempty" xml:",attr"`
	Source      *Source      `json:",omitempty"`
	Key         *Key         `json:",omitempty"`
	Data        *Data        `json:",omitempty"`
	Extension   *xsd.String  `json:",omitempty"`
	ParentTopic *xsd.String  `json:",omitempty"`
}

type Source struct {
	SimpleItemDescription  []SimpleItemDescription  `json:",omitempty"`
	ElementItemDescription []ElementItemDescription `json:",omitempty"`
	Extension              *xsd.String              `json:",omitempty"`
}

type Key struct {
	SimpleItemDescription  []SimpleItemDescription  `json:",omitempty"`
	ElementItemDescription []ElementItemDescription `json:",omitempty"`
	Extension              *xsd.String              `json:",omitempty"`
}

type Data struct {
	SimpleItemDescription  []SimpleItemDescription  `json:",omitempty"`
	ElementItemDescription []ElementItemDescription `json:",omitempty"`
	Extension              *xsd.String              `json:",omitempty"`
}
