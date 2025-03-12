package onvif

import (
	"github.com/IOTechSystems/onvif/xsd"
)

// BUG(r): Enum types implemented as simple string

//TODO: enumerations
//TODO: type <typeName> struct {Any string} convert to type <typeName> AnyType
//TODO: process restrictions

//todo посмотреть все Extensions (Any string)
//todo что делать с xs:any = Any
//todo IntList и ему подобные. Проверить нужен ли слайс. Изменить на slice
//todo посмотреть можно ли заменить StreamType и ему подобные типы на вмтроенные типы
//todo оттестировать тип VideoSourceMode из-за Description-а

//todo в документации описать, что Capabilities повторяеся у каждого сервиса, поэтому у каждого свой Capabilities (MediaCapabilities)
//todo AuxiliaryData и другие simpleTypes, как реализовать рестрикшн
//todo Name и ему подобные необходимо изучить на наличие "List of..." ошибок

//todo Add in buit in

type ContentType string // minLength value="3"
type DNSName xsd.Token

type DeviceEntity struct {
	Token ReferenceToken `xml:"token,attr"`
}

type ReferenceToken xsd.String

type Name xsd.String

type IntRectangle struct {
	X      int `xml:"x,attr"`
	Y      int `xml:"y,attr"`
	Width  int `xml:"width,attr"`
	Height int `xml:"height,attr"`
}

type IntRectangleRange struct {
	XRange      IntRange
	YRange      IntRange
	WidthRange  IntRange
	HeightRange IntRange
}

type IntRange struct {
	Min int
	Max int
}

type FloatRange struct {
	Min float64 `xml:"Min"`
	Max float64 `xml:"Max"`
}

type OSDConfiguration struct {
	DeviceEntity                  `xml:"token,attr"`
	VideoSourceConfigurationToken OSDReference              `xml:"VideoSourceConfigurationToken"`
	Type                          OSDType                   `xml:"Type"`
	Position                      OSDPosConfiguration       `xml:"Position"`
	TextString                    OSDTextConfiguration      `xml:"TextString"`
	Image                         OSDImgConfiguration       `xml:"Image"`
	Extension                     OSDConfigurationExtension `xml:"Extension"`
}

type OSDType xsd.String

type OSDPosConfiguration struct {
	Type      string                       `xml:"Type"`
	Pos       Vector                       `xml:"Pos"`
	Extension OSDPosConfigurationExtension `xml:"Extension"`
}

type Vector struct {
	X float64 `xml:"x,attr"`
	Y float64 `xml:"y,attr"`
}

type OSDPosConfigurationExtension xsd.AnyType

type OSDReference ReferenceToken

type OSDTextConfiguration struct {
	IsPersistentText xsd.Boolean `xml:"IsPersistentText,attr"`

	Type            xsd.String                    `xml:"Type"`
	DateFormat      xsd.String                    `xml:"DateFormat"`
	TimeFormat      xsd.String                    `xml:"TimeFormat"`
	FontSize        xsd.Int                       `xml:"FontSize"`
	FontColor       OSDColor                      `xml:"FontColor"`
	BackgroundColor OSDColor                      `xml:"BackgroundColor"`
	PlainText       xsd.String                    `xml:"PlainText"`
	Extension       OSDTextConfigurationExtension `xml:"Extension"`
}

type OSDColor struct {
	Transparent int `xml:"Transparent,attr"`

	Color Color `xml:"Color"`
}

type Color struct {
	X          float64    `xml:"X,attr"`
	Y          float64    `xml:"Y,attr"`
	Z          float64    `xml:"Z,attr"`
	Colorspace xsd.AnyURI `xml:"Colorspace,attr"`
}

type OSDTextConfigurationExtension xsd.AnyType

type OSDImgConfiguration struct {
	ImgPath   xsd.AnyURI                   `xml:"ImgPath"`
	Extension OSDImgConfigurationExtension `xml:"Extension"`
}

type OSDImgConfigurationExtension xsd.AnyType

type OSDConfigurationExtension xsd.AnyType

type VideoSource struct {
	DeviceEntity
	Framerate  float64
	Resolution VideoResolution
	Imaging    ImagingSettings
	Extension  VideoSourceExtension
}

type VideoResolution struct {
	Width  *xsd.Int `json:",omitempty"`
	Height *xsd.Int `json:",omitempty"`
}

type VideoResolutionRequest struct {
	Width  *xsd.Int `xml:"onvif:Width,omitempty"`
	Height *xsd.Int `xml:"onvif:Height,omitempty"`
}

type ImagingSettings struct {
	BacklightCompensation BacklightCompensation
	Brightness            float64
	ColorSaturation       float64
	Contrast              float64
	Exposure              Exposure
	Focus                 FocusConfiguration
	IrCutFilter           IrCutFilterMode
	Sharpness             float64
	WideDynamicRange      WideDynamicRange
	WhiteBalance          WhiteBalance
	Extension             ImagingSettingsExtension
}

type BacklightCompensation struct {
	Mode  BacklightCompensationMode
	Level float64
}

type BacklightCompensationMode xsd.String

type Exposure struct {
	Mode            ExposureMode
	Priority        ExposurePriority
	Window          Rectangle
	MinExposureTime float64
	MaxExposureTime float64
	MinGain         float64
	MaxGain         float64
	MinIris         float64
	MaxIris         float64
	ExposureTime    float64
	Gain            float64
	Iris            float64
}

type ExposureMode xsd.String

type ExposurePriority xsd.String

type Rectangle struct {
	Bottom float64 `xml:"bottom,attr"`
	Top    float64 `xml:"top,attr"`
	Right  float64 `xml:"right,attr"`
	Left   float64 `xml:"left,attr"`
}

type FocusConfiguration struct {
	AutoFocusMode AutoFocusMode
	DefaultSpeed  float64
	NearLimit     float64
	FarLimit      float64
}

type AutoFocusMode xsd.String

type IrCutFilterMode xsd.String

type WideDynamicRange struct {
	Mode  WideDynamicMode `xml:"Mode"`
	Level float64         `xml:"Level"`
}

type WideDynamicMode xsd.String

type WhiteBalance struct {
	Mode   WhiteBalanceMode
	CrGain float64
	CbGain float64
}

type WhiteBalanceMode xsd.String

type ImagingSettingsExtension xsd.AnyType

type VideoSourceExtension struct {
	Imaging   ImagingSettings20
	Extension VideoSourceExtension2
}

type ImagingSettings20 struct {
	BacklightCompensation BacklightCompensation20    `xml:"BacklightCompensation"`
	Brightness            float64                    `xml:"Brightness"`
	ColorSaturation       float64                    `xml:"ColorSaturation"`
	Contrast              float64                    `xml:"Contrast"`
	Exposure              Exposure20                 `xml:"Exposure"`
	Focus                 FocusConfiguration20       `xml:"Focus"`
	IrCutFilter           IrCutFilterMode            `xml:"IrCutFilter"`
	Sharpness             float64                    `xml:"Sharpness"`
	WideDynamicRange      WideDynamicRange20         `xml:"WideDynamicRange"`
	WhiteBalance          WhiteBalance20             `xml:"WhiteBalance"`
	Extension             ImagingSettingsExtension20 `xml:"Extension"`
}

type BacklightCompensation20 struct {
	Mode  BacklightCompensationMode `xml:"Mode"`
	Level float64                   `xml:"Level"`
}

type Exposure20 struct {
	Mode            ExposureMode     `xml:"Mode"`
	Priority        ExposurePriority `xml:"Priority"`
	Window          Rectangle        `xml:"Window"`
	MinExposureTime float64          `xml:"MinExposureTime"`
	MaxExposureTime float64          `xml:"MaxExposureTime"`
	MinGain         float64          `xml:"MinGain"`
	MaxGain         float64          `xml:"MaxGain"`
	MinIris         float64          `xml:"MinIris"`
	MaxIris         float64          `xml:"MaxIris"`
	ExposureTime    float64          `xml:"ExposureTime"`
	Gain            float64          `xml:"Gain"`
	Iris            float64          `xml:"Iris"`
}

type FocusConfiguration20 struct {
	AutoFocusMode AutoFocusMode                 `xml:"AutoFocusMode"`
	DefaultSpeed  float64                       `xml:"DefaultSpeed"`
	NearLimit     float64                       `xml:"NearLimit"`
	FarLimit      float64                       `xml:"FarLimit"`
	Extension     FocusConfiguration20Extension `xml:"Extension"`
}

type FocusConfiguration20Extension xsd.AnyType

type WideDynamicRange20 struct {
	Mode  WideDynamicMode `xml:"Mode"`
	Level float64         `xml:"Level"`
}

type WhiteBalance20 struct {
	Mode      WhiteBalanceMode        `xml:"Mode"`
	CrGain    float64                 `xml:"CrGain"`
	CbGain    float64                 `xml:"CbGain"`
	Extension WhiteBalance20Extension `xml:"Extension"`
}

type WhiteBalance20Extension xsd.AnyType

type ImagingSettingsExtension20 struct {
	ImageStabilization ImageStabilization          `xml:"ImageStabilization"`
	Extension          ImagingSettingsExtension202 `xml:"Extension"`
}

