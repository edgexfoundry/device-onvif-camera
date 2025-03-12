package recording

import (
	"encoding/xml"

	"github.com/IOTechSystems/onvif/xsd"
	"github.com/IOTechSystems/onvif/xsd/onvif"
)

// TrackType type
type TrackType string

const (
	// TrackTypeVideo const
	TrackTypeVideo TrackType = "Video"

	// TrackTypeAudio const
	TrackTypeAudio TrackType = "Audio"

	// TrackTypeMetadata const
	TrackTypeMetadata TrackType = "Metadata"

	// Placeholder for future extension.
	// TrackTypeExtended const
	TrackTypeExtended TrackType = "Extended"
)

// EncodingTypes type
type EncodingTypes []string

// GetServiceCapabilities type
type GetServiceCapabilities struct {
	XMLName xml.Name `xml:"tt:GetServiceCapabilities"`
}

// GetServiceCapabilitiesResponse type
type GetServiceCapabilitiesResponse struct {
	XMLName xml.Name `xml:"GetServiceCapabilitiesResponse"`

	// The capabilities for the recording service is returned in the Capabilities element.
	Capabilities Capabilities `xml:"Capabilities,omitempty"`
}

// CreateRecording type
type CreateRecording struct {
	XMLName xml.Name `xml:"tt:CreateRecording"`

	// Initial configuration for the recording.
	RecordingConfiguration RecordingConfiguration `xml:"tt:RecordingConfiguration,omitempty"`
}

// RecordingConfiguration type
type RecordingConfiguration struct {

	// Information about the source of the recording.
	Source RecordingSourceInformation `xml:"tt:Source,omitempty"`

	// Informative description of the source.
	Content onvif.Description `xml:"tt:Content,omitempty"`

	// Sspecifies the maximum time that data in any track within the
	// recording shall be stored. The device shall delete any data older than the maximum retention
	// time. Such data shall not be accessible anymore. If the MaximumRetentionPeriod is set to 0,
	// the device shall not limit the retention time of stored data, except by resource constraints.
	// Whatever the value of MaximumRetentionTime, the device may automatically delete
	// recordings to free up storage space for new recordings.
	MaximumRetentionTime xsd.Duration `xml:"tt:MaximumRetentionTime,omitempty"`
}

// RecordingSourceInformation type
type RecordingSourceInformation struct {

	//
	// Identifier for the source chosen by the client that creates the structure.
	// This identifier is opaque to the device. Clients may use any type of URI for this field. A device shall support at least 128 characters.
	SourceId xsd.AnyURI `xml:"tt:SourceId,omitempty"`

	// Informative user readable name of the source, e.g. "Camera23". A device shall support at least 20 characters.
	Name xsd.Name `xml:"tt:Name,omitempty"`

	// Informative description of the physical location of the source, e.g. the coordinates on a map.
	Location onvif.Description `xml:"tt:Location,omitempty"`

	// Informative description of the source.
	Description onvif.Description `xml:"tt:Description,omitempty"`

	// URI provided by the service supplying data to be recorded. A device shall support at least 128 characters.
	Address xsd.AnyURI `xml:"tt:Address,omitempty"`
}

// CreateRecordingResponse type
type CreateRecordingResponse struct {
	XMLName xml.Name `xml:"CreateRecordingResponse"`

	// The reference to the created recording.
	RecordingToken RecordingReference `xml:"RecordingToken,omitempty"`
}

// DeleteRecording type
type DeleteRecording struct {
	XMLName xml.Name `xml:"tt:DeleteRecording"`

	// The reference of the recording to be deleted.
	RecordingToken RecordingReference `xml:"tt:RecordingToken,omitempty"`
}

// DeleteRecordingResponse type
type DeleteRecordingResponse struct {
	XMLName xml.Name `xml:"DeleteRecordingResponse"`
}

// GetRecordings type
type GetRecordings struct {
	XMLName xml.Name `xml:"tt:GetRecordings"`
}

