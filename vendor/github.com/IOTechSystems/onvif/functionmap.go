package onvif

import (
	"fmt"
)

func FunctionByServiceAndFunctionName(serviceName, functionName string) (Function, error) {
	var functionMap map[string]Function

	switch serviceName {
	case DeviceWebService:
		functionMap = DeviceFunctionMap
	case MediaWebService:
		functionMap = MediaFunctionMap
	case Media2WebService:
		functionMap = Media2FunctionMap
	case PTZWebService:
		functionMap = PTZFunctionMap
	case EventWebService:
		functionMap = EventFunctionMap
	case AnalyticsWebService:
		functionMap = AnalyticsFunctionMap
	case ImagingWebService:
		functionMap = ImagingFunctionMap
	default:
		return nil, fmt.Errorf("the web service '%s' is not supported", serviceName)
	}

	if function, found := functionMap[functionName]; !found {
		return nil, fmt.Errorf("the web service '%s' does not support the function '%s'", serviceName, functionName)
	} else {
		return function, nil
	}
}