type ImageStabilization struct {
	Mode      ImageStabilizationMode      `xml:"Mode"`
	Level     float64                     `xml:"Level"`
	Extension ImageStabilizationExtension `xml:"Extension"`
}

type ImageStabilizationMode xsd.String

type ImageStabilizationExtension xsd.AnyType

type ImagingSettingsExtension202 struct {
	IrCutFilterAutoAdjustment IrCutFilterAutoAdjustment   `xml:"IrCutFilterAutoAdjustment"`
	Extension                 ImagingSettingsExtension203 `xml:"Extension"`
}

type IrCutFilterAutoAdjustment struct {
	BoundaryType   string                             `xml:"BoundaryType"`
	BoundaryOffset float64                            `xml:"BoundaryOffset"`
	ResponseTime   xsd.Duration                       `xml:"ResponseTime"`
	Extension      IrCutFilterAutoAdjustmentExtension `xml:"Extension"`
}

type IrCutFilterAutoAdjustmentExtension xsd.AnyType

type ImagingSettingsExtension203 struct {
	ToneCompensation ToneCompensation            `xml:"ToneCompensation"`
	Defogging        Defogging                   `xml:"Defogging"`
	NoiseReduction   NoiseReduction              `xml:"NoiseReduction"`
	Extension        ImagingSettingsExtension204 `xml:"Extension"`
}

type ToneCompensation struct {
	Mode      string                    `xml:"Mode"`
	Level     float64                   `xml:"Level"`
	Extension ToneCompensationExtension `xml:"Extension"`
}

type ToneCompensationExtension xsd.AnyType

type Defogging struct {
	Mode      string
	Level     float64
	Extension DefoggingExtension
}

type DefoggingExtension xsd.AnyType

type NoiseReduction struct {
	Level float64 `xml:"Level"`
}

type ImagingSettingsExtension204 xsd.AnyType

type VideoSourceExtension2 xsd.AnyType

type AudioSource struct {
	DeviceEntity
	Channels int
}

type AudioOutput struct {
	DeviceEntity
}

type Profile struct {
	Token                       ReferenceToken `xml:"token,attr"`
	Fixed                       bool           `xml:"fixed,attr"`
	Name                        Name
	VideoSourceConfiguration    *VideoSourceConfiguration    `xml:",omitempty"`
	AudioSourceConfiguration    *AudioSourceConfiguration    `xml:",omitempty"`
	VideoEncoderConfiguration   *VideoEncoderConfiguration   `xml:",omitempty"`
	AudioEncoderConfiguration   *AudioEncoderConfiguration   `xml:",omitempty"`
	VideoAnalyticsConfiguration *VideoAnalyticsConfiguration `xml:",omitempty"`
	PTZConfiguration            *PTZConfiguration            `xml:",omitempty"`
	MetadataConfiguration       *MetadataConfiguration       `xml:",omitempty"`
	Extension                   *ProfileExtension            `xml:",omitempty"`
}

type VideoSourceConfiguration struct {
	ConfigurationEntity
	ViewMode    string                             `xml:"ViewMode,attr"`
	SourceToken *ReferenceToken                    `xml:",omitempty"`
	Bounds      *IntRectangle                      `xml:",omitempty"`
	Extension   *VideoSourceConfigurationExtension `xml:",omitempty"`
}

type ConfigurationEntity struct {
	Token    ReferenceToken `json:",omitempty" xml:"token,attr,omitempty"`
	Name     Name           `json:",omitempty" xml:",omitempty"`
	UseCount int            `json:",omitempty" xml:",omitempty"`
}

type ConfigurationEntityRequest struct {
	Token    ReferenceToken `xml:"token,attr,omitempty"`
	Name     Name           `xml:"onvif:Name,omitempty"`
	UseCount int            `xml:"onvif:UseCount,omitempty"`
}

type VideoSourceConfigurationExtension struct {
	Rotate    *Rotate                             `xml:",omitempty"`
	Extension *VideoSourceConfigurationExtension2 `xml:",omitempty"`
}

type Rotate struct {
	Mode      RotateMode      `xml:"Mode"`
	Degree    xsd.Int         `xml:"Degree"`
	Extension RotateExtension `xml:"Extension"`
}

type RotateMode xsd.String

type RotateExtension xsd.AnyType

type VideoSourceConfigurationExtension2 struct {
	LensDescription  LensDescription  `xml:"LensDescription"`
	SceneOrientation SceneOrientation `xml:"SceneOrientation"`
}

type LensDescription struct {
	FocalLength float64        `xml:"FocalLength,attr"`
	Offset      LensOffset     `xml:"Offset"`
	Projection  LensProjection `xml:"Projection"`
	XFactor     float64        `xml:"XFactor"`
}

type LensOffset struct {
	X float64 `xml:"x,attr"`
	Y float64 `xml:"y,attr"`
}

type LensProjection struct {
	Angle         float64 `xml:"Angle"`
	Radius        float64 `xml:"Radius"`
	Transmittance float64 `xml:"Transmittance"`
}

type SceneOrientation struct {
	Mode        SceneOrientationMode `xml:"Mode"`
	Orientation xsd.String           `xml:"Orientation"`
}

type SceneOrientationMode xsd.String

type AudioSourceConfiguration struct {
	ConfigurationEntity
	SourceToken ReferenceToken `xml:"SourceToken"`
}

type VideoEncoderConfiguration struct {
	ConfigurationEntity
	Encoding       *VideoEncoding          `json:",omitempty"`
	Resolution     *VideoResolution        `json:",omitempty"`
	Quality        float64                 `json:",omitempty"`
	RateControl    *VideoRateControl       `json:",omitempty"`
	MPEG4          *Mpeg4Configuration     `json:",omitempty"`
	H264           *H264Configuration      `json:",omitempty"`
	Multicast      *MulticastConfiguration `json:",omitempty"`
	SessionTimeout *xsd.Duration           `json:",omitempty"`
}

type VideoEncoderConfigurationRequest struct {
	ConfigurationEntityRequest
	Encoding       *VideoEncoding                 `xml:"onvif:Encoding,omitempty"`
	Resolution     *VideoResolutionRequest        `xml:"onvif:Resolution,omitempty"`
	Quality        *xsd.Float                     `xml:"onvif:Quality,omitempty"`
	RateControl    *VideoRateControlRequest       `xml:"onvif:RateControl,omitempty"`
	MPEG4          *Mpeg4ConfigurationRequest     `xml:"onvif:MPEG4,omitempty"`
	H264           *H264ConfigurationRequest      `xml:"onvif:H264,omitempty"`
	Multicast      *MulticastConfigurationRequest `xml:"onvif:Multicast,omitempty"`
	SessionTimeout *xsd.Duration                  `xml:"onvif:SessionTimeout,omitempty"`
}

type VideoEncoding xsd.String

type VideoRateControl struct {
	FrameRateLimit   *xsd.Int `json:",omitempty"`
	EncodingInterval *xsd.Int `json:",omitempty"`
	BitrateLimit     *xsd.Int `json:",omitempty"`
}

type VideoRateControlRequest struct {
	FrameRateLimit   *xsd.Int `xml:"onvif:FrameRateLimit,omitempty"`
	EncodingInterval *xsd.Int `xml:"onvif:EncodingInterval,omitempty"`
	BitrateLimit     *xsd.Int `xml:"onvif:BitrateLimit,omitempty"`
}

type Mpeg4Configuration struct {
	GovLength    *xsd.Int      `json:",omitempty"`
	Mpeg4Profile *Mpeg4Profile `json:",omitempty"`
}

type Mpeg4ConfigurationRequest struct {
	GovLength    *xsd.Int      `xml:"onvif:GovLength,omitempty"`
	Mpeg4Profile *Mpeg4Profile `xml:"onvif:Mpeg4Profile,omitempty"`
}

type Mpeg4Profile xsd.String

type H264Configuration struct {
	GovLength   *xsd.Int     `json:",omitempty"`
	H264Profile *H264Profile `json:",omitempty"`
}

type H264ConfigurationRequest struct {
	GovLength   *xsd.Int     `xml:"onvif:GovLength,omitempty"`
	H264Profile *H264Profile `xml:"onvif:H264Profile,omitempty"`
}

type H264Profile xsd.String

type MulticastConfiguration struct {
	Address   *IPAddress   `json:",omitempty"`
	Port      *xsd.Int     `json:",omitempty"`
	TTL       *xsd.Int     `json:",omitempty"`
	AutoStart *xsd.Boolean `json:",omitempty"`
}

type MulticastConfigurationRequest struct {
	Address   *IPAddressRequest `xml:"onvif:Address,omitempty"`
	Port      *xsd.Int          `xml:"onvif:Port,omitempty"`
	TTL       *xsd.Int          `xml:"onvif:TTL,omitempty"`
	AutoStart *xsd.Boolean      `xml:"onvif:AutoStart,omitempty"`
}

type IPAddress struct {
	Type        IPType      `json:",omitempty"`
	IPv4Address IPv4Address `json:",omitempty"`
	IPv6Address IPv6Address `json:",omitempty"`
}

