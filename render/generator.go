package render

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/beevik/etree"
	"github.com/tutils/xmlschema/data"
	proto "github.com/tutils/xmlschema/output"
)

var GlobalNSMap = map[string]string{
	"http://www.w3.org/2001/XMLSchema":                                          "xsd",
	"http://www.w3.org/XML/1998/namespace":                                      "xml",
	"http://purl.org/dc/elements/1.1/":                                          "dc",
	"http://purl.org/dc/terms/":                                                 "dcterms",
	"http://purl.org/dc/dcmitype/":                                              "dcmitype",
	"http://schemas.openxmlformats.org/drawingml/2006/chart":                    "c",
	"http://schemas.openxmlformats.org/officeDocument/2006/relationships":       "r",
	"http://schemas.openxmlformats.org/drawingml/2006/main":                     "a",
	"http://schemas.openxmlformats.org/officeDocument/2006/sharedTypes":         "s",
	"http://schemas.openxmlformats.org/drawingml/2006/diagram":                  "dgm",
	"http://schemas.openxmlformats.org/drawingml/2006/picture":                  "dpct",
	"http://schemas.openxmlformats.org/drawingml/2006/lockedCanvas":             "dlc",
	"http://schemas.openxmlformats.org/drawingml/2006/chartDrawing":             "cdr",
	"http://schemas.openxmlformats.org/drawingml/2006/spreadsheetDrawing":       "xdr",
	"http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing":    "wp",
	"http://schemas.openxmlformats.org/wordprocessingml/2006/main":              "w",
	"http://schemas.openxmlformats.org/officeDocument/2006/math":                "m",
	"http://schemas.openxmlformats.org/schemaLibrary/2006/main":                 "sl",
	"http://schemas.openxmlformats.org/presentationml/2006/main":                "p",
	"http://schemas.openxmlformats.org/officeDocument/2006/characteristics":     "char",
	"http://schemas.openxmlformats.org/officeDocument/2006/bibliography":        "b",
	"http://schemas.openxmlformats.org/officeDocument/2006/customXml":           "cxml",
	"http://schemas.openxmlformats.org/officeDocument/2006/custom-properties":   "cp",
	"http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes":      "vt",
	"http://schemas.openxmlformats.org/officeDocument/2006/extended-properties": "ep",
	"http://schemas.openxmlformats.org/spreadsheetml/2006/main":                 "x",
	"urn:schemas-microsoft-com:vml":                                             "v",
	"urn:schemas-microsoft-com:office:office":                                   "ovml",
	"urn:schemas-microsoft-com:office:word":                                     "wvml",
	"urn:schemas-microsoft-com:office:excel":                                    "xvml",
	"urn:schemas-microsoft-com:office:powerpoint":                               "pvml",
	"http://schemas.openxmlformats.org/package/2006/content-types":              "pct",
	"http://schemas.openxmlformats.org/package/2006/metadata/core-properties":   "pcp",
	"http://schemas.openxmlformats.org/package/2006/digital-signature":          "pds",
	"http://schemas.openxmlformats.org/package/2006/relationships":              "pr",
	"http://example.org": "exam",
	"http://tutils.com":  "ts",
}

func ToTitleUpper(s string) string {
	return strings.ToUpper(string(s[0])) + s[1:]
}