// GetRecordingsResponse type
type GetRecordingsResponse struct {
	XMLName xml.Name `xml:"GetRecordingsResponse"`

	// List of recording items.
	RecordingItem []GetRecordingsResponseItem
}

// GetRecordingsResponseItem type
type GetRecordingsResponseItem struct {
	// Token of the recording.
	RecordingToken RecordingReference

	// Configuration of the recording.
	Configuration struct {
		Source struct {
			SourceId    xsd.AnyURI
			Name        xsd.Name
			Location    xsd.String
			Description xsd.String
			Address     xsd.AnyURI
		}
		Content              xsd.String
		MaximumRetentionTime xsd.Duration
	}

	// List of tracks.
	Tracks GetTracksResponseList
}

// GetTracksResponseList type
type GetTracksResponseList struct {
	// Configuration of a track.
	Track []GetTracksResponseItem
}

// GetTracksResponseItem type
type GetTracksResponseItem struct {
	// Token of the track.
	TrackToken TrackReference
	// Configuration of the track.
	Configuration struct {
		TrackType   TrackType
		Description xsd.String
	}
}

// TrackConfiguration type
type TrackConfiguration struct {

	// Type of the track. It shall be equal to the strings “Video”,
	// “Audio” or “Metadata”. The track shall only be able to hold data of that type.
	TrackType TrackType `xml:"tt:TrackType,omitempty"`

	// Informative description of the track.
	Description onvif.Description `xml:"tt:Description,omitempty"`
}

// SetRecordingConfiguration type
type SetRecordingConfiguration struct {
	XMLName xml.Name `xml:"tt:SetRecordingConfiguration"`

	// Token of the recording that shall be changed.
	RecordingToken RecordingReference `xml:"tt:RecordingToken,omitempty"`

	// The new configuration.
	RecordingConfiguration RecordingConfiguration `xml:"tt:RecordingConfiguration,omitempty"`
}

// SetRecordingConfigurationResponse type
type SetRecordingConfigurationResponse struct {
	XMLName xml.Name `xml:"SetRecordingConfigurationResponse"`
}

// GetRecordingConfiguration type
type GetRecordingConfiguration struct {
	XMLName xml.Name `xml:"tt:GetRecordingConfiguration"`

	// Token of the configuration to be retrieved.
	RecordingToken RecordingReference `xml:"tt:RecordingToken,omitempty"`
}

// GetRecordingConfigurationResponse type
type GetRecordingConfigurationResponse struct {
	XMLName xml.Name `xml:"GetRecordingConfigurationResponse"`

	// Configuration of the recording.
	RecordingConfiguration RecordingConfiguration `xml:"RecordingConfiguration,omitempty"`
}

// CreateTrack type
type CreateTrack struct {
	XMLName xml.Name `xml:"tt:CreateTrack"`

	// Identifies the recording to which a track shall be added.
	RecordingToken RecordingReference `xml:"tt:RecordingToken,omitempty"`

	// The configuration of the new track.
	TrackConfiguration TrackConfiguration `xml:"tt:TrackConfiguration,omitempty"`
}

// CreateTrackResponse type
type CreateTrackResponse struct {
	XMLName xml.Name `xml:"CreateTrackResponse"`

	// The TrackToken shall identify the newly created track. The
	// TrackToken shall be unique within the recoding to which
	// the new track belongs.
	TrackToken TrackReference `xml:"TrackToken,omitempty"`
}

// DeleteTrack type
type DeleteTrack struct {
	XMLName xml.Name `xml:"tt:DeleteTrack"`

	// Token of the recording the track belongs to.
	RecordingToken RecordingReference `xml:"tt:RecordingToken,omitempty"`

	// Token of the track to be deleted.
	TrackToken TrackReference `xml:"tt:TrackToken,omitempty"`
}

// DeleteTrackResponse type
type DeleteTrackResponse struct {
	XMLName xml.Name `xml:"DeleteTrackResponse"`
}

