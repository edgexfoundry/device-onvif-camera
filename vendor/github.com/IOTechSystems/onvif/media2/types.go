package media2

//go:generate python3 ../python/gen_commands.py

import (
	"github.com/IOTechSystems/onvif/xsd"
	"github.com/IOTechSystems/onvif/xsd/onvif"
)

type GetVideoEncoderConfigurations struct {
	XMLName            string               `xml:"tr2:GetVideoEncoderConfigurations"`
	ConfigurationToken onvif.ReferenceToken `xml:"tr2:ConfigurationToken,omitempty"`
	ProfileToken       onvif.ReferenceToken `xml:"tr2:ProfileToken,omitempty"`
}

type GetVideoEncoderConfigurationsResponse struct {
	Configurations []onvif.VideoEncoder2Configuration
}

type SetVideoEncoderConfiguration struct {
	XMLName       string                                  `xml:"tr2:SetVideoEncoderConfiguration"`
	Configuration onvif.VideoEncoder2ConfigurationRequest `xml:"tr2:Configuration"`
}

type SetVideoEncoderConfigurationResponse struct{}

type GetVideoEncoderConfigurationOptions struct {
	XMLName            string               `xml:"tr2:GetVideoEncoderConfigurationOptions"`
	ConfigurationToken onvif.ReferenceToken `xml:"tr2:ConfigurationToken,omitempty"`
	ProfileToken       onvif.ReferenceToken `xml:"tr2:ProfileToken,omitempty"`
}

type GetVideoEncoderConfigurationOptionsResponse struct {
	Options []onvif.VideoEncoder2ConfigurationOptions
}

type GetProfiles struct {
	XMLName string `xml:"tr2:GetProfiles"`
}

type GetProfilesResponse struct {
	Profiles []Profile
}

type Profile struct {
	Token string `xml:"token,attr"`
	Fixed bool   `xml:"fixed,attr"`
	Name  string
}

type GetAnalyticsConfigurations struct {
	XMLName string `xml:"tr2:GetAnalyticsConfigurations"`
}

type GetAnalyticsConfigurationsResponse struct {
	Configurations []Configurations
}

type Configurations struct {
	onvif.ConfigurationEntity
	AnalyticsEngineConfiguration *AnalyticsEngineConfiguration `json:",omitempty"`
	RuleEngineConfiguration      *RuleEngineConfiguration      `json:",omitempty"`
}

type AnalyticsEngineConfiguration struct {
	AnalyticsModule []AnalyticsModule
}

type AnalyticsModule struct {
	Name       string `xml:",attr"`
	Type       string `xml:",attr"`
	Parameters Parameters
}

type RuleEngineConfiguration struct {
	Rule []Rule `json:",omitempty"`
}

type Rule struct {
	Name       string `xml:",attr"`
	Type       string `xml:",attr"`
	Parameters Parameters
}

type Parameters struct {
	SimpleItem  []SimpleItem  `json:",omitempty"`
	ElementItem []ElementItem `json:",omitempty"`
}

type SimpleItem struct {
	Name  string `xml:",attr"`
	Value string `xml:",attr"`
}

type ElementItem struct {
	Name string `xml:",attr"`
}

type AddConfiguration struct {
	XMLName       string `xml:"tr2:AddConfiguration"`
	ProfileToken  string `xml:"tr2:ProfileToken"`
	Name          string `xml:"tr2:Name,omitempty"`
	Configuration []Configuration
}

type AddConfigurationResponse struct{}

type RemoveConfiguration struct {
	XMLName       string `xml:"tr2:RemoveConfiguration"`
	ProfileToken  string `xml:"tr2:ProfileToken"`
	Configuration []Configuration
}

type RemoveConfigurationResponse struct{}

type Configuration struct {
	XMLName xsd.String  `xml:"tr2:Configuration"`
	Type    *xsd.String `xml:"tr2:Type,omitempty"`
	Token   *xsd.String `xml:"tr2:Token,omitempty"`
}