func GenAllSymbolText() {
	gs := NewGlobalScope()
	gs.LoadSchema(DCSchemaLocation)
	gs.LoadSchema(DCTermsSchemaLocation)
	gs.LoadSchema(DCMITypeSchemaLocation)
	gs.LoadSchemaFilesFromDirectory("data/schemas/OfficeOpenXML-XMLSchema-Transitional")
	gs.LoadSchemaFilesFromDirectory("data/schemas/OpenPackagingConventions-XMLSchema")
	gs.LoadSchemaFilesFromDirectory("data/schemas/example")
	// // 生成定义顺序
	// orderCtx := newParseContext(gs)
	// orderCtx.parseRecursive = true
	// orderCtx.verbose = false

	// // 按顺序生成
	// genCtx := newParseContext(gs)
	// genCtx.parseRecursive = false
	// genCtx.verbose = true
	// for _, symb := range orderCtx.order.Order() {
	// 	info := orderCtx.order.MustGet(symb)
	// 	switch typ := symb.(type) {
	// 	case *proto.ComplexType:
	// 		r.ParseSchemaComplexType(typ, info.FileName)
	// 	case *proto.SimpleType:
	// 		r.ParseSchemaSimpleType(typ, info.FileName)
	// 	}
	// }

	// 按命名空间生成
	r := NewRenderer(gs)
	r.parseRecursive = false
	r.verbose = true

	defer ioredirect(stdout)()

	for _, ns := range gs.namespaceMap.Order() {
		u, err := url.Parse(ns)
		if err != nil {
			panic(err)
		}

		pth := path.Join(u.Host, u.Path)
		if len(pth) == 0 {
			pth = strings.ReplaceAll(ns, ":", "_")
		}

		baseDir := filepath.Join("output", strings.ReplaceAll(strings.ReplaceAll(pth, "/", "_"), ".", "_"))
		symbs := gs.namespaceMap.MustGet(ns)
		os.MkdirAll(baseDir, os.FileMode(0755))

		func() {
			if symbs.elementMap.Len() == 0 {
				return
			}
			fileName := filepath.Join(baseDir, "element.txt")
			fp, err := os.Create(fileName)
			if err != nil {
				panic(err)
			}
			defer fp.Close()
			os.Stdout = fp
			for _, name := range symbs.elementMap.Order() {
				symbol := symbs.elementMap.MustGet(name)
				r.ParseSchemaElement(symbol.Symbol, symbol.FileName)
			}
		}()

		func() {
			if symbs.groupMap.Len() == 0 {
				return
			}
			fileName := filepath.Join(baseDir, "group.txt")
			fp, err := os.Create(fileName)
			if err != nil {
				panic(err)
			}
			defer fp.Close()
			os.Stdout = fp
			for _, name := range symbs.groupMap.Order() {
				symbol := symbs.groupMap.MustGet(name)
				r.ParseSchemaGroup(symbol.Symbol, symbol.FileName)
			}
		}()

		func() {
			if symbs.attributeMap.Len() == 0 {
				return
			}
			fileName := filepath.Join(baseDir, "attribute.txt")
			fp, err := os.Create(fileName)
			if err != nil {
				panic(err)
			}
			defer fp.Close()
			os.Stdout = fp
			for _, name := range symbs.attributeMap.Order() {
				symbol := symbs.attributeMap.MustGet(name)
				r.ParseSchemaAttribute(symbol.Symbol, symbol.FileName)
			}
		}()

		func() {
			if symbs.attributeGroupMap.Len() == 0 {
				return
			}
			fileName := filepath.Join(baseDir, "attributeGroup.txt")
			fp, err := os.Create(fileName)
			if err != nil {
				panic(err)
			}
			defer fp.Close()
			os.Stdout = fp
			for _, name := range symbs.attributeGroupMap.Order() {
				symbol := symbs.attributeGroupMap.MustGet(name)
				r.ParseSchemaAttributeGroup(symbol.Symbol, symbol.FileName)
			}
		}()

		func() {
			if symbs.simpleTypeMap.Len() == 0 {
				return
			}
			fileName := filepath.Join(baseDir, "simpleType.txt")
			fp, err := os.Create(fileName)
			if err != nil {
				panic(err)
			}
			defer fp.Close()
			os.Stdout = fp
			for _, name := range symbs.simpleTypeMap.Order() {
				symbol := symbs.simpleTypeMap.MustGet(name)
				r.ParseSchemaSimpleType(symbol.Symbol, symbol.FileName)
			}
		}()

		func() {
			if symbs.complexTypeMap.Len() == 0 {
				return
			}
			fileName := filepath.Join(baseDir, "complexType.txt")
			fp, err := os.Create(fileName)
			if err != nil {
				panic(err)
			}
			defer fp.Close()
			os.Stdout = fp
			for _, name := range symbs.complexTypeMap.Order() {
				symbol := symbs.complexTypeMap.MustGet(name)
				r.ParseSchemaComplexType(symbol.Symbol, symbol.FileName)
			}
		}()
	}

	Debug()

	// 生成 SDK
	Generate(r, filepath.Join("output", "record.go"))
}

func ioredirect(fp *os.File) func() {
	save := os.Stdout
	os.Stdout = fp
	return func() {
		os.Stdout = save
	}
}