// GetTrackConfiguration type
type GetTrackConfiguration struct {
	XMLName xml.Name `xml:"tt:GetTrackConfiguration"`

	// Token of the recording the track belongs to.
	RecordingToken RecordingReference `xml:"tt:RecordingToken,omitempty"`

	// Token of the track.
	TrackToken TrackReference `xml:"tt:TrackToken,omitempty"`
}

// GetTrackConfigurationResponse type
type GetTrackConfigurationResponse struct {
	XMLName xml.Name `xml:"GetTrackConfigurationResponse"`

	// Configuration of the track.
	TrackConfiguration TrackConfiguration `xml:"TrackConfiguration,omitempty"`
}

// SetTrackConfiguration type
type SetTrackConfiguration struct {
	XMLName xml.Name `xml:"tt:SetTrackConfiguration"`

	// Token of the recording the track belongs to.
	RecordingToken RecordingReference `xml:"tt:RecordingToken,omitempty"`

	// Token of the track to be modified.
	TrackToken TrackReference `xml:"tt:TrackToken,omitempty"`

	// New configuration for the track.
	TrackConfiguration TrackConfiguration `xml:"tt:TrackConfiguration,omitempty"`
}

// SetTrackConfigurationResponse type
type SetTrackConfigurationResponse struct {
	XMLName xml.Name `xml:"SetTrackConfigurationResponse"`
}

// CreateRecordingJob type
type CreateRecordingJob struct {
	XMLName xml.Name `xml:"tt:CreateRecordingJob"`

	// The initial configuration of the new recording job.
	JobConfiguration RecordingJobConfiguration `xml:"tt:JobConfiguration,omitempty"`
}

// CreateRecordingJobResponse type
type CreateRecordingJobResponse struct {
	XMLName xml.Name `xml:"CreateRecordingJobResponse"`

	// The JobToken shall identify the created recording job.
	JobToken RecordingJobReference `xml:"JobToken,omitempty"`

	//
	// The JobConfiguration structure shall be the configuration as it is used by the device. This may be different from the
	// JobConfiguration passed to CreateRecordingJob.
	JobConfiguration RecordingJobConfiguration `xml:"JobConfiguration,omitempty"`
}

// RecordingJobReference type
type RecordingJobReference ReferenceToken

// RecordingJobConfiguration type
type RecordingJobConfiguration struct {

	// Identifies the recording to which this job shall store the received data.
	RecordingToken RecordingReference `xml:"tt:RecordingToken,omitempty"`

	// The mode of the job. If it is idle, nothing shall happen. If it is active, the device shall try
	// to obtain data from the receivers. A client shall use GetRecordingJobState to determine if data transfer is really taking place.
	// The only valid values for Mode shall be “Idle” and “Active”.
	Mode RecordingJobMode `xml:"tt:Mode,omitempty"`

	// This shall be a non-negative number. If there are multiple recording jobs that store data to
	// the same track, the device will only store the data for the recording job with the highest
	// priority. The priority is specified per recording job, but the device shall determine the priority
	// of each track individually. If there are two recording jobs with the same priority, the device
	// shall record the data corresponding to the recording job that was activated the latest.
	Priority int32 `xml:"tt:Priority,omitempty"`

	// Source of the recording.
	Source []RecordingJobSource `xml:"tt:Source,omitempty"`

	Extension RecordingJobConfigurationExtension `xml:"tt:Extension,omitempty"`

	// This attribute adds an additional requirement for activating the recording job.
	// If this optional field is provided the job shall only record if the schedule exists and is active.
	//

	ScheduleToken string `xml:"tt:ScheduleToken,attr,omitempty"`
}

// RecordingJobConfigurationExtension type
type RecordingJobConfigurationExtension struct {
}

