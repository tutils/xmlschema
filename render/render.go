package render

import "fmt"

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

	gs.LoadSchemaFilesFromDirectory("data/schemas/ooxml")
	for _, name := range gs.fileMap.Order() {
		fs := gs.fileMap.MustGet(name)
		if len(fs.schema.AttributeList) != 0 {
			fmt.Println(fs.name)
		}
	}

	fmt.Println("exit.")
}
