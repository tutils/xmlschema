package render

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
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
	genCtx := NewRenderer(gs)
	genCtx.parseRecursive = false
	genCtx.verbose = true

	stdout := os.Stdout
	defer func() {
		os.Stdout = stdout
	}()

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
				genCtx.ParseSchemaElement(symbol.Symbol, symbol.FileName)
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
				genCtx.ParseSchemaGroup(symbol.Symbol, symbol.FileName)
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
				genCtx.ParseSchemaAttribute(symbol.Symbol, symbol.FileName)
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
				genCtx.ParseSchemaAttributeGroup(symbol.Symbol, symbol.FileName)
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
				genCtx.ParseSchemaSimpleType(symbol.Symbol, symbol.FileName)
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
				genCtx.ParseSchemaComplexType(symbol.Symbol, symbol.FileName)
			}
		}()
	}

	fp, err := os.Create(filepath.Join("output", "record.go"))
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	os.Stdout = fp

	for _, ns := range gs.namespaceMap.Order() {
		fmt.Println("// " + ns)
	}

	fmt.Println("package proto")
	for _, block := range genCtx.defBlocks {
		// fmt.Println(block.Name)
		blockPrefix := GlobalNSMap[block.NS]
		block.Name = strings.ReplaceAll(block.Name, "-", "_")
		fmt.Printf("type %s_%s struct {\n", strings.ToUpper(blockPrefix), block.Name)
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
			if def.TypeName == "" {
				if def.MaxOccurs == -1 || def.MaxOccurs > 1 {
					fmt.Printf("    e_%s_%s []*XMLElementRaw\n", prefix, def.Name)
				} else {
					fmt.Printf("    e_%s_%s *XMLElementRaw\n", prefix, def.Name)
				}
			} else {
				if def.MaxOccurs == -1 || def.MaxOccurs > 1 {
					fmt.Printf("    e_%s_%s []*%s_%s\n", prefix, def.Name, strings.ToUpper(typePrefix), def.TypeName)
				} else {
					fmt.Printf("    e_%s_%s *%s_%s\n", prefix, def.Name, strings.ToUpper(typePrefix), def.TypeName)
				}
			}

		}
		for _, attr := range block.Attrs.Order() {
			def := block.Attrs.MustGet(attr)
			if def.Occurs > 1 {
				// fmt.Printf("    %s: %d (dup attr)\n", attr, def.Occurs)
			} else {
				// fmt.Printf("    %s: %d\n", attr, def.Occurs)
			}
			typePrefix := GlobalNSMap[def.TypeNS]
			def.Name = strings.ReplaceAll(def.Name, "-", "_")
			def.TypeName = strings.ReplaceAll(def.TypeName, "-", "_")
			if def.TypeName == "" {
				fmt.Printf("    a_%s *string\n", def.Name)
			} else if def.NS == "" {
				fmt.Printf("    a_%s *%s_%s\n", def.Name, strings.ToUpper(typePrefix), def.TypeName)
			} else {
				prefix := GlobalNSMap[def.NS]
				fmt.Printf("    a_%s_%s *%s_%s\n", prefix, def.Name, strings.ToUpper(typePrefix), def.TypeName)
			}
		}
		fmt.Println("}\n")
	}
}
