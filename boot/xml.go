package boot

import (
	"encoding/xml"
	"fmt"
	"os"

	"github.com/beevik/etree"
	"github.com/tutils/xmlschema/proto"
)

type XMLElement struct {
	XMLName     xml.Name
	XMLAttrs    []*xml.Attr   `xml:",any,attr"`
	XMLContent  xml.CharData  `xml:",chardata"`
	XMLElements []*XMLElement `xml:",any"`
}

func Test() {
	fp, err := os.Open("schemas/ooxml/pml-slide.xsd")
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	dec := xml.NewDecoder(fp)
	var schema XMLElement
	err = dec.Decode(&schema)
	if err != nil {
		panic(err)
	}

	fmt.Println(schema)
	bs, err := xml.MarshalIndent(&schema, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bs))
}

func TestProto() {
	fp, err := os.Open("schemas/ooxml/pml-slide.xsd")
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	dec := xml.NewDecoder(fp)
	var schema proto.Schema
	err = dec.Decode(&schema)
	if err != nil {
		panic(err)
	}

	bs, err := xml.MarshalIndent(&schema, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bs))
}

func TestProtoETree() {
	doc := etree.NewDocument()
	err := doc.ReadFromFile("schemas/ooxml/pml-slide.xsd")
	if err != nil {
		panic(err)
	}

	doc.WriteTo(os.Stdout)
}