// RecordingJobSource type
type RecordingJobSource struct {

	// This field shall be a reference to the source of the data. The type of the source
	// is determined by the attribute Type in the SourceToken structure. If Type is
	// http://www.onvif.org/ver10/schema/Receiver, the token is a ReceiverReference. In this case
	// the device shall receive the data over the network. If Type is
	// http://www.onvif.org/ver10/schema/Profile, the token identifies a media profile, instructing the
	// device to obtain data from a profile that exists on the local device.
	SourceToken SourceReference `xml:"tt:SourceToken,omitempty"`

	// If this field is TRUE, and if the SourceToken is omitted, the device
	// shall create a receiver object (through the receiver service) and assign the
	// ReceiverReference to the SourceToken field. When retrieving the RecordingJobConfiguration
	// from the device, the AutoCreateReceiver field shall never be present.
	AutoCreateReceiver bool `xml:"tt:AutoCreateReceiver,omitempty"`

	// List of tracks associated with the recording.
	Tracks []RecordingJobTrack `xml:"tt:Tracks,omitempty"`

	Extension RecordingJobSourceExtension `xml:"tt:Extension,omitempty"`
}

// RecordingJobTrack type
type RecordingJobTrack struct {

	// If the received RTSP stream contains multiple tracks of the same type, the
	// SourceTag differentiates between those Tracks. This field can be ignored in case of recording a local source.
	SourceTag string `xml:"tt:SourceTag,omitempty"`

	// The destination is the tracktoken of the track to which the device shall store the
	// received data.
	Destination TrackReference `xml:"tt:Destination,omitempty"`
}

// RecordingJobSourceExtension type
type RecordingJobSourceExtension struct {
}

// RecordingJobMode type
type RecordingJobMode string

// RecordingJobState type
type RecordingJobState string

// ModeOfOperation type
type ModeOfOperation string

// DeleteRecordingJob type
type DeleteRecordingJob struct {
	XMLName xml.Name `xml:"tt:DeleteRecordingJob"`

	// The token of the job to be deleted.
	JobToken RecordingJobReference `xml:"tt:JobToken,omitempty"`
}

// DeleteRecordingJobResponse type
type DeleteRecordingJobResponse struct {
	XMLName xml.Name `xml:"DeleteRecordingJobResponse"`
}

// GetRecordingJobs type
type GetRecordingJobs struct {
	XMLName xml.Name `xml:"tt:GetRecordingJobs"`
}

// GetRecordingJobsResponse type
type GetRecordingJobsResponse struct {
	XMLName xml.Name `xml:"GetRecordingJobsResponse"`

	// List of recording jobs.
	JobItem []GetRecordingJobsResponseItem `xml:"JobItem,omitempty"`
}

// GetRecordingJobsResponseItem type
type GetRecordingJobsResponseItem struct {
	JobToken RecordingJobReference `xml:"JobToken,omitempty"`

	JobConfiguration RecordingJobConfiguration `xml:"JobConfiguration,omitempty"`
}

// SetRecordingJobConfiguration type
type SetRecordingJobConfiguration struct {
	XMLName xml.Name `xml:"tt:SetRecordingJobConfiguration"`

	// Token of the job to be modified.
	JobToken RecordingJobReference `xml:"tt:JobToken,omitempty"`

	// New configuration of the recording job.
	JobConfiguration RecordingJobConfiguration `xml:"tt:JobConfiguration,omitempty"`
}

// SetRecordingJobConfigurationResponse type
type SetRecordingJobConfigurationResponse struct {
	XMLName xml.Name `xml:"SetRecordingJobConfigurationResponse"`

	// The JobConfiguration structure shall be the configuration
	// as it is used by the device. This may be different from the JobConfiguration passed to SetRecordingJobConfiguration.
	JobConfiguration RecordingJobConfiguration `xml:"JobConfiguration,omitempty"`
}

// GetRecordingJobConfiguration type
type GetRecordingJobConfiguration struct {
	XMLName xml.Name `xml:"tt:GetRecordingJobConfiguration"`

	// Token of the recording job.
	JobToken RecordingJobReference `xml:"tt:JobToken,omitempty"`
}