func Generate(r *Renderer, output string) {
	fp, err := os.Create(output)
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	defer ioredirect(fp)()

	for _, ns := range r.gs.namespaceMap.Order() {
		fmt.Printf("// %s: %s\n", GlobalNSMap[ns], ns)
	}
	fmt.Printf("\n")

	fmt.Printf("package proto\n\n")
	fmt.Printf("import \"github.com/beevik/etree\"\n\n")

	// 类型定义
	fallbackMap := make(map[string]struct{})
	for _, block := range r.defBlocks {
		if block.Fallback {
			blockPrefix := GlobalNSMap[block.NS]
			block.Name = strings.ReplaceAll(block.Name, "-", "_")
			block.CodeName = strings.ToUpper(blockPrefix) + "_" + block.Name
			fallbackMap[block.CodeName] = struct{}{}
		}
	}
	for _, block := range r.defBlocks {
		// fmt.Println(block.Name)
		blockPrefix := GlobalNSMap[block.NS]
		block.Name = strings.ReplaceAll(block.Name, "-", "_")

		block.CodeName = strings.ToUpper(blockPrefix) + "_" + block.Name
		if block.Fallback {
			// fmt.Printf("type %s = XMLElementRaw\n\n", block.CodeName)
			continue
		}

		// type define
		if !block.SimpleType {
			fmt.Printf("var _ XMLElement = (*%s)(nil)\n\n", block.CodeName)
			fmt.Printf("type %s struct {\n", block.CodeName)
			fmt.Printf("	base *XMLElementBase\n\n")
			for _, elem := range block.Elems.Order() {
				def := block.Elems.MustGet(elem)
				if def.Occurs > 1 {
					// fmt.Printf("    %s: %d (dup elem)\n", elem, def.Occurs)
				} else {
					// fmt.Printf("    %s: %d\n", elem, def.Occurs)
				}
				prefix := GlobalNSMap[def.NS]
				typePrefix := GlobalNSMap[def.TypeNS]

				def.Name = strings.ReplaceAll(def.Name, "-", "_")
				def.TypeName = strings.ReplaceAll(def.TypeName, "-", "_")

				def.CodeName = fmt.Sprintf("e_%s_%s", prefix, def.Name)
				if def.TypeName == "" {
					// 暂时也按照fallback逻辑走后，TODO：内嵌类型定义外置
					def.CodeTypeName = "XMLElementRaw"
				} else {
					def.CodeTypeName = fmt.Sprintf("%s_%s", strings.ToUpper(typePrefix), def.TypeName)
				}
				if _, ok := fallbackMap[def.CodeTypeName]; ok {
					// 降级类型转换
					def.CodeTypeName = "XMLElementRaw"
				}

				if def.MaxOccurs == -1 || def.MaxOccurs > 1 {
					fmt.Printf("    %s []*%s // %s\n", def.CodeName, def.CodeTypeName, def.Path)
				} else {
					fmt.Printf("    %s *%s // %s\n", def.CodeName, def.CodeTypeName, def.Path)
				}
			}
			for _, attr := range block.Attrs.Order() {
				def := block.Attrs.MustGet(attr)
				if def.Occurs > 1 {
					// fmt.Printf("    %s: %d (dup attr)\n", attr, def.Occurs)
				} else {
					// fmt.Printf("    %s: %d\n", attr, def.Occurs)
				}
				def.Name = strings.ReplaceAll(def.Name, "-", "_")
				def.TypeName = strings.ReplaceAll(def.TypeName, "-", "_")

				if def.NS == "" {
					// 不是引用
					def.CodeName = fmt.Sprintf("a_%s", def.Name)
				} else {
					// 引用
					prefix := GlobalNSMap[def.NS]
					def.CodeName = fmt.Sprintf("a_%s_%s", prefix, def.Name)
				}

				if def.TypeName == "" {
					// 没有类型信息，或者内嵌类型
					def.CodeTypeName = "XSD_string"
				} else {
					typePrefix := GlobalNSMap[def.TypeNS]
					def.CodeTypeName = fmt.Sprintf("%s_%s", strings.ToUpper(typePrefix), def.TypeName)
				}
				fmt.Printf("    %s *%s\n", def.CodeName, def.CodeTypeName)
			}
			fmt.Printf("}\n")
		} else {
			fmt.Printf("type %s string\n\n", block.CodeName)
		}

		// 构造函数
		if !block.SimpleType {
			fmt.Printf("func New%s(base *XMLElementBase) *%s {\n", block.CodeName, block.CodeName)
			fmt.Printf("	if base == nil {\n")
			fmt.Printf("		base = NewXMLElementBase()\n")
			fmt.Printf("	}\n")
			fmt.Printf("	e := &%s{base: base}\n", block.CodeName)
			fmt.Printf("	return e\n")
		} else {
			fmt.Printf("func New%s(v string) *%s {\n", block.CodeName, block.CodeName)
			fmt.Printf("	e := %s(v)\n", block.CodeName)
			fmt.Printf("	return &e\n")
		}
		fmt.Printf("}\n\n")

		if !block.SimpleType {
			// base
			fmt.Printf("func (e *%s) Base() *XMLElementBase {\n", block.CodeName)
			fmt.Printf("	return e.base\n")
			fmt.Printf("}\n\n")

			// MarshalXML
			fmt.Printf("func (e *%s) MarshalXML(ns string, tag string) *etree.Element {\n", block.CodeName)
			fmt.Printf("	eb := e.Base()\n")
			fmt.Printf("	tag = eb.AutoETreeTag(ns, tag)\n")
			fmt.Printf("	ee := etree.NewElement(tag)\n")
			fmt.Printf("	eb.Marshal(ee)\n")
			fmt.Printf("\n")
			for _, attr := range block.Attrs.Order() {
				def := block.Attrs.MustGet(attr)
				fmt.Printf("	if e.%s != nil {\n", def.CodeName)
				if def.NS == "" {
					fmt.Printf("		ee.CreateAttr(\"%s\", e.%s.String())\n", def.Name, def.CodeName)
				} else {
					fmt.Printf("		prefix := eb.MustGetPrefix(\"%s\")\n", def.NS)
					fmt.Printf("		ee.CreateAttr(prefix+\":%s\", e.%s.String())\n", def.Name, def.CodeName)
				}
				fmt.Printf("	}\n\n")
			}
			for _, elem := range block.Elems.Order() {
				def := block.Elems.MustGet(elem)
				if def.MaxOccurs == -1 || def.MaxOccurs > 1 {
					fmt.Printf("	for _, ce := range e.%s {\n", def.CodeName)
					if !def.SimpleType {
						fmt.Printf("		cee := ce.MarshalXML(\"%s\", \"%s\")\n", def.NS, def.Name)
					} else {
						fmt.Printf("		ctag := eb.AutoETreeTag(\"%s\", \"%s\")\n", def.NS, def.Name)
						fmt.Printf("		cee := etree.NewElement(ctag)\n")
						fmt.Printf("		cee.SetText(ce.String())\n")
					}
					fmt.Printf("		ee.AddChild(cee)\n")
					fmt.Printf("	}\n\n")
				} else {
					fmt.Printf("	if ce := e.%s; ce != nil {\n", def.CodeName)
					if !def.SimpleType {
						fmt.Printf("		cee := ce.MarshalXML(\"%s\", \"%s\")\n", def.NS, def.Name)
					} else {
						fmt.Printf("		ctag := eb.AutoETreeTag(\"%s\", \"%s\")\n", def.NS, def.Name)
						fmt.Printf("		cee := etree.NewElement(ctag)\n")
						fmt.Printf("		cee.SetText(ce.String())\n")
					}
					fmt.Printf("		ee.AddChild(cee)\n")
					fmt.Printf("	}\n\n")
				}
			}
			fmt.Printf("	return ee\n")
			fmt.Printf("}\n\n")

			// UnmarshalXML
			fmt.Printf("func (e *%s) UnmarshalXML(ee *etree.Element) {\n", block.CodeName)
			fmt.Printf("	eb := e.Base()\n")
			fmt.Printf("	if eb.parent == nil {\n")
			fmt.Printf("		eb.Unmarshal(ee)\n")
			fmt.Printf("	}\n\n")

			if block.Attrs.Len() > 0 {
				for _, attr := range block.Attrs.Order() {
					def := block.Attrs.MustGet(attr)
					fmt.Printf("	e.%s = nil\n", def.CodeName)
				}
				fmt.Printf("\n")

				fmt.Printf("	for _, attr := range ee.Attr {\n")
				for _, attr := range block.Attrs.Order() {
					def := block.Attrs.MustGet(attr)
					if def.NS == "" {
						fmt.Printf("		if attr.Key == \"%s\" && attr.Space == \"\" {\n", def.Name)

					} else {
						fmt.Printf("		if attr.Key == \"%s\" && attr.Space == eb.MustGetPrefix(\"%s\") {\n", def.Name, def.NS)
					}
					fmt.Printf("			e.%s = New%s(attr.Value)\n", def.CodeName, def.CodeTypeName)
					fmt.Printf("			continue\n")
					fmt.Printf("		}\n\n")
				}
				fmt.Printf("	}\n\n")
			}

			if block.Elems.Len() > 0 {
				fmt.Printf("	for _, child := range ee.Child {\n")
				fmt.Printf("		cee, ok := child.(*etree.Element)\n")
				fmt.Printf("		if !ok {\n")
				fmt.Printf("			continue\n")
				fmt.Printf("		}\n\n")
				fmt.Printf("		base := &XMLElementBase{}\n")
				fmt.Printf("		base.Unmarshal(cee)\n")
				fmt.Printf("		base.SetParent(e.base)\n\n")
				for _, elem := range block.Elems.Order() {
					def := block.Elems.MustGet(elem)
					fmt.Printf("		if base.VarifyETreeTag(cee, \"%s\", \"%s\") {\n", def.NS, def.Name)
					if !def.SimpleType {
						fmt.Printf("			ce := New%s(base)\n", def.CodeTypeName)
						// if _, ok := fallbackMap[def.CodeTypeName]; ok {
						// 	fmt.Printf("			ce := NewXMLElementRaw(base)\n")
						// } else {

						// }
						fmt.Printf("			ce.UnmarshalXML(cee)\n")
					} else {
						fmt.Printf("			ce := New%s(cee.Text())\n", def.CodeTypeName)
					}

					if def.MaxOccurs == -1 || def.MaxOccurs > 1 {
						fmt.Printf("			e.%s = append(e.%s, ce)\n", def.CodeName, def.CodeName)
					} else {
						fmt.Printf("			e.%s = ce\n", def.CodeName)
					}

					fmt.Printf("			continue\n")
					fmt.Printf("		}\n")
				}
				fmt.Printf("	}\n")
			}
			fmt.Printf("}\n\n")

			// SetAttrXxxx
			for _, attr := range block.Attrs.Order() {
				def := block.Attrs.MustGet(attr)
				var attrTitleName string
				if def.NS == "" {
					// 非引用属性
					attrTitleName = ToTitleUpper(def.Name)
				} else {
					prefix := GlobalNSMap[def.NS]
					attrTitleName = strings.ToUpper(prefix) + ToTitleUpper(def.Name)
				}
				fmt.Printf("func (e *%s) SetAttr%s(v %s) {\n", block.CodeName, attrTitleName, def.CodeTypeName)
				fmt.Printf("	e.%s = &v\n", def.CodeName)
				fmt.Printf("}\n\n")
			}

			// SetElemXxxx or AddElemXxxx
			for _, elem := range block.Elems.Order() {
				def := block.Elems.MustGet(elem)
				var elemTitleName string
				if block.NS == def.NS {
					// 本空间元素
					elemTitleName = ToTitleUpper(def.Name)
				} else {
					prefix := GlobalNSMap[def.NS]
					elemTitleName = strings.ToUpper(prefix) + ToTitleUpper(def.Name)
				}
				if def.MaxOccurs == -1 || def.MaxOccurs > 1 {
					// 数组
					fmt.Printf("func (e *%s) AddElem%s(ce *%s) {\n", block.CodeName, elemTitleName, def.CodeTypeName)
					if !def.SimpleType {
						fmt.Printf("	ce.Base().SetParent(e.Base())\n")
					}
					fmt.Printf("	e.%s = append(e.%s, ce)\n", def.CodeName, def.CodeName)
				} else {
					fmt.Printf("func (e *%s) SetElem%s(ce *%s) {\n", block.CodeName, elemTitleName, def.CodeTypeName)
					if !def.SimpleType {
						fmt.Printf("	ce.Base().SetParent(e.Base())\n")
					}
					fmt.Printf("	e.%s = ce\n", def.CodeName)
				}
				fmt.Printf("}\n\n")
			}
		} else {
			fmt.Printf("func (e %s) String() string {\n", block.CodeName)
			fmt.Printf("	return string(e)\n")
			fmt.Printf("}\n\n")
		}
	}
}

