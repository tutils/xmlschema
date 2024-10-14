package proto

import (
	"encoding/xml"
	"errors"
	"io"

	"github.com/tutils/xmlschema/boot"
)

var _ xml.Unmarshaler = (*Sequence)(nil)

// sequence
//
// [Attributes]
//
//	minOccurs
//	maxOccurs
//
// [Elements]
//
//	element
//	group
//	any
//	choice
//	sequence
//	annotation
type Sequence struct {
	XMLName xml.Name

	MinOccurs            string     `xml:"minOccurs,attr,omitempty"`
	MaxOccurs            string     `xml:"maxOccurs,attr,omitempty"`
	UnknownAttributeList []xml.Attr `xml:",any,attr"`

	Annotation         *Annotation        `xml:"annotation"`
	NestedParticleList []any              `xml:"-"`
	UnknownElementList []*boot.XMLElement `xml:",any"`
}

func (s *Sequence) unmarshalParticle(d *xml.Decoder, start *xml.StartElement, v any) error {
	if err := d.DecodeElement(v, start); err != nil {
		return err
	}
	s.NestedParticleList = append(s.NestedParticleList, v)
	return nil
}

// UnmarshalXML implements xml.Unmarshaler.
func (s *Sequence) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for {
		token, err := d.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		for _, attr := range start.Attr {
			switch attr.Name.Local {
			case "minOccurs":
				s.MinOccurs = attr.Value
			case "maxOccurs":
				s.MaxOccurs = attr.Value
			default:
				s.UnknownAttributeList = append(s.UnknownAttributeList, attr)
			}
		}

		if se, ok := token.(xml.StartElement); ok {
			switch se.Name.Local {
			case "element":
				var elem Element
				if err := s.unmarshalParticle(d, &se, &elem); err != nil {
					return err
				}
			case "group":
				var grp Group
				if err := s.unmarshalParticle(d, &se, &grp); err != nil {
					return err
				}
			case "choice":
				var cho Choice
				if err := s.unmarshalParticle(d, &se, &cho); err != nil {
					return err
				}
			case "sequence":
				var seq Sequence
				if err := s.unmarshalParticle(d, &se, &seq); err != nil {
					return err
				}
			case "any":
				var an Any
				if err := s.unmarshalParticle(d, &se, &an); err != nil {
					return err
				}
			case "annotation":
				var ann Annotation
				if err := d.DecodeElement(&ann, &se); err != nil {
					return err
				}
				s.Annotation = &ann
			default:
				var unknown boot.XMLElement
				if err := d.DecodeElement(&unknown, &se); err != nil {
					return err
				}
				s.UnknownElementList = append(s.UnknownElementList, &unknown)
			}
		}
	}
}