// GetRecordingJobConfigurationResponse type
type GetRecordingJobConfigurationResponse struct {
	XMLName xml.Name `xml:"GetRecordingJobConfigurationResponse"`

	// Current configuration of the recording job.
	JobConfiguration RecordingJobConfiguration `xml:"JobConfiguration,omitempty"`
}

// SetRecordingJobMode type
type SetRecordingJobMode struct {
	XMLName xml.Name `xml:"tt:SetRecordingJobMode"`

	// Token of the recording job.
	JobToken RecordingJobReference `xml:"tt:JobToken,omitempty"`

	// The new mode for the recording job.
	Mode RecordingJobMode `xml:"tt:Mode,omitempty"`
}

// SetRecordingJobModeResponse type
type SetRecordingJobModeResponse struct {
	XMLName xml.Name `xml:"SetRecordingJobModeResponse"`
}

// GetRecordingJobState type
type GetRecordingJobState struct {
	XMLName xml.Name `xml:"tt:GetRecordingJobState"`

	// Token of the recording job.
	JobToken RecordingJobReference `xml:"tt:JobToken,omitempty"`
}

// GetRecordingJobStateResponse type
type GetRecordingJobStateResponse struct {
	XMLName xml.Name `xml:"GetRecordingJobStateResponse"`

	// The current state of the recording job.
	State RecordingJobStateInformation `xml:"State,omitempty"`
}

// RecordingJobStateInformation type
type RecordingJobStateInformation struct {

	// Identification of the recording that the recording job records to.
	RecordingToken RecordingReference `xml:"tt:RecordingToken,omitempty"`

	// Holds the aggregated state over the whole RecordingJobInformation structure.
	State RecordingJobState `xml:"tt:State,omitempty"`

	// Identifies the data source of the recording job.
	Sources []RecordingJobStateSource `xml:"tt:Sources,omitempty"`

	Extension RecordingJobStateInformationExtension `xml:"tt:Extension,omitempty"`
}

// RecordingJobStateInformationExtension type
type RecordingJobStateInformationExtension struct {
}

// RecordingJobStateSource type
type RecordingJobStateSource struct {

	// Identifies the data source of the recording job.
	SourceToken SourceReference `xml:"tt:SourceToken,omitempty"`

	// Holds the aggregated state over all substructures of RecordingJobStateSource.
	State RecordingJobState `xml:"tt:State,omitempty"`

	// List of track items.
	Tracks RecordingJobStateTracks `xml:"tt:Tracks,omitempty"`
}

// RecordingJobStateTracks type
type RecordingJobStateTracks struct {
	Track []RecordingJobStateTrack `xml:"tt:Track,omitempty"`
}

// RecordingJobStateTrack type
type RecordingJobStateTrack struct {

	// Identifies the track of the data source that provides the data.
	SourceTag string `xml:"tt:SourceTag,omitempty"`

	// Indicates the destination track.
	Destination TrackReference `xml:"tt:Destination,omitempty"`

	// Optionally holds an implementation defined string value that describes the error.
	// The string should be in the English language.
	Error string `xml:"tt:Error,omitempty"`

	// Provides the job state of the track. The valid
	// values of state shall be “Idle”, “Active” and “Error”. If state equals “Error”, the Error field may be filled in with an implementation defined value.
	State RecordingJobState `xml:"tt:State,omitempty"`
}

// GetRecordingOptions type
type GetRecordingOptions struct {
	XMLName xml.Name `xml:"tt:GetRecordingOptions"`

	// Token of the recording.
	RecordingToken RecordingReference `xml:"tt:RecordingToken,omitempty"`
}

// GetRecordingOptionsResponse type
type GetRecordingOptionsResponse struct {
	XMLName xml.Name `xml:"GetRecordingOptionsResponse"`

	// Configuration of the recording.
	Options RecordingOptions `xml:"Options,omitempty"`
}