var stdout = os.Stdout

func Debug() {
	defer ioredirect(stdout)()
	// for def, c := range chrec2 {
	// 	if def.NS == XSNamespace {
	// 		continue
	// 	}
	// 	fmt.Printf("%s | %s: %d\n", def.Name, def.NS, c)
	// }

	// m := map[string]any{
	// 	"s": "\uf0fc",
	// }
	// en := json.NewEncoder(os.Stdout)
	// en.Encode(m)

	// http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Content-Type", "application/json")
	// 	en := json.NewEncoder(w)
	// 	en.Encode(m)
	// })
	// http.ListenAndServe(":8080", nil)
}

func Test() {
	// Test Unmarshal
	epsld := etree.NewDocument()
	fmt.Println(string(data.PSld))
	fmt.Println("--------------")
	epsld.ReadFromBytes(data.PSld)

	// psld := proto.NewP_CT_Slide(nil)
	psld := proto.NewXMLElementRaw(nil)
	psld.UnmarshalXML(epsld.Root())

	root := psld.MarshalXML(
		"http://schemas.openxmlformats.org/presentationml/2006/main",
		"sld",
	)
	epsld2 := etree.NewDocumentWithRoot(root)
	epsld2.Indent(4)
	epsld2.WriteTo(os.Stdout)
	fmt.Println("exit")
}
