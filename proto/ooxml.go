package proto

import (
    "encoding/xml"
)
// em
//
type Em = string
// a
//
// [Attributes]
//     href
//     name
type A struct {
    XMLName xml.Name

    Href string `xml:"href,attr,omitempty"`
    Name string `xml:"name,attr,omitempty"`

    XMLContent xml.CharData `xml:",chardata"`
}
// code
//
type Code = string
// p
//
// [Elements]
//     em
//     a
//     code
type P struct {
    XMLName xml.Name

    XMLContent xml.CharData `xml:",chardata"`
    Em *Em `xml:"em"`
    AList []*A `xml:"a"`
    CodeList []*Code `xml:"code"`
}
// h1
//
type H1 = string
// h3
//
type H3 = string
// h4
//
type H4 = string
// blockquote
//
// [Elements]
//     p
type Blockquote struct {
    XMLName xml.Name

    P *P `xml:"p"`
}
// h2
//
// [Elements]
//     a
type H2 struct {
    XMLName xml.Name

    A *A `xml:"a"`
}
// pre
//
type Pre = string
// li
//
// [Elements]
//     a
type Li struct {
    XMLName xml.Name

    A *A `xml:"a"`
}
// ul
//
// [Elements]
//     li
type Ul struct {
    XMLName xml.Name

    LiList []*Li `xml:"li"`
}
// div
//
// [Attributes]
//     class
//     id
// [Elements]
//     h1
//     div
//     p
//     h3
//     h4
//     blockquote
//     h2
//     pre
//     ul
type Div struct {
    XMLName xml.Name

    Class string `xml:"class,attr,omitempty"`
    Id string `xml:"id,attr,omitempty"`

    H1 *H1 `xml:"h1"`
    Div *Div `xml:"div"`
    PList []*P `xml:"p"`
    H3 *H3 `xml:"h3"`
    H4 *H4 `xml:"h4"`
    Blockquote *Blockquote `xml:"blockquote"`
    H2 *H2 `xml:"h2"`
    PreList []*Pre `xml:"pre"`
    Ul *Ul `xml:"ul"`
}
// documentation
//
// [Attributes]
//     source
// [Elements]
//     p
//     div
type Documentation struct {
    XMLName xml.Name

    Source string `xml:"source,attr,omitempty"`

    XMLContent xml.CharData `xml:",chardata"`
    PList []*P `xml:"p"`
    DivList []*Div `xml:"div"`
}
// hasFacet
//
// [Attributes]
//     name
type HasFacet struct {
    XMLName xml.Name

    Name string `xml:"name,attr,omitempty"`

}
// hasProperty
//
// [Attributes]
//     name
//     value
type HasProperty struct {
    XMLName xml.Name

    Name string `xml:"name,attr,omitempty"`
    Value string `xml:"value,attr,omitempty"`

}
// appinfo
//
// [Elements]
//     hasFacet
//     hasProperty
type Appinfo struct {
    XMLName xml.Name

    HasFacetList []*HasFacet `xml:"hasFacet"`
    HasPropertyList []*HasProperty `xml:"hasProperty"`
}
// annotation
//
// [Elements]
//     documentation
//     appinfo
type Annotation struct {
    XMLName xml.Name

    DocumentationList []*Documentation `xml:"documentation"`
    Appinfo *Appinfo `xml:"appinfo"`
}
// import
//
// [Attributes]
//     namespace
//     schemaLocation
//     id
// [Elements]
//     annotation
type Import struct {
    XMLName xml.Name

    Namespace string `xml:"namespace,attr,omitempty"`
    SchemaLocation string `xml:"schemaLocation,attr,omitempty"`
    Id string `xml:"id,attr,omitempty"`

    Annotation *Annotation `xml:"annotation"`
}
// enumeration
//
// [Attributes]
//     value
// [Elements]
//     annotation
type Enumeration struct {
    XMLName xml.Name

    Value string `xml:"value,attr,omitempty"`

    Annotation *Annotation `xml:"annotation"`
}
// minInclusive
//
// [Attributes]
//     value
//     id
type MinInclusive struct {
    XMLName xml.Name

    Value string `xml:"value,attr,omitempty"`
    Id string `xml:"id,attr,omitempty"`

}
// maxInclusive
//
// [Attributes]
//     value
//     id
type MaxInclusive struct {
    XMLName xml.Name

    Value string `xml:"value,attr,omitempty"`
    Id string `xml:"id,attr,omitempty"`

}
// pattern
//
// [Attributes]
//     value
//     id
// [Elements]
//     annotation
type Pattern struct {
    XMLName xml.Name

    Value string `xml:"value,attr,omitempty"`
    Id string `xml:"id,attr,omitempty"`

    XMLContent xml.CharData `xml:",chardata"`
    Annotation *Annotation `xml:"annotation"`
}
// minExclusive
//
// [Attributes]
//     value
type MinExclusive struct {
    XMLName xml.Name

    Value string `xml:"value,attr,omitempty"`

}
// maxExclusive
//
// [Attributes]
//     value
type MaxExclusive struct {
    XMLName xml.Name

    Value string `xml:"value,attr,omitempty"`

}
// length
//
// [Attributes]
//     value
//     fixed
type Length struct {
    XMLName xml.Name

    Value string `xml:"value,attr,omitempty"`
    Fixed string `xml:"fixed,attr,omitempty"`

}
// minLength
//
// [Attributes]
//     value
//     id
type MinLength struct {
    XMLName xml.Name

    Value string `xml:"value,attr,omitempty"`
    Id string `xml:"id,attr,omitempty"`

}
// maxLength
//
// [Attributes]
//     value
type MaxLength struct {
    XMLName xml.Name

    Value string `xml:"value,attr,omitempty"`

}
// anyAttribute
//
// [Attributes]
//     namespace
//     processContents
type AnyAttribute struct {
    XMLName xml.Name

    Namespace string `xml:"namespace,attr,omitempty"`
    ProcessContents string `xml:"processContents,attr,omitempty"`

}
// any
//
// [Attributes]
//     processContents
//     minOccurs
//     maxOccurs
//     namespace
type Any struct {
    XMLName xml.Name

    ProcessContents string `xml:"processContents,attr,omitempty"`
    MinOccurs string `xml:"minOccurs,attr,omitempty"`
    MaxOccurs string `xml:"maxOccurs,attr,omitempty"`
    Namespace string `xml:"namespace,attr,omitempty"`

}
// selector
//
// [Attributes]
//     xpath
type Selector struct {
    XMLName xml.Name

    Xpath string `xml:"xpath,attr,omitempty"`

}
// field
//
// [Attributes]
//     xpath
type Field struct {
    XMLName xml.Name

    Xpath string `xml:"xpath,attr,omitempty"`

}
// unique
//
// [Attributes]
//     name
// [Elements]
//     selector
//     field
type Unique struct {
    XMLName xml.Name

    Name string `xml:"name,attr,omitempty"`

    Selector *Selector `xml:"selector"`
    Field *Field `xml:"field"`
}
// key
//
// [Attributes]
//     name
// [Elements]
//     selector
//     field
type Key struct {
    XMLName xml.Name

    Name string `xml:"name,attr,omitempty"`

    Selector *Selector `xml:"selector"`
    Field *Field `xml:"field"`
}
// element
//
// [Attributes]
//     name
//     type
//     minOccurs
//     maxOccurs
//     ref
//     id
// [Elements]
//     annotation
//     unique
//     complexType
//     key
type Element struct {
    XMLName xml.Name

    Name string `xml:"name,attr,omitempty"`
    Type string `xml:"type,attr,omitempty"`
    MinOccurs string `xml:"minOccurs,attr,omitempty"`
    MaxOccurs string `xml:"maxOccurs,attr,omitempty"`
    Ref string `xml:"ref,attr,omitempty"`
    Id string `xml:"id,attr,omitempty"`

    Annotation *Annotation `xml:"annotation"`
    Unique *Unique `xml:"unique"`
    ComplexType *ComplexType `xml:"complexType"`
    KeyList []*Key `xml:"key"`
}
// group
//
// [Attributes]
//     name
//     ref
//     minOccurs
//     maxOccurs
// [Elements]
//     sequence
//     choice
//     annotation
type Group struct {
    XMLName xml.Name

    Name string `xml:"name,attr,omitempty"`
    Ref string `xml:"ref,attr,omitempty"`
    MinOccurs string `xml:"minOccurs,attr,omitempty"`
    MaxOccurs string `xml:"maxOccurs,attr,omitempty"`

    Sequence *Sequence `xml:"sequence"`
    Choice *Choice `xml:"choice"`
    Annotation *Annotation `xml:"annotation"`
}
// choice
//
// [Attributes]
//     minOccurs
//     maxOccurs
// [Elements]
//     element
//     group
//     sequence
//     choice
//     any
//     annotation
type Choice struct {
    XMLName xml.Name

    MinOccurs string `xml:"minOccurs,attr,omitempty"`
    MaxOccurs string `xml:"maxOccurs,attr,omitempty"`

    ElementList []*Element `xml:"element"`
    GroupList []*Group `xml:"group"`
    Sequence *Sequence `xml:"sequence"`
    Choice *Choice `xml:"choice"`
    Any *Any `xml:"any"`
    Annotation *Annotation `xml:"annotation"`
}
// sequence
//
// [Attributes]
//     minOccurs
//     maxOccurs
// [Elements]
//     any
//     element
//     choice
//     group
//     sequence
//     annotation
// type Sequence Skipped
// whiteSpace
//
// [Attributes]
//     value
//     id
//     fixed
type WhiteSpace struct {
    XMLName xml.Name

    Value string `xml:"value,attr,omitempty"`
    Id string `xml:"id,attr,omitempty"`
    Fixed string `xml:"fixed,attr,omitempty"`

}
// fractionDigits
//
// [Attributes]
//     value
//     fixed
//     id
type FractionDigits struct {
    XMLName xml.Name

    Value string `xml:"value,attr,omitempty"`
    Fixed string `xml:"fixed,attr,omitempty"`
    Id string `xml:"id,attr,omitempty"`

}
// restriction
//
// [Attributes]
//     base
// [Elements]
//     enumeration
//     minInclusive
//     maxInclusive
//     pattern
//     minExclusive
//     maxExclusive
//     length
//     minLength
//     maxLength
//     anyAttribute
//     sequence
//     attribute
//     group
//     annotation
//     whiteSpace
//     simpleType
//     fractionDigits
type Restriction struct {
    XMLName xml.Name

    Base string `xml:"base,attr,omitempty"`

    EnumerationList []*Enumeration `xml:"enumeration"`
    MinInclusive *MinInclusive `xml:"minInclusive"`
    MaxInclusive *MaxInclusive `xml:"maxInclusive"`
    PatternList []*Pattern `xml:"pattern"`
    MinExclusive *MinExclusive `xml:"minExclusive"`
    MaxExclusive *MaxExclusive `xml:"maxExclusive"`
    Length *Length `xml:"length"`
    MinLength *MinLength `xml:"minLength"`
    MaxLength *MaxLength `xml:"maxLength"`
    AnyAttribute *AnyAttribute `xml:"anyAttribute"`
    Sequence *Sequence `xml:"sequence"`
    AttributeList []*Attribute `xml:"attribute"`
    Group *Group `xml:"group"`
    Annotation *Annotation `xml:"annotation"`
    WhiteSpace *WhiteSpace `xml:"whiteSpace"`
    SimpleType *SimpleType `xml:"simpleType"`
    FractionDigits *FractionDigits `xml:"fractionDigits"`
}
// union
//
// [Attributes]
//     memberTypes
// [Elements]
//     simpleType
type Union struct {
    XMLName xml.Name

    MemberTypes string `xml:"memberTypes,attr,omitempty"`

    SimpleTypeList []*SimpleType `xml:"simpleType"`
}
// list
//
// [Attributes]
//     itemType
// [Elements]
//     simpleType
type List struct {
    XMLName xml.Name

    ItemType string `xml:"itemType,attr,omitempty"`

    SimpleType *SimpleType `xml:"simpleType"`
}
// simpleType
//
// [Attributes]
//     name
//     final
//     id
// [Elements]
//     restriction
//     union
//     list
//     annotation
type SimpleType struct {
    XMLName xml.Name

    Name string `xml:"name,attr,omitempty"`
    Final string `xml:"final,attr,omitempty"`
    Id string `xml:"id,attr,omitempty"`

    Restriction *Restriction `xml:"restriction"`
    Union *Union `xml:"union"`
    List *List `xml:"list"`
    Annotation *Annotation `xml:"annotation"`
}
// attribute
//
// [Attributes]
//     name
//     type
//     use
//     default
//     ref
//     form
// [Elements]
//     annotation
//     simpleType
type Attribute struct {
    XMLName xml.Name

    Name string `xml:"name,attr,omitempty"`
    Type string `xml:"type,attr,omitempty"`
    Use string `xml:"use,attr,omitempty"`
    Default string `xml:"default,attr,omitempty"`
    Ref string `xml:"ref,attr,omitempty"`
    Form string `xml:"form,attr,omitempty"`

    Annotation *Annotation `xml:"annotation"`
    SimpleType *SimpleType `xml:"simpleType"`
}
// attributeGroup
//
// [Attributes]
//     name
//     ref
// [Elements]
//     attribute
//     attributeGroup
//     annotation
type AttributeGroup struct {
    XMLName xml.Name

    Name string `xml:"name,attr,omitempty"`
    Ref string `xml:"ref,attr,omitempty"`

    AttributeList []*Attribute `xml:"attribute"`
    AttributeGroupList []*AttributeGroup `xml:"attributeGroup"`
    Annotation *Annotation `xml:"annotation"`
}
// all
//
// [Attributes]
//     minOccurs
// [Elements]
//     element
type All struct {
    XMLName xml.Name

    MinOccurs string `xml:"minOccurs,attr,omitempty"`

    ElementList []*Element `xml:"element"`
}
// extension
//
// [Attributes]
//     base
// [Elements]
//     attribute
//     sequence
//     choice
//     attributeGroup
//     group
type Extension struct {
    XMLName xml.Name

    Base string `xml:"base,attr,omitempty"`

    AttributeList []*Attribute `xml:"attribute"`
    Sequence *Sequence `xml:"sequence"`
    Choice *Choice `xml:"choice"`
    AttributeGroupList []*AttributeGroup `xml:"attributeGroup"`
    Group *Group `xml:"group"`
}
// simpleContent
//
// [Elements]
//     extension
type SimpleContent struct {
    XMLName xml.Name

    Extension *Extension `xml:"extension"`
}
// complexContent
//
// [Elements]
//     extension
//     restriction
type ComplexContent struct {
    XMLName xml.Name

    Extension *Extension `xml:"extension"`
    Restriction *Restriction `xml:"restriction"`
}
// complexType
//
// [Attributes]
//     name
//     mixed
//     abstract
// [Elements]
//     attribute
//     sequence
//     attributeGroup
//     choice
//     group
//     all
//     simpleContent
//     complexContent
//     annotation
//     anyAttribute
type ComplexType struct {
    XMLName xml.Name

    Name string `xml:"name,attr,omitempty"`
    Mixed string `xml:"mixed,attr,omitempty"`
    Abstract string `xml:"abstract,attr,omitempty"`

    AttributeList []*Attribute `xml:"attribute"`
    Sequence *Sequence `xml:"sequence"`
    AttributeGroupList []*AttributeGroup `xml:"attributeGroup"`
    Choice *Choice `xml:"choice"`
    Group *Group `xml:"group"`
    All *All `xml:"all"`
    SimpleContent *SimpleContent `xml:"simpleContent"`
    ComplexContent *ComplexContent `xml:"complexContent"`
    Annotation *Annotation `xml:"annotation"`
    AnyAttribute *AnyAttribute `xml:"anyAttribute"`
}
// include
//
// [Attributes]
//     schemaLocation
type Include struct {
    XMLName xml.Name

    SchemaLocation string `xml:"schemaLocation,attr,omitempty"`

}
// notation
//
// [Attributes]
//     name
//     public
//     system
type Notation struct {
    XMLName xml.Name

    Name string `xml:"name,attr,omitempty"`
    Public string `xml:"public,attr,omitempty"`
    System string `xml:"system,attr,omitempty"`

}
// schema
//
// [Attributes]
//     xsd
//     a
//     r
//     xmlns
//     cdr
//     s
//     targetNamespace
//     elementFormDefault
//     attributeFormDefault
//     blockDefault
//     w
//     dpct
//     p
//     vt
//     m
//     xdr
//     sl
//     wp
//     pvml
//     o
//     w10
//     x
//     v
//     xs
//     dc
//     dcterms
//     t
//     wvml
//     hfp
//     xhtml
//     xsi
//     schemaLocation
//     version
//     lang
// [Elements]
//     import
//     complexType
//     simpleType
//     group
//     element
//     attributeGroup
//     attribute
//     include
//     annotation
//     notation
type Schema struct {
    XMLName xml.Name

    Xmlns string `xml:"xmlns,attr,omitempty"`
    TargetNamespace string `xml:"targetNamespace,attr,omitempty"`
    ElementFormDefault string `xml:"elementFormDefault,attr,omitempty"`
    AttributeFormDefault string `xml:"attributeFormDefault,attr,omitempty"`
    BlockDefault string `xml:"blockDefault,attr,omitempty"`
    SchemaLocation string `xml:"schemaLocation,attr,omitempty"`
    Version string `xml:"version,attr,omitempty"`
    Lang string `xml:"lang,attr,omitempty"`
    XmlnsList []*xml.Attr `xml:",any,attr,omitempty"`

    ImportList []*Import `xml:"import"`
    ComplexTypeList []*ComplexType `xml:"complexType"`
    SimpleTypeList []*SimpleType `xml:"simpleType"`
    GroupList []*Group `xml:"group"`
    ElementList []*Element `xml:"element"`
    AttributeGroupList []*AttributeGroup `xml:"attributeGroup"`
    AttributeList []*Attribute `xml:"attribute"`
    IncludeList []*Include `xml:"include"`
    AnnotationList []*Annotation `xml:"annotation"`
    Notation *Notation `xml:"notation"`
}