type IPAddressRequest struct {
	Type        IPType      `xml:"onvif:Type,omitempty"`
	IPv4Address IPv4Address `xml:"onvif:IPv4Address,omitempty"`
	IPv6Address IPv6Address `xml:"onvif:IPv6Address,omitempty"`
}

type IPType xsd.String

// IPv4 address
type IPv4Address xsd.Token

// IPv6 address
type IPv6Address xsd.Token

type AudioEncoderConfiguration struct {
	ConfigurationEntity
	Encoding       AudioEncoding          `xml:"Encoding"`
	Bitrate        int                    `xml:"Bitrate"`
	SampleRate     int                    `xml:"SampleRate"`
	Multicast      MulticastConfiguration `xml:"Multicast"`
	SessionTimeout xsd.Duration           `xml:"SessionTimeout"`
}

type AudioEncoding xsd.String

type VideoAnalyticsConfiguration struct {
	ConfigurationEntity
	AnalyticsEngineConfiguration *AnalyticsEngineConfiguration `xml:"AnalyticsEngineConfiguration"`
	RuleEngineConfiguration      *RuleEngineConfiguration      `xml:"RuleEngineConfiguration"`
}

type AnalyticsEngineConfiguration struct {
	AnalyticsModule []AnalyticsModule                      `json:",omitempty"`
	Extension       *AnalyticsEngineConfigurationExtension `json:",omitempty"`
}

type AnalyticsModule struct {
	Name       string `xml:",attr"`
	Type       string `xml:",attr"`
	Parameters Parameters
}

type Parameters struct {
	SimpleItem  []SimpleItem  `json:",omitempty"`
	ElementItem []ElementItem `json:",omitempty"`
}

type AnalyticsEngineConfigurationRequest struct {
	AnalyticsModule *ConfigRequest                         `xml:"onvif:AnalyticsEngineConfigurationRequest,omitempty"`
	Extension       *AnalyticsEngineConfigurationExtension `xml:"onvif:Extension,omitempty"`
}

type Config struct {
	Name       string     `json:",omitempty" xml:",attr"`
	Type       *xsd.QName `json:",omitempty" xml:",attr"`
	Parameters *ItemList  `json:",omitempty"`
}

type ItemList struct {
	SimpleItem  []SimpleItem       `json:",omitempty"`
	ElementItem []ElementItem      `json:",omitempty"`
	Extension   *ItemListExtension `json:",omitempty"`
}

type SimpleItem struct {
	Name  *xsd.String `json:",omitempty" xml:",attr"`
	Value *xsd.String `json:",omitempty" xml:",attr"`
}

type ElementItem struct {
	Name  *xsd.String `json:",omitempty" xml:",attr"`
	Value *xsd.String `json:",omitempty" xml:",attr"`
}

type ConfigRequest struct {
	Name       string           `xml:",attr,omitempty"`
	Type       *xsd.QName       `xml:",attr,omitempty"`
	Parameters *ItemListRequest `xml:"onvif:Parameters,omitempty"`
}

type ItemListRequest struct {
	SimpleItem  []SimpleItemRequest  `xml:"onvif:SimpleItem,omitempty"`
	ElementItem []ElementItemRequest `xml:"onvif:ElementItem,omitempty"`
	Extension   *ItemListExtension   `xml:"onvif:Extension,omitempty"`
}

type ElementItemRequest struct {
	Name     string    `xml:",attr,omitempty"`
	Polyline *Polyline `xml:"onvif:Polyline,omitempty"`
}

type Polyline struct {
	Point []Point `xml:"onvif:Point,omitempty"`
}

type Point struct {
	X *xsd.String `xml:"x,attr,omitempty"`
	Y *xsd.String `xml:"onvif:y,attr,omitempty"`
}

type SimpleItemRequest struct {
	Name  string            `xml:",attr,omitempty"`
	Value xsd.AnySimpleType `xml:",attr,omitempty"`
}

type ItemListExtension xsd.AnyType

type AnalyticsEngineConfigurationExtension xsd.AnyType

type RuleEngineConfiguration struct {
	Rule      *Config                           `json:",omitempty"`
	Extension *RuleEngineConfigurationExtension `json:",omitempty"`
}

type RuleEngineConfigurationExtension xsd.AnyType

type PTZConfiguration struct {
	PTZConfigurationEntity
	Token                                  ReferenceToken             `xml:"token,attr"`
	MoveRamp                               int                        `json:",omitempty" xml:"MoveRamp,attr,omitempty"`
	PresetRamp                             int                        `json:",omitempty" xml:"PresetRamp,attr,omitempty"`
	PresetTourRamp                         int                        `json:",omitempty" xml:"PresetTourRamp,attr,omitempty"`
	NodeToken                              *ReferenceToken            `json:",omitempty" xml:"tptz:NodeToken,omitempty"`
	DefaultAbsolutePantTiltPositionSpace   *xsd.AnyURI                `json:",omitempty" xml:",omitempty"`
	DefaultAbsoluteZoomPositionSpace       *xsd.AnyURI                `json:",omitempty" xml:",omitempty"`
	DefaultRelativePanTiltTranslationSpace *xsd.AnyURI                `json:",omitempty" xml:",omitempty"`
	DefaultRelativeZoomTranslationSpace    *xsd.AnyURI                `json:",omitempty" xml:",omitempty"`
	DefaultContinuousPanTiltVelocitySpace  *xsd.AnyURI                `json:",omitempty" xml:",omitempty"`
	DefaultContinuousZoomVelocitySpace     *xsd.AnyURI                `json:",omitempty" xml:",omitempty"`
	DefaultPTZSpeed                        *PTZSpeed                  `json:",omitempty" xml:",omitempty"`
	DefaultPTZTimeout                      *xsd.Duration              `json:",omitempty" xml:",omitempty"`
	PanTiltLimits                          *PanTiltLimits             `json:",omitempty" xml:",omitempty"`
	ZoomLimits                             *ZoomLimits                `json:",omitempty" xml:",omitempty"`
	Extension                              *PTZConfigurationExtension `json:",omitempty" xml:",omitempty"`
}

type PTZConfigurationEntity struct {
	Token    ReferenceToken `json:",omitempty" xml:"token,attr,omitempty"`
	Name     Name           `json:",omitempty" xml:"tptz:Name,omitempty"`
	UseCount int            `json:",omitempty" xml:"tptz:UseCount,omitempty"`
}

type PTZSpeed struct {
	PanTilt *Vector2D `json:",omitempty" xml:"onvif:PanTilt,omitempty"`
	Zoom    *Vector1D `json:",omitempty" xml:"onvif:Zoom,omitempty"`
}

type Vector2D struct {
	X     float64     `xml:"x,attr,omitempty"`
	Y     float64     `xml:"y,attr,omitempty"`
	Space *xsd.AnyURI `xml:"space,attr,omitempty"`
}

type Vector1D struct {
	X     float64     `xml:"x,attr,omitempty"`
	Space *xsd.AnyURI `xml:"space,attr,omitempty"`
}

type PanTiltLimits struct {
	Range *Space2DDescription `xml:"Range,omitempty"`
}

type Space2DDescription struct {
	URI    *xsd.AnyURI `xml:"URI,omitempty"`
	XRange *FloatRange `xml:"XRange,omitempty"`
	YRange *FloatRange `xml:"YRange,omitempty"`
}

type ZoomLimits struct {
	Range Space1DDescription `xml:"Range"`
}

type Space1DDescription struct {
	URI    xsd.AnyURI `xml:"URI"`
	XRange FloatRange `xml:"XRange"`
}

type PTZConfigurationExtension struct {
	PTControlDirection *PTControlDirection         `xml:"PTControlDirection,omitempty"`
	Extension          *PTZConfigurationExtension2 `xml:"Extension,omitempty"`
}

type PTControlDirection struct {
	EFlip     EFlip                       `xml:"EFlip"`
	Reverse   Reverse                     `xml:"Reverse"`
	Extension PTControlDirectionExtension `xml:"Extension"`
}

type EFlip struct {
	Mode EFlipMode `xml:"Mode"`
}

type EFlipMode xsd.String

type Reverse struct {
	Mode ReverseMode `xml:"Mode"`
}

type ReverseMode xsd.String

type PTControlDirectionExtension xsd.AnyType

type PTZConfigurationExtension2 xsd.AnyType

type MetadataConfiguration struct {
	ConfigurationEntity
	CompressionType              string                          `json:",omitempty" xml:",attr,omitempty"`
	PTZStatus                    *PTZFilter                      `json:",omitempty" xml:",omitempty"`
	Events                       *EventSubscription              `json:",omitempty" xml:",omitempty"`
	Analytics                    *xsd.Boolean                    `json:",omitempty" xml:",omitempty"`
	Multicast                    *MulticastConfiguration         `json:",omitempty" xml:",omitempty"`
	SessionTimeout               *xsd.Duration                   `json:",omitempty" xml:",omitempty"`
	AnalyticsEngineConfiguration *AnalyticsEngineConfiguration   `json:",omitempty" xml:",omitempty"`
	Extension                    *MetadataConfigurationExtension `json:",omitempty" xml:",omitempty"`
}

