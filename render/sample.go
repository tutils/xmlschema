package render

import (
	"archive/zip"
	"fmt"
	"io"
	"unicode/utf16"

	"github.com/beevik/etree"
	"github.com/tutils/xmlschema/proto"
)

// 遍历某个文件夹下的全部schema文件
func LoadAllSchema() {
	gs := NewGlobalScope()
	// fs := gs.LoadSchema("data/schemas/ooxml/pml-slide.xsd")
	// prefix, symb, ok := fs.GetElement("p:sld")
	// fmt.Println(prefix, symb, ok)
	// fs = gs.LoadSchema("data/schemas/ooxml/pml-presentation.xsd")
	// prefix, symb, ok = fs.GetElement("p:presentation")
	// fmt.Println(prefix, symb, ok)
	// gs.LoadSchema(XSSchemaLocation)
	// gs.LoadSchema(XMLSchemaLocation)

	gs.LoadSchemaFilesFromDirectory("data/schemas/OfficeOpenXML-XMLSchema-Transitional")
	// for _, name := range gs.fileMap.Order() {
	// 	fs := gs.fileMap.MustGet(name)
	// 	if len(fs.schema.AttributeList) != 0 {
	// 		fmt.Println(fs.name)
	// 	}
	// }

	// symbol, ok := gs.GetElement(XSNamespace, "element")
	// if !ok {
	// 	panic("symbol not found")
	// }

	// r := newParseContext(gs)
	// ParseSchemaElement(symbol.Symbol, symbol.FileName)

	zr, err := zip.OpenReader("testlist.pptx")
	if err != nil {
		panic(err)
	}
	defer zr.Close()
	for _, file := range zr.Reader.File {
		fp, err := file.Open()
		if err != nil {
			panic(err)
		}
		defer fp.Close()
		// fmt.Println(file.Name)
		if file.Name == "ppt/slides/slide2.xml" {
			TestParse(fp, gs)
		}
	}

	fmt.Println("exit.")
}

func escapedStr(s string) string {
	var escapedStr string
	for _, r := range s {
		// 检查是否是高代理项的 Unicode 字符
		if r >= 0xD800 && r <= 0xDFFF {
			// 处理 UTF-16 代理对
			r1, r2 := utf16.EncodeRune(r)
			escapedStr += fmt.Sprintf("\\u%04x\\u%04x", r1, r2)
		} else {
			escapedStr += fmt.Sprintf("\\u%04x", r)
		}
	}
	return escapedStr
}

func TestParse(r io.ReadCloser, gs *GlobalScope) {
	doc := etree.NewDocument()
	_, err := doc.ReadFrom(r)
	if err != nil {
		panic(err)
	}

	root := doc.Root()
	root.ChildElements()

	prefixMap := make(map[string]string)
	for _, attr := range root.Attr {
		if attr.Space == "xmlns" {
			prefixMap[attr.Key] = attr.Value
		}
	}

	orderCtx := NewRenderer(gs)
	orderCtx.parseRecursive = true
	orderCtx.verbose = false
	ParseDocElement(orderCtx, root)
	sli := root.FindElements("//p:sld/p:cSld/p:spTree/p:sp/p:txBody/a:p/a:pPr[a:buChar]")
	fmt.Println(sli)
	for _, e := range sli {
		buFont := e.SelectElement("a:buFont")
		buChar := e.SelectElement("a:buChar")
		fmt.Println("typeface:", buFont.SelectAttr("typeface").Value)
		ch := buChar.SelectAttr("char").Value
		fmt.Println("char:", escapedStr(ch))
	}

	genCtx := NewRenderer(gs)
	genCtx.parseRecursive = false
	genCtx.verbose = true
	for _, symb := range orderCtx.order.Order() {
		info := orderCtx.order.MustGet(symb)
		switch typ := symb.(type) {
		case *proto.ComplexType:
			genCtx.ParseSchemaComplexType(typ, info.FileName)
		case *proto.SimpleType:
			genCtx.ParseSchemaSimpleType(typ, info.FileName)
		}
	}

	// fmt.Println(doc, root)
	// var root boot.XMLElement
	// dec := xml.NewDecoder(fp)
	// err := dec.Decode(&root)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(root)

	fmt.Println("end.")
}
