package wsdiscovery

import (
	"strings"

	"github.com/IOTechSystems/onvif/gosoap"
	"github.com/beevik/etree"
)

// BuildProbeMessage generates a SOAP ws-discovery Probe message
//
// Example Message:
//
//	<?xml version="1.0" encoding="UTF-8"?>
//	<soap-env:Envelope xmlns:soap-env="http://www.w3.org/2003/05/soap-envelope"
//				       xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing"
//				       xmlns:d="http://schemas.xmlsoap.org/ws/2005/04/discovery"
//				       xmlns:dn="http://www.onvif.org/ver10/network/wsdl"
//				       xmlns:soap-enc="http://www.w3.org/2003/05/soap-encoding">
//	   <soap-env:Header>
//		  <a:Action mustUnderstand="1">http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</a:Action>
//		  <a:MessageID>uuid:a277f13a-ecae-4492-9d6a-218982122d1c</a:MessageID>
//		  <a:To mustUnderstand="1">urn:schemas-xmlsoap-org:ws:2005:04:discovery</a:To>
//	   </soap-env:Header>
//	   <soap-env:Body>
//		  <d:Probe>
//			 <d:Types>dn:NetworkVideoTransmitter</d:Types>
//		  </d:Probe>
//	   </soap-env:Body>
//	</soap-env:Envelope>
func BuildProbeMessage(uuidV4 string, scopes, types []string, nmsp map[string]string) gosoap.SoapMessage {
	// Namespace List
	namespaces := make(map[string]string)
	namespaces["a"] = "http://schemas.xmlsoap.org/ws/2004/08/addressing"
	namespaces["d"] = "http://schemas.xmlsoap.org/ws/2005/04/discovery"
	namespaces["dn"] = "http://www.onvif.org/ver10/network/wsdl"

	probeMessage := gosoap.NewEmptySOAP()
	probeMessage.AddRootNamespaces(namespaces)
	if len(nmsp) != 0 {
		probeMessage.AddRootNamespaces(nmsp)
	}

	// Probe Header
	var headerContent []*etree.Element

	action := etree.NewElement("a:Action")
	action.SetText("http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe")
	action.CreateAttr("mustUnderstand", "1")

	msgID := etree.NewElement("a:MessageID")
	msgID.SetText("uuid:" + uuidV4)

	to := etree.NewElement("a:To")
	to.SetText("urn:schemas-xmlsoap-org:ws:2005:04:discovery")
	to.CreateAttr("mustUnderstand", "1")

	headerContent = append(headerContent, action, msgID, to)
	probeMessage.AddHeaderContents(headerContent)

	// Probe Body
	probe := etree.NewElement("d:Probe")

	if len(types) != 0 {
		typesTag := etree.NewElement("d:Types")
		typesTag.SetText(strings.Join(types, " "))
		probe.AddChild(typesTag)
	}

	if len(scopes) != 0 {
		scopesTag := etree.NewElement("d:Scopes")
		scopesTag.SetText(strings.Join(scopes, " "))
		probe.AddChild(scopesTag)
	}

	probeMessage.AddBodyContent(probe)

	return probeMessage
}