type MetadataConfigurationRequest struct {
	ConfigurationEntity
	CompressionType              string                               `xml:"onvif:CompressionType,attr,omitempty"`
	PTZStatus                    *PTZFilterRequest                    `xml:"onvif:PTZStatus,omitempty"`
	Events                       *EventSubscriptionRequest            `xml:"onvif:Events,omitempty"`
	Analytics                    *xsd.Boolean                         `xml:"onvif:Analytics,omitempty"`
	Multicast                    *MulticastConfigurationRequest       `xml:"onvif:Multicast,omitempty"`
	SessionTimeout               *xsd.Duration                        `xml:"onvif:CompressionType,omitempty"`
	AnalyticsEngineConfiguration *AnalyticsEngineConfigurationRequest `xml:"onvif:AnalyticsEngineConfiguration,omitempty"`
	Extension                    *MetadataConfigurationExtension      `xml:"onvif:Extension,omitempty"`
}

type PTZFilter struct {
	Status   bool `xml:"Status"`
	Position bool `xml:"Position"`
}

type PTZFilterRequest struct {
	Status   bool `xml:"onvif:Status,omitempty"`
	Position bool `xml:"onvif:Position,omitempty"`
}

type EventSubscription struct {
	Filter             *FilterType         `json:",omitempty"`
	SubscriptionPolicy *SubscriptionPolicy `json:",omitempty"`
}

type EventSubscriptionRequest struct {
	Filter             FilterType         `xml:"onvif:Filter,omitempty"`
	SubscriptionPolicy SubscriptionPolicy `xml:"onvif:SubscriptionPolicy,omitempty"`
}

type FilterType xsd.AnyType

type SubscriptionPolicy xsd.AnyType

type MetadataConfigurationExtension xsd.AnyType

type ProfileExtension struct {
	AudioOutputConfiguration  *AudioOutputConfiguration  `xml:",omitempty"`
	AudioDecoderConfiguration *AudioDecoderConfiguration `xml:",omitempty"`
	Extension                 *ProfileExtension2         `xml:",omitempty"`
}

type AudioOutputConfiguration struct {
	ConfigurationEntity
	OutputToken ReferenceToken `xml:"OutputToken"`
	SendPrimacy xsd.AnyURI     `xml:"SendPrimacy"`
	OutputLevel int            `xml:"OutputLevel"`
}

type AudioDecoderConfiguration struct {
	ConfigurationEntity
}

type ProfileExtension2 xsd.AnyType

type VideoSourceConfigurationOptions struct {
	MaximumNumberOfProfiles    int `xml:"MaximumNumberOfProfiles,attr"`
	BoundsRange                IntRectangleRange
	VideoSourceTokensAvailable ReferenceToken
	Extension                  VideoSourceConfigurationOptionsExtension
}

type VideoSourceConfigurationOptionsExtension struct {
	Rotate    RotateOptions
	Extension VideoSourceConfigurationOptionsExtension2
}

type RotateOptions struct {
	Mode       RotateMode
	DegreeList IntList
	Extension  RotateOptionsExtension
}

type IntList struct {
	Items []int
}

type RotateOptionsExtension xsd.AnyType

type VideoSourceConfigurationOptionsExtension2 struct {
	SceneOrientationMode SceneOrientationMode
}

type VideoEncoderConfigurationOptions struct {
	QualityRange *IntRange                     `json:",omitempty"`
	JPEG         *JpegOptions                  `json:",omitempty"`
	MPEG4        *Mpeg4Options                 `json:",omitempty"`
	H264         *H264Options                  `json:",omitempty"`
	Extension    *VideoEncoderOptionsExtension `json:",omitempty"`
}

type JpegOptions struct {
	ResolutionsAvailable  []VideoResolution
	FrameRateRange        IntRange
	EncodingIntervalRange IntRange
}

type Mpeg4Options struct {
	ResolutionsAvailable   VideoResolution
	GovLengthRange         IntRange
	FrameRateRange         IntRange
	EncodingIntervalRange  IntRange
	Mpeg4ProfilesSupported Mpeg4Profile
}

type H264Options struct {
	ResolutionsAvailable  []VideoResolution
	GovLengthRange        IntRange
	FrameRateRange        IntRange
	EncodingIntervalRange IntRange
	H264ProfilesSupported []H264Profile
}

type VideoEncoderOptionsExtension struct {
	JPEG      *JpegOptions2                  `json:",omitempty"`
	MPEG4     *Mpeg4Options2                 `json:",omitempty"`
	H264      *H264Options2                  `json:",omitempty"`
	Extension *VideoEncoderOptionsExtension2 `json:",omitempty"`
}

type JpegOptions2 struct {
	JpegOptions
	BitrateRange IntRange
}

type Mpeg4Options2 struct {
	Mpeg4Options
	BitrateRange IntRange
}

type H264Options2 struct {
	H264Options
	BitrateRange IntRange
}

type VideoEncoderOptionsExtension2 xsd.AnyType

type AudioSourceConfigurationOptions struct {
	InputTokensAvailable ReferenceToken
	Extension            AudioSourceOptionsExtension
}

type AudioSourceOptionsExtension xsd.AnyType

type AudioEncoderConfigurationOptions struct {
	Options AudioEncoderConfigurationOption
}

type AudioEncoderConfigurationOption struct {
	Encoding       AudioEncoding
	BitrateList    IntList
	SampleRateList IntList
}

type MetadataConfigurationOptions struct {
	PTZStatusFilterOptions *PTZStatusFilterOptions                `json:",omitempty" xml:",omitempty"`
	Extension              *MetadataConfigurationOptionsExtension `json:",omitempty" xml:",omitempty"`
}

type PTZStatusFilterOptions struct {
	PanTiltStatusSupported   bool
	ZoomStatusSupported      bool
	PanTiltPositionSupported bool
	ZoomPositionSupported    bool
	Extension                *PTZStatusFilterOptionsExtension `json:",omitempty" xml:",omitempty"`
}

type PTZStatusFilterOptionsExtension xsd.AnyType

type MetadataConfigurationOptionsExtension struct {
	CompressionType string                                  `json:",omitempty" xml:",omitempty"`
	Extension       *MetadataConfigurationOptionsExtension2 `json:",omitempty" xml:",omitempty"`
}

type MetadataConfigurationOptionsExtension2 xsd.AnyType

type AudioOutputConfigurationOptions struct {
	OutputTokensAvailable ReferenceToken
	SendPrimacyOptions    xsd.AnyURI
	OutputLevelRange      IntRange
}

type AudioDecoderConfigurationOptions struct {
	AACDecOptions  AACDecOptions
	G711DecOptions G711DecOptions
	G726DecOptions G726DecOptions
	Extension      AudioDecoderConfigurationOptionsExtension
}

type AACDecOptions struct {
	Bitrate         IntList
	SampleRateRange IntList
}

type G711DecOptions struct {
	Bitrate         IntList
	SampleRateRange IntList
}

type G726DecOptions struct {
	Bitrate         IntList
	SampleRateRange IntList
}

type AudioDecoderConfigurationOptionsExtension xsd.AnyType

type StreamSetup struct {
	Stream    *StreamType `xml:"onvif:Stream,omitempty"`
	Transport *Transport  `xml:"onvif:Transport,omitempty"`
}

type StreamType xsd.String

type Transport struct {
	Protocol *TransportProtocol `xml:"onvif:Protocol,omitempty"`
	Tunnel   *Transport         `xml:"onvif:Tunnel,omitempty"`
}

// enum
type TransportProtocol xsd.String

type MediaUri struct {
	Uri                 xsd.AnyURI
	InvalidAfterConnect bool
	InvalidAfterReboot  bool
	Timeout             xsd.Duration
}

type VideoSourceMode struct {
	Token         ReferenceToken `xml:"token,attr"`
	Enabled       bool           `xml:"Enabled,attr"`
	MaxFramerate  float64
	MaxResolution VideoResolution
	Encodings     EncodingTypes
	Reboot        bool
	Description   Description
	Extension     VideoSourceModeExtension
}

type EncodingTypes struct {
	EncodingTypes []string
}

type Description struct {
	Description string
}

type VideoSourceModeExtension xsd.AnyType

type OSDConfigurationOptions struct {
	MaximumNumberOfOSDs MaximumNumberOfOSDs
	Type                OSDType
	PositionOption      string
	TextOption          OSDTextOptions
	ImageOption         OSDImgOptions
	Extension           OSDConfigurationOptionsExtension
}

type MaximumNumberOfOSDs struct {
	Total       int `xml:"Total,attr"`
	Image       int `xml:"Image,attr"`
	PlainText   int `xml:"PlainText,attr"`
	Date        int `xml:"Date,attr"`
	Time        int `xml:"Time,attr"`
	DateAndTime int `xml:"DateAndTime,attr"`
}

type OSDTextOptions struct {
	Type            string
	FontSizeRange   IntRange
	DateFormat      string
	TimeFormat      string
	FontColor       OSDColorOptions
	BackgroundColor OSDColorOptions
	Extension       OSDTextOptionsExtension
}