// ExportRecordedData type
type ExportRecordedData struct {
	XMLName xml.Name `xml:"tt:ExportRecordedData"`

	// Optional parameter that specifies start time for the exporting.
	StartPoint string `xml:"tt:StartPoint,omitempty"`

	// Optional parameter that specifies end time for the exporting.
	EndPoint string `xml:"tt:EndPoint,omitempty"`

	// Indicates the selection criterion on the existing recordings. .
	SearchScope SearchScope `xml:"tt:SearchScope,omitempty"`

	// Indicates which export file format to be used.
	FileFormat string `xml:"tt:FileFormat,omitempty"`

	// Indicates the target storage and relative directory path.
	StorageDestination StorageReferencePath `xml:"tt:StorageDestination,omitempty"`
}

// StorageReferencePath type
type StorageReferencePath struct {

	// identifier of an existing Storage Configuration.
	StorageToken ReferenceToken `xml:"tt:StorageToken,omitempty"`

	// gives the relative directory path on the storage
	RelativePath string `xml:"tt:RelativePath,omitempty"`

	Extension StorageReferencePathExtension `xml:"tt:Extension,omitempty"`
}

// StorageReferencePathExtension type
type StorageReferencePathExtension struct {
}

// ExportRecordedDataResponse type
type ExportRecordedDataResponse struct {
	XMLName xml.Name `xml:"ExportRecordedDataResponse"`

	// Unique operation token for client to associate the relevant events.
	OperationToken ReferenceToken `xml:"OperationToken,omitempty"`

	// List of exported file names. The device can also use AsyncronousOperationStatus event to publish this list.
	FileNames []string `xml:"FileNames,omitempty"`

	Extension struct {
	} `xml:"Extension,omitempty"`
}

// StopExportRecordedData type
type StopExportRecordedData struct {
	XMLName xml.Name `xml:"tt:StopExportRecordedData"`

	// Unique ExportRecordedData operation token
	OperationToken ReferenceToken `xml:"tt:OperationToken,omitempty"`
}

// StopExportRecordedDataResponse type
type StopExportRecordedDataResponse struct {
	XMLName xml.Name `xml:"StopExportRecordedDataResponse"`

	// Progress percentage of ExportRecordedData operation.
	Progress float32 `xml:"Progress,omitempty"`

	FileProgressStatus ArrayOfFileProgress `xml:"FileProgressStatus,omitempty"`
}

// ArrayOfFileProgress type
type ArrayOfFileProgress struct {

	// Exported file name and export progress information
	FileProgress []FileProgress `xml:"tt:FileProgress,omitempty"`

	Extension ArrayOfFileProgressExtension `xml:"tt:Extension,omitempty"`
}

// FileProgress type
type FileProgress struct {

	// Exported file name
	FileName string `xml:"tt:FileName,omitempty"`

	// Normalized percentage completion for uploading the exported file
	Progress float32 `xml:"tt:Progress,omitempty"`
}

// ArrayOfFileProgressExtension type
type ArrayOfFileProgressExtension struct {
}

// GetExportRecordedDataState type
type GetExportRecordedDataState struct {
	XMLName xml.Name `xml:"tt:GetExportRecordedDataState"`

	// Unique ExportRecordedData operation token
	OperationToken ReferenceToken `xml:"tt:OperationToken,omitempty"`
}

// GetExportRecordedDataStateResponse type
type GetExportRecordedDataStateResponse struct {
	XMLName xml.Name `xml:"GetExportRecordedDataStateResponse"`

	// Progress percentage of ExportRecordedData operation.
	Progress float32 `xml:"Progress,omitempty"`

	FileProgressStatus ArrayOfFileProgress `xml:"FileProgressStatus,omitempty"`
}

