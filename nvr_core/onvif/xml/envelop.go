package xml

import "encoding/xml"

// Envelope is a generic SOAP envelope for both Requests and Responses.
type Envelope[T any] struct {
	// The tag is simply "Envelope" to easily catch responses
	XMLName  xml.Name    `xml:"Envelope"`

	XmlnsS   string      `xml:"xmlns:s,attr,omitempty"`
	XmlnsTt  string      `xml:"xmlns:tt,attr,omitempty"`
	XmlnsTrt string      `xml:"xmlns:trt,attr,omitempty"`

	Header  *SOAPHeader `xml:"Header,omitempty"`
	Body    SOAPBody[T] `xml:"Body"`
}

type SOAPHeader struct {
	XMLName xml.Name `xml:"Header"`
}

// SOAPBody holds the dynamic generic content
type SOAPBody[T any] struct {
	// Added XMLName here so we can force the "s:Body" prefix later
	XMLName xml.Name `xml:"Body"`
	Content T
}