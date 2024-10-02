package main

import (
	"encoding/xml"
	"fmt"

	"github.com/tutils/xmlschema/boot"
	"github.com/tutils/xmlschema/render"
)

type GroupShapeNonVisual struct {
}

type GroupShapeProperties struct {
}

var _ xml.Marshaler = (*GroupShape)(nil)

type GroupShape struct {
	XMLName              xml.Name
	GroupShapeNonVisual  GroupShapeNonVisual
	GroupShapeProperties GroupShapeProperties
	Name                 string
}

// MarshalXML implements xml.Marshaler.
func (g *GroupShape) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	fmt.Println(g)
	var data boot.XMLElement
	data.XMLAttrs = append(data.XMLAttrs,
		&xml.Attr{
			Name: xml.Name{
				Space: "",
				Local: "name",
			},
			Value: g.Name,
		},
	)
	data.XMLElements = append(data.XMLElements,
		&boot.XMLElement{
			XMLName: xml.Name{
				Space: "",
				Local: "nvGrpSpPr",
			},
			XMLContent: ([]byte)("GroupShapeNonVisual"),
		},
		&boot.XMLElement{
			XMLName: xml.Name{
				Space: "",
				Local: "grpSpPr",
			},
			XMLContent: ([]byte)("GroupShapeProperty"),
		},
	)
	return e.EncodeElement(&data, start)
}

type SlideTest struct {
	XMLName  xml.Name    ``
	XMLAttrs []*xml.Attr `xml:",any,attr"`

	Name       string     `xml:"name,attr"`
	GroupShape GroupShape `xml:"grpSp"`
}

func TestSample() {
	var grpSp GroupShape
	// grpSp.GroupShapeNonVisual = "nv"
	// grpSp.GroupShapeProperties = "sp"
	grpSp.Name = "GroupShape"

	var slide SlideTest
	slide.XMLName = xml.Name{Space: "http://pptxxx", Local: "p:sld"}
	slide.Name = "slideTest"
	slide.XMLAttrs = append(slide.XMLAttrs,
		&xml.Attr{
			Name: xml.Name{
				Space: "",
				Local: "xmlns",
			},
			Value: "http://pptxxx",
		},
		&xml.Attr{
			Name: xml.Name{
				Space: "http://pptxxx",
				Local: "p",
			},
			Value: "http://pptxxx",
		},
		&xml.Attr{
			Name: xml.Name{
				Space: "http://pptxxx",
				Local: "a",
			},
			Value: "http://aaaa",
		},
	)
	grpSp.XMLName.Space = "http://aaaa"

	slide.GroupShape = grpSp
	bs, err := xml.MarshalIndent(&slide, "", "   ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bs))
}

func main() {
	ctx := boot.NewContext()
	boot.ScanAllSchemaFiles(ctx)
	boot.RenderAll(ctx)
	// tree.TestProto()
	// tree.TestProtoETree()
	// tree.TestSample()
	render.LoadAllSchema()
}
