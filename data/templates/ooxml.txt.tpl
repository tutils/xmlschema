{{- range $block := .Blocks}}
<{{$block.Type}}>
    {{- if eq $block.Type "ForwardDecl"}}
        {{- range $def := $block.Define}}
    {{$def.Name}}
        {{- end}}
    {{- else}}
        {{- range $def := $block.Define}}
    {{$def.Name}}
            {{- if gt (len $def.Attrs) 0}}
        [Attributes]
                {{- range $attr := $def.Attrs}}
                    {{- if $attr.IsXMLNS}}
            {{$attr.Name}} (xmlns)
                    {{- else}}
            {{$attr.Name}}
                    {{- end}}

                {{- end}}
            {{- end}}
            {{- if gt (len $def.Elems) 0}}
        [Elements]
                {{- range $elem := $def.Elems}}
                    {{- if $elem.IsArray}}
            {{$elem.Name}} (array)
                    {{- else}}
            {{$elem.Name}}
                    {{- end}}
                {{- end}}
            {{- end}}
        {{- end}}
    {{- end}}
{{- end}}