type OSDColorOptions struct {
	Color       ColorOptions
	Transparent IntRange
	Extension   OSDColorOptionsExtension
}

type ColorOptions struct {
	ColorList       Color
	ColorspaceRange ColorspaceRange
}

type ColorspaceRange struct {
	X          FloatRange
	Y          FloatRange
	Z          FloatRange
	Colorspace xsd.AnyURI
}

type OSDColorOptionsExtension xsd.AnyType

type OSDTextOptionsExtension xsd.AnyType

type OSDImgOptions struct {
	FormatsSupported StringAttrList `xml:"FormatsSupported,attr"`
	MaxSize          int            `xml:"MaxSize,attr"`
	MaxWidth         int            `xml:"MaxWidth,attr"`
	MaxHeight        int            `xml:"MaxHeight,attr"`

	ImagePath xsd.AnyURI
	Extension OSDImgOptionsExtension
}

type StringAttrList []string

type OSDImgOptionsExtension xsd.AnyType

type OSDConfigurationOptionsExtension xsd.AnyType

//PTZ

type PTZNode struct {
	DeviceEntity
	FixedHomePosition      *xsd.Boolean      `json:",omitempty" xml:",attr,omitempty"`
	GeoMove                *xsd.Boolean      `json:",omitempty" xml:",attr,omitempty"`
	Name                   *Name             `json:",omitempty" xml:",omitempty"`
	SupportedPTZSpaces     *PTZSpaces        `json:",omitempty" xml:",omitempty"`
	MaximumNumberOfPresets int               `json:",omitempty" xml:",omitempty"`
	HomeSupported          *xsd.Boolean      `json:",omitempty" xml:",omitempty"`
	AuxiliaryCommands      *AuxiliaryData    `json:",omitempty" xml:",omitempty"`
	Extension              *PTZNodeExtension `json:",omitempty" xml:",omitempty"`
}

type PTZSpaces struct {
	AbsolutePanTiltPositionSpace    Space2DDescription
	AbsoluteZoomPositionSpace       Space1DDescription
	RelativePanTiltTranslationSpace Space2DDescription
	RelativeZoomTranslationSpace    Space1DDescription
	ContinuousPanTiltVelocitySpace  Space2DDescription
	ContinuousZoomVelocitySpace     Space1DDescription
	PanTiltSpeedSpace               Space1DDescription
	ZoomSpeedSpace                  Space1DDescription
	Extension                       PTZSpacesExtension
}

type PTZSpacesExtension xsd.AnyType

// TODO: restriction
type AuxiliaryData xsd.String

type PTZNodeExtension struct {
	SupportedPresetTour PTZPresetTourSupported
	Extension           PTZNodeExtension2
}

type PTZPresetTourSupported struct {
	MaximumNumberOfPresetTours int
	PTZPresetTourOperation     PTZPresetTourOperation
	Extension                  PTZPresetTourSupportedExtension
}

type PTZPresetTourOperation xsd.String
type PTZPresetTourSupportedExtension xsd.AnyType

type PTZNodeExtension2 xsd.AnyType

type PTZConfigurationOptions struct {
	PTZRamps           *IntAttrList               `json:",omitempty" xml:",attr,omitempty"`
	Spaces             *PTZSpaces                 `json:",omitempty" xml:",omitempty"`
	PTZTimeout         *DurationRange             `json:",omitempty" xml:",omitempty"`
	PTControlDirection *PTControlDirectionOptions `json:",omitempty" xml:",omitempty"`
	Extension          *PTZConfigurationOptions2  `json:",omitempty" xml:",omitempty"`
}

type IntAttrList []int

type DurationRange struct {
	Min xsd.Duration
	Max xsd.Duration
}

type PTControlDirectionOptions struct {
	EFlip     EFlipOptions
	Reverse   ReverseOptions
	Extension PTControlDirectionOptionsExtension
}

type EFlipOptions struct {
	Mode      EFlipMode
	Extension EFlipOptionsExtension
}

type EFlipOptionsExtension xsd.AnyType

type ReverseOptions struct {
	Mode      ReverseMode
	Extension ReverseOptionsExtension
}

type ReverseOptionsExtension xsd.AnyType

type PTControlDirectionOptionsExtension xsd.AnyType

type PTZConfigurationOptions2 xsd.AnyType

type PTZPreset struct {
	Token       ReferenceToken `xml:"token,attr"`
	Name        Name
	PTZPosition PTZVector
}

type PTZVector struct {
	PanTilt *Vector2D `json:",omitempty" xml:"PanTilt,omitempty"`
	Zoom    *Vector1D `json:",omitempty" xml:"Zoom,omitempty"`
}

type PTZStatus struct {
	Position   PTZVector     `json:",omitempty" xml:",omitempty"`
	MoveStatus PTZMoveStatus `json:",omitempty" xml:",omitempty"`
	Error      string        `json:",omitempty" xml:",omitempty"`
	UtcTime    string        `json:",omitempty" xml:",omitempty"`
}

type PTZMoveStatus struct {
	PanTilt string `json:",omitempty" xml:",omitempty"`
	Zoom    string `json:",omitempty" xml:",omitempty"`
}

type MoveStatus struct {
	Status string
}

type GeoLocation struct {
	Lon       xsd.Double `xml:"lon,attr"`
	Lat       xsd.Double `xml:"lat,attr"`
	Elevation xsd.Float  `xml:"elevation,attr"`
}

type PresetTour struct {
	Token             ReferenceToken                 `xml:"token,attr"`
	Name              Name                           `xml:"Name"`
	Status            PTZPresetTourStatus            `xml:"Status"`
	AutoStart         xsd.Boolean                    `xml:"AutoStart"`
	StartingCondition PTZPresetTourStartingCondition `xml:"StartingCondition"`
	TourSpot          PTZPresetTourSpot              `xml:"TourSpot"`
	Extension         PTZPresetTourExtension         `xml:"Extension"`
}

type PTZPresetTourStatus struct {
	State           PTZPresetTourState           `xml:"State"`
	CurrentTourSpot PTZPresetTourSpot            `xml:"CurrentTourSpot"`
	Extension       PTZPresetTourStatusExtension `xml:"Extension"`
}

type PTZPresetTourState xsd.String

type PTZPresetTourSpot struct {
	PresetDetail PTZPresetTourPresetDetail  `xml:"PresetDetail"`
	Speed        PTZSpeed                   `xml:"Speed"`
	StayTime     xsd.Duration               `xml:"StayTime"`
	Extension    PTZPresetTourSpotExtension `xml:"Extension"`
}

type PTZPresetTourPresetDetail struct {
	PresetToken   ReferenceToken             `xml:"PresetToken"`
	Home          xsd.Boolean                `xml:"Home"`
	PTZPosition   PTZVector                  `xml:"PTZPosition"`
	TypeExtension PTZPresetTourTypeExtension `xml:"TypeExtension"`
}

type PTZPresetTourTypeExtension xsd.AnyType

type PTZPresetTourSpotExtension xsd.AnyType

type PTZPresetTourStatusExtension xsd.AnyType

type PTZPresetTourStartingCondition struct {
	RandomPresetOrder xsd.Boolean                             `xml:"RandomPresetOrder,attr"`
	RecurringTime     xsd.Int                                 `xml:"RecurringTime"`
	RecurringDuration xsd.Duration                            `xml:"RecurringDuration"`
	Direction         PTZPresetTourDirection                  `xml:"Direction"`
	Extension         PTZPresetTourStartingConditionExtension `xml:"Extension"`
}

type PTZPresetTourDirection xsd.String

type PTZPresetTourStartingConditionExtension xsd.AnyType

type PTZPresetTourExtension xsd.AnyType

type PTZPresetTourOptions struct {
	AutoStart         xsd.Boolean
	StartingCondition PTZPresetTourStartingConditionOptions
	TourSpot          PTZPresetTourSpotOptions
}

type PTZPresetTourStartingConditionOptions struct {
	RecurringTime     IntRange
	RecurringDuration DurationRange
	Direction         PTZPresetTourDirection
	Extension         PTZPresetTourStartingConditionOptionsExtension
}

type PTZPresetTourStartingConditionOptionsExtension xsd.AnyType

type PTZPresetTourSpotOptions struct {
	PresetDetail PTZPresetTourPresetDetailOptions
	StayTime     DurationRange
}

type PTZPresetTourPresetDetailOptions struct {
	PresetToken          ReferenceToken
	Home                 xsd.Boolean
	PanTiltPositionSpace Space2DDescription
	ZoomPositionSpace    Space1DDescription
	Extension            PTZPresetTourPresetDetailOptionsExtension
}

type PTZPresetTourPresetDetailOptionsExtension xsd.AnyType

//Device

type OnvifVersion struct {
	Major int
	Minor int
}

type SetDateTimeType xsd.String

type TimeZone struct {
	TZ xsd.Token `xml:"TZ"`
}

