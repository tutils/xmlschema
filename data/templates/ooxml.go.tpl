package proto

import (
    "encoding/xml"
)

{{- range $block := .Blocks}}
    {{- if eq $block.Type "ForwardDecl"}}
    {{- else}}
        {{- range $def := $block.Define}}
            {{- $titleCase := totilecase $def.Name}}
// {{$def.Name}}
//
            {{- if gt (len $def.Attrs) 0}}
// [Attributes]
                {{- range $attr := $def.Attrs}}
//     {{$attr.Name}}
                {{- end}}
            {{- end}}
            {{- if gt (len $def.Elems) 0}}
// [Elements]
                {{- range $elem := $def.Elems}}
//     {{$elem.Name}}
                {{- end}}
            {{- end}}
            {{- if eq $def.Name "sequence"}}
// type {{$titleCase}} Skipped
                {{- continue}}
            {{- end}}
            {{- if and (eq (len $def.Attrs) 0) (eq (len $def.Elems) 0)}}
type {{$titleCase}} = string
            {{- else}}
type {{$titleCase}} struct {
    XMLName xml.Name
{{""}}
                {{- if gt (len $def.Attrs) 0}}
                    {{- range $attr := $def.Attrs}}
                        {{- if not $attr.IsXMLNS}}
    {{totilecase $attr.Name}} string `xml:"{{$attr.Name}},attr,omitempty"`
                        {{- end}}
                    {{- end}}
                    {{- if $def.HasXMLNS }}
    XmlnsList []*xml.Attr `xml:",any,attr,omitempty"`
                    {{- end}}
{{""}}
                {{- end}}
                {{- if $def.IsMixed}}
    XMLContent xml.CharData `xml:",chardata"`
                {{- end}}
                {{- if gt (len $def.Elems) 0}}
                    {{- range $elem := $def.Elems}}
                        {{- if $elem.IsArray}}
    {{totilecase $elem.Name}}List []*{{totilecase $elem.Name}} `xml:"{{$elem.Name}}"`
                        {{- else}}
    {{totilecase $elem.Name}} *{{totilecase $elem.Name}} `xml:"{{$elem.Name}}"`
                        {{- end}}
                    {{- end}}
                {{- end}}
}
            {{- end}}
        {{- end}}
    {{- end}}
{{- end}}