// Capabilities type
type Capabilities struct {

	// Indication if the device supports dynamic creation and deletion of recordings

	DynamicRecordings bool `xml:"tt:DynamicRecordings,attr,omitempty"`

	// Indication if the device supports dynamic creation and deletion of tracks

	DynamicTracks bool `xml:"tt:DynamicTracks,attr,omitempty"`

	// Indication which encodings are supported for recording. The list may contain one or more enumeration values of tt:VideoEncoding and tt:AudioEncoding. For encodings that are neither defined in tt:VideoEncoding nor tt:AudioEncoding the device shall use the  defintions. Note, that a device without audio support shall not return audio encodings.

	Encoding EncodingTypes `xml:"tt:Encoding,attr,omitempty"`

	// Maximum supported bit rate for all tracks of a recording in kBit/s.

	MaxRate float32 `xml:"tt:MaxRate,attr,omitempty"`

	// Maximum supported bit rate for all recordings in kBit/s.

	MaxTotalRate float32 `xml:"tt:MaxTotalRate,attr,omitempty"`

	// Maximum number of recordings supported. (Integer values only.)

	MaxRecordings float32 `xml:"tt:MaxRecordings,attr,omitempty"`

	// Maximum total number of supported recording jobs by the device.

	MaxRecordingJobs int32 `xml:"tt:MaxRecordingJobs,attr,omitempty"`

	// Indication if the device supports the GetRecordingOptions command.

	Options bool `xml:"tt:Options,attr,omitempty"`

	// Indication if the device supports recording metadata.

	MetadataRecording bool `xml:"tt:MetadataRecording,attr,omitempty"`

	//
	// Indication that the device supports ExportRecordedData command for the listed export file formats.
	// The list shall return at least one export file format value. The value of 'ONVIF' refers to
	// ONVIF Export File Format specification.
	//

	SupportedExportFileFormats onvif.StringAttrList `xml:"tt:SupportedExportFileFormats,attr,omitempty"`
}

// RecordingOptions type
type RecordingOptions struct {
	Job JobOptions `xml:"tt:Job,omitempty"`

	Track TrackOptions `xml:"tt:Track,omitempty"`
}

// JobOptions type
type JobOptions struct {

	// Number of spare jobs that can be created for the recording.

	Spare int32 `xml:"tt:Spare,attr,omitempty"`

	// A device that supports recording of a restricted set of Media/Media2 Service Profiles returns the list of profiles that can be recorded on the given Recording.

	CompatibleSources onvif.StringAttrList `xml:"tt:CompatibleSources,attr,omitempty"`
}

// TrackOptions type
type TrackOptions struct {

	// Total spare number of tracks that can be added to this recording.

	SpareTotal int32 `xml:"tt:SpareTotal,attr,omitempty"`

	// Number of spare Video tracks that can be added to this recording.

	SpareVideo int32 `xml:"tt:SpareVideo,attr,omitempty"`

	// Number of spare Aduio tracks that can be added to this recording.

	SpareAudio int32 `xml:"tt:SpareAudio,attr,omitempty"`

	// Number of spare Metadata tracks that can be added to this recording.

	SpareMetadata int32 `xml:"tt:SpareMetadata,attr,omitempty"`
}

// SearchScope type
type SearchScope struct {

	// A list of sources that are included in the scope. If this list is included, only data from one of these sources shall be searched.
	IncludedSources []SourceReference `xml:"tt:IncludedSources,omitempty"`

	// A list of recordings that are included in the scope. If this list is included, only data from one of these recordings shall be searched.
	IncludedRecordings []RecordingReference `xml:"tt:IncludedRecordings,omitempty"`

	// An xpath expression used to specify what recordings to search. Only those recordings with an RecordingInformation structure that matches the filter shall be searched.
	RecordingInformationFilter XPathExpression `xml:"tt:RecordingInformationFilter,omitempty"`

	// Extension point
	Extension SearchScopeExtension `xml:"tt:Extension,omitempty"`
}

// XPathExpression type
type XPathExpression string

// SourceReference type
type SourceReference struct {
	Token ReferenceToken `xml:"tt:Token,omitempty"`

	Type xsd.AnyURI `xml:"tt:Type,attr,omitempty"`
}

// RecordingReference type
type RecordingReference ReferenceToken

// TrackReference type
type TrackReference ReferenceToken

// ReferenceToken type
type ReferenceToken string

// SearchScopeExtension type
type SearchScopeExtension struct {
}