type SystemDateTime struct {
	DateTimeType    SetDateTimeType
	DaylightSavings xsd.Boolean
	TimeZone        TimeZone
	UTCDateTime     xsd.DateTime
	LocalDateTime   xsd.DateTime
	Extension       SystemDateTimeExtension
}

type SystemDateTimeExtension xsd.AnyType

type FactoryDefaultType xsd.String

type AttachmentData struct {
	ContentType ContentType `xml:"contentType,attr"`
	Include     Include     `xml:"inc:Include"`
}

type Include struct {
	Href xsd.AnyURI `xml:"href,attr"`
}

type BackupFile struct {
	Name string         `xml:"Name"`
	Data AttachmentData `xml:"Data"`
}

type SystemLogType xsd.String

type SystemLog struct {
	Binary AttachmentData
	String string
}

type SupportInformation struct {
	Binary AttachmentData
	String string
}

type Scope struct {
	ScopeDef  ScopeDefinition
	ScopeItem xsd.AnyURI
}

type ScopeDefinition xsd.String

type DiscoveryMode xsd.String

type NetworkHost struct {
	Type        NetworkHostType      `xml:"Type"`
	IPv4Address IPv4Address          `xml:"IPv4Address"`
	IPv6Address IPv6Address          `xml:"IPv6Address"`
	DNSname     DNSName              `xml:"DNSname"`
	Extension   NetworkHostExtension `xml:"Extension"`
}

type NetworkHostType xsd.String

type NetworkHostExtension xsd.String

type RemoteUser struct {
	Username           string      `xml:"Username"`
	Password           string      `xml:"Password"`
	UseDerivedPassword xsd.Boolean `xml:"UseDerivedPassword"`
}

type User struct {
	Username  string         `json:",omitempty" xml:",omitempty"`
	Password  string         `json:",omitempty" xml:",omitempty"`
	UserLevel *UserLevel     `json:",omitempty" xml:",omitempty"`
	Extension *UserExtension `json:",omitempty" xml:",omitempty"`
}

type UserRequest struct {
	Username  string         `xml:"onvif:Username,omitempty"`
	Password  string         `xml:"onvif:Password,omitempty"`
	UserLevel *UserLevel     `xml:"onvif:UserLevel,omitempty"`
	Extension *UserExtension `xml:"onvif:Extension,omitempty"`
}

type UserLevel xsd.String

type UserExtension xsd.String

type CapabilityCategory xsd.String

// Capabilities of device
type Capabilities struct {
	Analytics AnalyticsCapabilities
	Device    DeviceCapabilities
	Events    EventCapabilities
	Imaging   ImagingCapabilities
	Media     MediaCapabilities
	PTZ       PTZCapabilities
	Extension CapabilitiesExtension
}

// AnalyticsCapabilities Check
type AnalyticsCapabilities struct {
	XAddr                  xsd.AnyURI
	RuleSupport            xsd.Boolean
	AnalyticsModuleSupport xsd.Boolean
}

// DeviceCapabilities Check
type DeviceCapabilities struct {
	XAddr     xsd.AnyURI
	Network   NetworkCapabilities
	System    SystemCapabilities
	IO        IOCapabilities
	Security  SecurityCapabilities
	Extension DeviceCapabilitiesExtension
}

// NetworkCapabilities Check
type NetworkCapabilities struct {
	IPFilter          xsd.Boolean
	ZeroConfiguration xsd.Boolean
	IPVersion6        xsd.Boolean
	DynDNS            xsd.Boolean
	Extension         NetworkCapabilitiesExtension
}

// NetworkCapabilitiesExtension Check
type NetworkCapabilitiesExtension struct {
	Dot11Configuration xsd.Boolean
	Extension          NetworkCapabilitiesExtension2
}

// NetworkCapabilitiesExtension2 Extension2
type NetworkCapabilitiesExtension2 xsd.AnyType

// SystemCapabilities check
type SystemCapabilities struct {
	DiscoveryResolve  xsd.Boolean
	DiscoveryBye      xsd.Boolean
	RemoteDiscovery   xsd.Boolean
	SystemBackup      xsd.Boolean
	SystemLogging     xsd.Boolean
	FirmwareUpgrade   xsd.Boolean
	SupportedVersions OnvifVersion
	Extension         SystemCapabilitiesExtension
}

type SystemCapabilitiesExtension struct {
	HttpFirmwareUpgrade    xsd.Boolean
	HttpSystemBackup       xsd.Boolean
	HttpSystemLogging      xsd.Boolean
	HttpSupportInformation xsd.Boolean
	Extension              SystemCapabilitiesExtension2
}

type SystemCapabilitiesExtension2 xsd.AnyType

type IOCapabilities struct {
	InputConnectors int
	RelayOutputs    int
	Extension       IOCapabilitiesExtension
}

type IOCapabilitiesExtension struct {
	Auxiliary         xsd.Boolean
	AuxiliaryCommands AuxiliaryData
	Extension         IOCapabilitiesExtension2
}

type IOCapabilitiesExtension2 xsd.AnyType

type SecurityCapabilities struct {
	TLS1_1               xsd.Boolean
	TLS1_2               xsd.Boolean
	OnboardKeyGeneration xsd.Boolean
	AccessPolicyConfig   xsd.Boolean
	X_509Token           xsd.Boolean
	SAMLToken            xsd.Boolean
	KerberosToken        xsd.Boolean
	RELToken             xsd.Boolean
	Extension            SecurityCapabilitiesExtension
}

type SecurityCapabilitiesExtension struct {
	TLS1_0    xsd.Boolean
	Extension SecurityCapabilitiesExtension2
}

type SecurityCapabilitiesExtension2 struct {
	Dot1X              xsd.Boolean
	SupportedEAPMethod int
	RemoteUserHandling xsd.Boolean
}

type DeviceCapabilitiesExtension xsd.AnyType

type EventCapabilities struct {
	XAddr                                         xsd.AnyURI
	WSSubscriptionPolicySupport                   xsd.Boolean
	WSPullPointSupport                            xsd.Boolean
	WSPausableSubscriptionManagerInterfaceSupport xsd.Boolean
}

type ImagingCapabilities struct {
	XAddr xsd.AnyURI
}

type MediaCapabilities struct {
	XAddr                 xsd.AnyURI
	StreamingCapabilities RealTimeStreamingCapabilities
	Extension             MediaCapabilitiesExtension
}

type RealTimeStreamingCapabilities struct {
	RTPMulticast xsd.Boolean
	RTP_TCP      xsd.Boolean
	RTP_RTSP_TCP xsd.Boolean
	Extension    RealTimeStreamingCapabilitiesExtension
}

type RealTimeStreamingCapabilitiesExtension xsd.AnyType

type MediaCapabilitiesExtension struct {
	ProfileCapabilities ProfileCapabilities
}

type ProfileCapabilities struct {
	MaximumNumberOfProfiles int
}

type PTZCapabilities struct {
	XAddr xsd.AnyURI
}

type CapabilitiesExtension struct {
	DeviceIO        DeviceIOCapabilities
	Display         DisplayCapabilities
	Recording       RecordingCapabilities
	Search          SearchCapabilities
	Replay          ReplayCapabilities
	Receiver        ReceiverCapabilities
	AnalyticsDevice AnalyticsDeviceCapabilities
	Extensions      CapabilitiesExtension2
}

type DeviceIOCapabilities struct {
	XAddr        xsd.AnyURI
	VideoSources int
	VideoOutputs int
	AudioSources int
	AudioOutputs int
	RelayOutputs int
}

type DisplayCapabilities struct {
	XAddr       xsd.AnyURI
	FixedLayout xsd.Boolean
}

type RecordingCapabilities struct {
	XAddr              xsd.AnyURI
	ReceiverSource     xsd.Boolean
	MediaProfileSource xsd.Boolean
	DynamicRecordings  xsd.Boolean
	DynamicTracks      xsd.Boolean
	MaxStringLength    int
}

type SearchCapabilities struct {
	XAddr          xsd.AnyURI
	MetadataSearch xsd.Boolean
}

type ReplayCapabilities struct {
	XAddr xsd.AnyURI
}

type ReceiverCapabilities struct {
	XAddr                xsd.AnyURI
	RTP_Multicast        xsd.Boolean
	RTP_TCP              xsd.Boolean
	RTP_RTSP_TCP         xsd.Boolean
	SupportedReceivers   int
	MaximumRTSPURILength int
}

type AnalyticsDeviceCapabilities struct {
	XAddr       xsd.AnyURI
	RuleSupport xsd.Boolean
	Extension   AnalyticsDeviceExtension
}

type AnalyticsDeviceExtension xsd.AnyType

type CapabilitiesExtension2 xsd.AnyType

type HostnameInformation struct {
	FromDHCP  *xsd.Boolean                  `json:"FromDHCP,omitempty"`
	Name      *xsd.Token                    `json:"Name,omitempty"`
	Extension *HostnameInformationExtension `json:"Extension,omitempty"`
}

type HostnameInformationExtension xsd.AnyType

