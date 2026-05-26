package imaging

//go:generate python3 ../python/gen_commands.py

import (
	"github.com/IOTechSystems/onvif/xsd"
	"github.com/IOTechSystems/onvif/xsd/onvif"
)

type GetServiceCapabilities struct {
	XMLName string `xml:"timg:GetServiceCapabilities"`
}

// todo: fill in response type
type GetServiceCapabilitiesResponse struct {
}

type GetImagingSettings struct {
	XMLName          string               `xml:"timg:GetImagingSettings"`
	VideoSourceToken onvif.ReferenceToken `xml:"timg:VideoSourceToken"`
}

type GetImagingSettingsResponse struct {
	ImagingSettings onvif.ImagingSettings20 `xml:"timg:ImagingSettings"`
}

type SetImagingSettings struct {
	XMLName          string                  `xml:"timg:SetImagingSettings"`
	VideoSourceToken onvif.ReferenceToken    `xml:"timg:VideoSourceToken"`
	ImagingSettings  onvif.ImagingSettings20 `xml:"timg:ImagingSettings"`
	ForcePersistence xsd.Boolean             `xml:"timg:ForcePersistence"`
}

type SetImagingSettingsResponse struct {
}

type GetOptions struct {
	XMLName          string               `xml:"timg:GetOptions"`
	VideoSourceToken onvif.ReferenceToken `xml:"timg:VideoSourceToken"`
}

// todo: fill in response type
type GetOptionsResponse struct {
}

type Move struct {
	XMLName          string               `xml:"timg:Move"`
	VideoSourceToken onvif.ReferenceToken `xml:"timg:VideoSourceToken"`
	Focus            onvif.FocusMove      `xml:"timg:Focus"`
}

// todo: fill in response type
type MoveResponse struct {
}

type GetMoveOptions struct {
	XMLName          string               `xml:"timg:GetMoveOptions"`
	VideoSourceToken onvif.ReferenceToken `xml:"timg:VideoSourceToken"`
}

// todo: fill in response type
type GetMoveOptionsResponse struct {
}

type Stop struct {
	XMLName          string               `xml:"timg:Stop"`
	VideoSourceToken onvif.ReferenceToken `xml:"timg:VideoSourceToken"`
}

// todo: fill in response type
type StopResponse struct {
}

type GetStatus struct {
	XMLName          string               `xml:"timg:GetStatus"`
	VideoSourceToken onvif.ReferenceToken `xml:"timg:VideoSourceToken"`
}

// todo: fill in response type
type GetStatusResponse struct {
}

type GetPresets struct {
	XMLName          string               `xml:"timg:GetPresets"`
	VideoSourceToken onvif.ReferenceToken `xml:"timg:VideoSourceToken"`
}

// todo: fill in response type
type GetPresetsResponse struct {
}

type GetCurrentPreset struct {
	XMLName          string               `xml:"timg:GetCurrentPreset"`
	VideoSourceToken onvif.ReferenceToken `xml:"timg:VideoSourceToken"`
}

// todo: fill in response type
type GetCurrentPresetResponse struct {
}

type SetCurrentPreset struct {
	XMLName          string               `xml:"timg:SetCurrentPreset"`
	VideoSourceToken onvif.ReferenceToken `xml:"timg:VideoSourceToken"`
	PresetToken      onvif.ReferenceToken `xml:"timg:PresetToken"`
}

type SetCurrentPresetResponse struct {
}
