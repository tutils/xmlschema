package boot

import (
	"encoding/xml"
)

type XMLElement struct {
	XMLName     xml.Name
	XMLAttrs    []*xml.Attr   `xml:",any,attr"`
	XMLContent  xml.CharData  `xml:",chardata"`
	XMLElements []*XMLElement `xml:",any"`
}