type DNSInformation struct {
	FromDHCP     *xsd.Boolean             `json:"FromDHCP,omitempty"`
	SearchDomain *xsd.Token               `json:"SearchDomain,omitempty"`
	DNSFromDHCP  *IPAddress               `json:"DNSFromDHCP,omitempty"`
	DNSManual    *IPAddress               `json:"DNSManual,omitempty"`
	Extension    *DNSInformationExtension `json:"Extension,omitempty"`
}

type DNSInformationExtension xsd.AnyType

type NTPInformation struct {
	FromDHCP    xsd.Boolean
	NTPFromDHCP NetworkHost
	NTPManual   NetworkHost
	Extension   NTPInformationExtension
}

type NTPInformationExtension xsd.AnyType

type DynamicDNSInformation struct {
	Type      DynamicDNSType
	Name      DNSName
	TTL       xsd.Duration
	Extension DynamicDNSInformationExtension
}

// TODO: enumeration
type DynamicDNSType xsd.String

type DynamicDNSInformationExtension xsd.AnyType

type NetworkInterface struct {
	DeviceEntity
	Enabled   *xsd.Boolean               `json:",omitempty"`
	Info      *NetworkInterfaceInfo      `json:",omitempty"`
	Link      *NetworkInterfaceLink      `json:",omitempty"`
	IPv4      *IPv4NetworkInterface      `json:",omitempty"`
	IPv6      *IPv6NetworkInterface      `json:",omitempty"`
	Extension *NetworkInterfaceExtension `json:",omitempty"`
}

type NetworkInterfaceInfo struct {
	Name      xsd.String `json:"Name,omitempty"`
	HwAddress HwAddress  `json:"HwAddress,omitempty"`
	MTU       xsd.Int    `json:"MTU,omitempty"`
}

type HwAddress xsd.Token

type NetworkInterfaceLink struct {
	AdminSettings NetworkInterfaceConnectionSetting
	OperSettings  NetworkInterfaceConnectionSetting
	InterfaceType *IANA_IfTypes `xml:"IANA-IfTypes,omitempty" json:"IANA-IfTypes,omitempty"`
}

type IANA_IfTypes xsd.Int

type NetworkInterfaceConnectionSetting struct {
	AutoNegotiation *xsd.Boolean `xml:"onvif:AutoNegotiation,omitempty" json:"AutoNegotiation,omitempty"`
	Speed           *xsd.Int     `xml:"onvif:Speed,omitempty" json:"Speed,omitempty"`
	Duplex          *Duplex      `xml:"onvif:Duplex,omitempty" json:"Duplex,omitempty"`
}

// TODO: enum
type Duplex xsd.String

type NetworkInterfaceExtension struct {
	InterfaceType IANA_IfTypes
	Dot3          *Dot3Configuration  `xml:"Dot3,omitempty" json:"Dot3,omitempty"`
	Dot11         *Dot11Configuration `xml:"Dot11,omitempty" json:"Dot11,omitempty"`
	Extension     NetworkInterfaceExtension2
}

type NetworkInterfaceExtension2 xsd.AnyType

type Dot11Configuration struct {
	SSID     Dot11SSIDType                  `xml:"SSID,omitempty" json:"SSID,omitempty"`
	Mode     Dot11StationMode               `xml:"Mode,omitempty" json:"Mode,omitempty"`
	Alias    Name                           `xml:"Alias,omitempty" json:"Alias,omitempty"`
	Priority NetworkInterfaceConfigPriority `xml:"Priority,omitempty" json:"Priority,omitempty"`
	Security Dot11SecurityConfiguration     `xml:"Security,omitempty" json:"Security,omitempty"`
}

type Dot11SecurityConfiguration struct {
	Mode      Dot11SecurityMode                   `xml:"Mode,omitempty" json:"Mode,omitempty"`
	Algorithm Dot11Cipher                         `xml:"Algorithm,omitempty" json:"Algorithm,omitempty"`
	PSK       Dot11PSKSet                         `xml:"PSK,omitempty" json:"PSK,omitempty"`
	Dot1X     ReferenceToken                      `xml:"Dot1X,omitempty" json:"Dot1X,omitempty"`
	Extension Dot11SecurityConfigurationExtension `xml:"Extension,omitempty" json:"Extension,omitempty"`
}

type Dot11SecurityConfigurationExtension xsd.AnyType

type Dot11PSKSet struct {
	Key        Dot11PSK             `xml:"Key,omitempty" json:"Key,omitempty"`
	Passphrase Dot11PSKPassphrase   `xml:"Passphrase,omitempty" json:"Passphrase,omitempty"`
	Extension  Dot11PSKSetExtension `xml:"Extension,omitempty" json:"Extension,omitempty"`
}

type Dot11PSKSetExtension xsd.AnyType

type Dot11PSKPassphrase xsd.String

type Dot11PSK xsd.HexBinary

// TODO: enumeration
type Dot11Cipher xsd.String

// TODO: enumeration
type Dot11SecurityMode xsd.String

// TODO: restrictions
type NetworkInterfaceConfigPriority xsd.Integer

// TODO: enumeration
type Dot11StationMode xsd.String

// TODO: restrictions
type Dot11SSIDType xsd.HexBinary

type Dot3Configuration xsd.String

type IPv6NetworkInterface struct {
	Enabled xsd.Boolean
	Config  IPv6Configuration
}

type IPv6Configuration struct {
	AcceptRouterAdvert xsd.Boolean
	DHCP               IPv6DHCPConfiguration
	Manual             PrefixedIPv6Address
	LinkLocal          PrefixedIPv6Address
	FromDHCP           PrefixedIPv6Address
	FromRA             PrefixedIPv6Address
	Extension          IPv6ConfigurationExtension
}

type IPv6ConfigurationExtension xsd.AnyType

type PrefixedIPv6Address struct {
	Address      IPv6Address `xml:"Address,omitempty" json:"Address,omitempty"`
	PrefixLength xsd.Int     `xml:"PrefixLength,omitempty" json:"PrefixLength,omitempty"`
}

// TODO: enumeration
type IPv6DHCPConfiguration xsd.String

type IPv4NetworkInterface struct {
	Enabled *xsd.Boolean       `json:"Enabled,omitempty"`
	Config  *IPv4Configuration `json:"Config,omitempty"`
}

type IPv4Configuration struct {
	Manual    *PrefixedIPv4Address `json:"Manual,omitempty"`
	LinkLocal *PrefixedIPv4Address `json:"LinkLocal,omitempty"`
	FromDHCP  *PrefixedIPv4Address `json:"FromDHCP,omitempty"`
	DHCP      *xsd.Boolean         `json:"DHCP,omitempty"`
}

// optional, unbounded
type PrefixedIPv4Address struct {
	Address      IPv4Address `xml:"Address" json:"Address,omitempty"`
	PrefixLength xsd.Int     `xml:"PrefixLength" json:"PrefixLength,omitempty"`
}

type NetworkInterfaceSetConfiguration struct {
	Enabled   *xsd.Boolean                               `xml:"onvif:Enabled,omitempty"`
	Link      *NetworkInterfaceConnectionSetting         `xml:"onvif:Link,omitempty"`
	MTU       *xsd.Int                                   `xml:"onvif:MTU,omitempty"`
	IPv4      *IPv4NetworkInterfaceSetConfiguration      `xml:"onvif:IPv4,omitempty"`
	IPv6      *IPv6NetworkInterfaceSetConfiguration      `xml:"onvif:IPv6,omitempty"`
	Extension *NetworkInterfaceSetConfigurationExtension `xml:"onvif:Extension,omitempty"`
}

type NetworkInterfaceSetConfigurationExtension struct {
	Dot3      Dot3Configuration                          `xml:"onvif:Dot3,omitempty"`
	Dot11     Dot11Configuration                         `xml:"onvif:Dot11,omitempty"`
	Extension NetworkInterfaceSetConfigurationExtension2 `xml:"onvif:Extension,omitempty"`
}

type NetworkInterfaceSetConfigurationExtension2 xsd.AnyType

type IPv6NetworkInterfaceSetConfiguration struct {
	Enabled            *xsd.Boolean           `xml:"onvif:Enabled,omitempty" json:",omitempty"`
	AcceptRouterAdvert *xsd.Boolean           `xml:"onvif:AcceptRouterAdvert,omitempty" json:",omitempty"`
	Manual             *PrefixedIPv6Address   `xml:"onvif:Manual,omitempty" json:",omitempty"`
	DHCP               *IPv6DHCPConfiguration `xml:"onvif:DHCP,omitempty" json:",omitempty"`
}

type IPv4NetworkInterfaceSetConfiguration struct {
	Enabled *xsd.Boolean         `xml:"onvif:Enabled,omitempty"`
	Manual  *PrefixedIPv4Address `xml:"onvif:Manual,omitempty"`
	DHCP    *xsd.Boolean         `xml:"onvif:DHCP,omitempty"`
}

type NetworkProtocolResponse struct {
	Name      *NetworkProtocolType      `json:",omitempty"`
	Enabled   *xsd.Boolean              `json:",omitempty"`
	Port      *xsd.Int                  `json:",omitempty"`
	Extension *NetworkProtocolExtension `json:",omitempty"`
}

type NetworkProtocolRequest struct {
	Name      *NetworkProtocolType      `xml:"onvif:Name,omitempty"`
	Enabled   *xsd.Boolean              `xml:"onvif:Enabled,omitempty"`
	Port      *xsd.Int                  `xml:"onvif:Port,omitempty"`
	Extension *NetworkProtocolExtension `xml:"onvif:Extension,omitempty"`
}

type NetworkProtocolExtension xsd.AnyType

// TODO: enumeration
type NetworkProtocolType xsd.String

type NetworkGateway struct {
	IPv4Address *IPv4Address `json:"IPv4Address,omitempty"`
	IPv6Address *IPv6Address `json:"IPv6Address,omitempty"`
}

type NetworkZeroConfiguration struct {
	InterfaceToken ReferenceToken
	Enabled        xsd.Boolean
	Addresses      IPv4Address
	Extension      NetworkZeroConfigurationExtension
}

type NetworkZeroConfigurationExtension struct {
	Additional *NetworkZeroConfiguration
	Extension  NetworkZeroConfigurationExtension2
}

type NetworkZeroConfigurationExtension2 xsd.AnyType

type IPAddressFilter struct {
	Type        IPAddressFilterType      `xml:"Type,omitempty"`
	IPv4Address PrefixedIPv4Address      `xml:"IPv4Address,omitempty"`
	IPv6Address PrefixedIPv6Address      `xml:"IPv6Address,omitempty"`
	Extension   IPAddressFilterExtension `xml:"Extension,omitempty"`
}

type IPAddressFilterExtension xsd.AnyType

// enum { 'Allow', 'Deny' }
// TODO: enumeration
type IPAddressFilterType xsd.String

// TODO: attribite <xs:attribute ref="xmime:contentType" use="optional"/>
type BinaryData struct {
	X    ContentType      `xml:"xmime:contentType,attr"`
	Data xsd.Base64Binary `xml:"Data"`
}

type Certificate struct {
	CertificateID xsd.Token  `xml:"CertificateID"`
	Certificate   BinaryData `xml:"Certificate"`
}

type CertificateStatus struct {
	CertificateID xsd.Token   `xml:"CertificateID"`
	Status        xsd.Boolean `xml:"Status"`
}

type RelayOutput struct {
	DeviceEntity
	Properties RelayOutputSettings
}

type RelayOutputSettings struct {
	Mode      RelayMode      `xml:"Mode"`
	DelayTime xsd.Duration   `xml:"DelayTime"`
	IdleState RelayIdleState `xml:"IdleState"`
}

// TODO:enumeration
type RelayIdleState xsd.String

// TODO: enumeration
type RelayMode xsd.String

// TODO: enumeration
type RelayLogicalState xsd.String

type CertificateWithPrivateKey struct {
	CertificateID xsd.Token  `xml:"CertificateID"`
	Certificate   BinaryData `xml:"Certificate"`
	PrivateKey    BinaryData `xml:"PrivateKey"`
}

type CertificateInformation struct {
	CertificateID      xsd.Token
	IssuerDN           xsd.String
	SubjectDN          xsd.String
	KeyUsage           CertificateUsage
	ExtendedKeyUsage   CertificateUsage
	KeyLength          xsd.Int
	Version            xsd.String
	SerialNum          xsd.String
	SignatureAlgorithm xsd.String
	Validity           DateTimeRange
	Extension          CertificateInformationExtension
}

type CertificateInformationExtension xsd.AnyType

type DateTimeRange struct {
	From  xsd.DateTime
	Until xsd.DateTime
}

type CertificateUsage struct {
	Critical         xsd.Boolean `xml:"Critical,attr"`
	CertificateUsage xsd.String
}

type Dot1XConfiguration struct {
	Dot1XConfigurationToken ReferenceToken              `xml:"Dot1XConfigurationToken"`
	Identity                xsd.String                  `xml:"Identity"`
	AnonymousID             xsd.String                  `xml:"AnonymousID,omitempty"`
	EAPMethod               xsd.Int                     `xml:"EAPMethod"`
	CACertificateID         xsd.Token                   `xml:"CACertificateID,omitempty"`
	EAPMethodConfiguration  EAPMethodConfiguration      `xml:"EAPMethodConfiguration,omitempty"`
	Extension               Dot1XConfigurationExtension `xml:"Extension,omitempty"`
}

type Dot1XConfigurationExtension xsd.AnyType

type EAPMethodConfiguration struct {
	TLSConfiguration TLSConfiguration   `xml:"TLSConfiguration,omitempty"`
	Password         xsd.String         `xml:"Password,omitempty"`
	Extension        EapMethodExtension `xml:"Extension,omitempty"`
}

type EapMethodExtension xsd.AnyType

type TLSConfiguration struct {
	CertificateID xsd.Token `xml:"CertificateID,omitempty"`
}

type Dot11Capabilities struct {
	TKIP                  xsd.Boolean
	ScanAvailableNetworks xsd.Boolean
	MultipleConfiguration xsd.Boolean
	AdHocStationMode      xsd.Boolean
	WEP                   xsd.Boolean
}

type Dot11Status struct {
	SSID              Dot11SSIDType
	BSSID             xsd.String
	PairCipher        Dot11Cipher
	GroupCipher       Dot11Cipher
	SignalStrength    Dot11SignalStrength
	ActiveConfigAlias ReferenceToken
}

// TODO: enumeration
type Dot11SignalStrength xsd.String

type Dot11AvailableNetworks struct {
	SSID                  Dot11SSIDType
	BSSID                 xsd.String
	AuthAndMangementSuite Dot11AuthAndMangementSuite
	PairCipher            Dot11Cipher
	GroupCipher           Dot11Cipher
	SignalStrength        Dot11SignalStrength
	Extension             Dot11AvailableNetworksExtension
}

type Dot11AvailableNetworksExtension xsd.AnyType

// TODO: enumeration
type Dot11AuthAndMangementSuite xsd.String

type SystemLogUriList struct {
	SystemLog SystemLogUri
}

type SystemLogUri struct {
	Type SystemLogType
	Uri  xsd.AnyURI
}

type LocationEntity struct {
	Entity    xsd.String     `xml:"Entity,attr"`
	Token     ReferenceToken `xml:"Token,attr"`
	Fixed     xsd.Boolean    `xml:"Fixed,attr"`
	GeoSource xsd.AnyURI     `xml:"GeoSource,attr"`
	AutoGeo   xsd.Boolean    `xml:"AutoGeo,attr"`

	GeoLocation      GeoLocation      `xml:"GeoLocation"`
	GeoOrientation   GeoOrientation   `xml:"GeoOrientation"`
	LocalLocation    LocalLocation    `xml:"LocalLocation"`
	LocalOrientation LocalOrientation `xml:"LocalOrientation"`
}

type LocalOrientation struct {
	Lon       xsd.Double `xml:"lon,attr"`
	Lat       xsd.Double `xml:"lat,attr"`
	Elevation xsd.Float  `xml:"elevation,attr"`
}

type LocalLocation struct {
	X xsd.Float `xml:"x,attr"`
	Y xsd.Float `xml:"y,attr"`
	Z xsd.Float `xml:"z,attr"`
}

type GeoOrientation struct {
	Roll  xsd.Float `xml:"roll,attr"`
	Pitch xsd.Float `xml:"pitch,attr"`
	Yaw   xsd.Float `xml:"yaw,attr"`
}

type FocusMove struct {
	Absolute   AbsoluteFocus   `xml:"Absolute"`
	Relative   RelativeFocus   `xml:"Relative"`
	Continuous ContinuousFocus `xml:"Continuous"`
}

type ContinuousFocus struct {
	Speed xsd.Float `xml:"Speed"`
}

type RelativeFocus struct {
	Distance xsd.Float `xml:"Distance"`
	Speed    xsd.Float `xml:"Speed"`
}

type AbsoluteFocus struct {
	Position xsd.Float `xml:"Position"`
	Speed    xsd.Float `xml:"Speed"`
}

type DateTime struct {
	Time Time `xml:"Time"`
	Date Date `xml:"Date"`
}

type Time struct {
	Hour   xsd.Int `xml:"Hour"`
	Minute xsd.Int `xml:"Minute"`
	Second xsd.Int `xml:"Second"`
}

type Date struct {
	Year  xsd.Int `xml:"Year"`
	Month xsd.Int `xml:"Month"`
	Day   xsd.Int `xml:"Day"`
}

type DateTimeRequest struct {
	Time *TimeRequest `xml:"onvif:Time,omitempty"`
	Date *DateRequest `xml:"onvif:Date,omitempty"`
}

type TimeRequest struct {
	Hour   *xsd.Int `xml:"onvif:Hour,omitempty"`
	Minute *xsd.Int `xml:"onvif:Minute,omitempty"`
	Second *xsd.Int `xml:"onvif:Second,omitempty"`
}

type DateRequest struct {
	Year  *xsd.Int `xml:"onvif:Year,omitempty"`
	Month *xsd.Int `xml:"onvif:Month,omitempty"`
	Day   *xsd.Int `xml:"onvif:Day,omitempty"`
}
