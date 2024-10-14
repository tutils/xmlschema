package boot

import (
	"encoding/xml"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/tutils/xmlschema/data"
	"github.com/tutils/xmlschema/tplcontainer"
)

// AttributeBlock 属性定义块
type AttributeBlock struct {
	Name    string
	IsXMLNS bool
}

// ElementBlock 元素定义块
type ElementBlock struct {
	Name    string
	IsArray bool
}

// 类型定义块
type TypeDefineBlock struct {
	Name    string
	Attrs   []*AttributeBlock
	Elems   []*ElementBlock
	IsMixed bool
}

// HasXMLNS 是否存在命名空间定义属性。
// 该函数一般在模板里使用，用于判断是否需要定义XmlnsList []*xml.Attr来接收全部命名空间定义属性
// 被接收的属性为xmlns:p、xmlns:a等属性，不包括xmlns属性
func (db *TypeDefineBlock) HasXMLNS() bool {
	for _, attr := range db.Attrs {
		if attr.IsXMLNS {
			return true
		}
	}
	return false
}

const (
	ForwardDelcCodeBlock = "ForwardDecl"
	DefineCodeBlock      = "Define"
)

// CodeBlock 代码块，包含了一系列类型的定义或声明
type CodeBlock struct {
	Type   string // ForwardDelcCodeBlock, DefineCodeBlock
	Define []*TypeDefineBlock
}

// BaseCodeFile 对应了一个代码文件的内容
type BaseCodeFile struct {
	Blocks []*CodeBlock
}

type definedState int

const (
	undefined         definedState = iota // 未定义
	walkingDependency                     // 正在遍历依赖
	defined                               // 已定义
)

type CheckCounter map[*XMLElement]int

// MaxCount 返回当前子元素在某个父类下最多出现过多少次
func (ca CheckCounter) MaxCount() int {
	var max int
	for _, c := range ca {
		if c > max {
			max = c
		}
	}
	return max
}

type attrInfo struct {
	isXMLNS bool
}

func newAttrInfo() *attrInfo {
	return &attrInfo{}
}

type elemInfo struct {
	counter CheckCounter
}

func newElemInfo() *elemInfo {
	return &elemInfo{
		counter: CheckCounter{},
	}
}

// 某个元素的属性和子元素记录
type record struct {
	// 属性顺序映射[属性名, 是否是命名空间定义属性]
	attrs *tplcontainer.SequenceMap[string, *attrInfo]

	// 子元素顺序映射[子元素名, 具体父元素的子元素计数器]
	elements *tplcontainer.SequenceMap[string, *elemInfo]

	isMixed bool
}

type Context struct {
	// 全部元素顺序映射[元素名]元素信息记录
	Types *tplcontainer.SequenceMap[string, *record]
}

func NewContext() *Context {
	return &Context{
		Types: tplcontainer.NewSequenceMap[string, *record](),
	}
}

// ScanAllSchemaFiles 遍历指定目录下的全部xml schema文件，用于探测schema中使用的元素的数据结构
func ScanAllSchemaFiles(ctx *Context, path string) {
	filepath.Walk(path, func(path string, info fs.FileInfo, _ error) error {
		if info.IsDir() || filepath.Ext(path) != ".xsd" {
			return nil
		}

		fp, err := os.Open(path)
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

		// 从根节点递归遍历
		walkElements(ctx, &schema, nil)
		return nil
	})
}

// walkElements 递归分析出全部类型（元素）的依赖关系
func walkElements(ctx *Context, e, p *XMLElement) {
	types := ctx.Types
	rec, ok := types.Get(e.XMLName.Local)
	if !ok {
		// 没记录过这个类型
		rec = &record{
			attrs:    tplcontainer.NewSequenceMap[string, *attrInfo](),
			elements: tplcontainer.NewSequenceMap[string, *elemInfo](),
		}
		types.Set(e.XMLName.Local, rec)
	}

	if p != nil {
		// 有父亲节点，计算这个子元素出现的次数。将来渲染成类型时，根据子元素最大出现次数判断该子元素是否为数组
		if elemInfo, ok := types.MustGet(p.XMLName.Local).elements.Get(e.XMLName.Local); !ok {
			// 这类父元素下，没出现过该类子元素
			types.MustGet(p.XMLName.Local).elements.Set(e.XMLName.Local, newElemInfo())
		} else {
			// 找到具体的父元素，将子元素出现次数+1。同类父元素分开计数，为了统计每个父元素一次性最多挂载多少个这种子元素
			elemInfo.counter[p]++
		}
	}

	// 遍历全部属性
	for _, a := range e.XMLAttrs {
		if _, ok := types.MustGet(e.XMLName.Local).attrs.Get(a.Name.Local); !ok {
			// 没记录过这个属性，同时标记该属性是否属于命名空间定义属性
			attrInfo := newAttrInfo()
			types.MustGet(e.XMLName.Local).attrs.Set(a.Name.Local, attrInfo)
			attrInfo.isXMLNS = (a.Name.Space == "xmlns")
		}
	}

	if len(e.XMLElements) == 0 && len(e.XMLContent) > 0 {
		rec.isMixed = true
	}

	// 递归遍历全部子元素
	for _, ch := range e.XMLElements {
		walkElements(ctx, ch, e)
	}
}

// makeDefineOrder 根据类型依赖记录types和已定义记录，确定代码定义顺序和前置声明顺序
// 通过递归遍历当前元素的全部子元素（将元素视为类型），如果出现环形引用，则将其输出至前置声明返回队列
func makeDefineOrder(ctx *Context, name string, defineRecord map[string]definedState) (order []string, forwardDelcOrder []string) {
	types := ctx.Types
	switch defineRecord[name] {
	case walkingDependency:
		// circle ref
		return nil, []string{name}
	case defined:
		return nil, nil
	}
	defineRecord[name] = walkingDependency
	defer func() {
		defineRecord[name] = defined
	}()

	if types.MustGet(name).elements.Len() == 0 {
		// 如果没有子元素，说明该类型不依赖其他类型，直接返回
		return []string{name}, nil
	}

	// 递归遍历所有子元素
	for _, elem := range types.MustGet(name).elements.Order() {
		elemOrder, elemFdOrder := makeDefineOrder(ctx, elem, defineRecord)
		order = append(order, elemOrder...)
		forwardDelcOrder = append(forwardDelcOrder, elemFdOrder...)
	}
	order = append(order, name)
	return order, forwardDelcOrder
}

// genBaseCodeFile 生成用于渲染模板的数据
func genBaseCodeFile(ctx *Context) *BaseCodeFile {
	baseCode := &BaseCodeFile{}

	// 定义状态
	defined := make(map[string]definedState)

	types := ctx.Types

	// 当前代码块，用于将类型一致的代码块合并
	var curCodeBlock *CodeBlock
	// 遍历全部元素
	for _, name := range types.Order() {
		// 当前元素及其依赖元素的定义顺序
		order, wdOrder := makeDefineOrder(ctx, name, defined)

		if len(wdOrder) > 0 {
			// 有前置声明列表，渲染数据中追加前置声明代码块
			curCodeBlock = &CodeBlock{
				Type: ForwardDelcCodeBlock,
			}
			baseCode.Blocks = append(baseCode.Blocks, curCodeBlock)

			// 根据前置声明序列，添加类型定义块
			for _, oname := range wdOrder {
				defBlock := &TypeDefineBlock{
					Name: oname,
				}
				curCodeBlock.Define = append(curCodeBlock.Define, defBlock)
			}
			// 因为前置声明序列是在分析某个类型的依赖时产生的。所以存在前置声明时，必会有一段类型定义序列。
			// 因此前置声明不会连续出现，所以可认为相同的代码块到此结束
			curCodeBlock = nil
		}

		if len(order) > 0 {
			// 有类型定义序列
			if curCodeBlock == nil {
				// 如果相同类型代码块中断，新建一个新的代码块
				curCodeBlock = &CodeBlock{
					Type: DefineCodeBlock,
				}
				baseCode.Blocks = append(baseCode.Blocks, curCodeBlock)
			}

			// 遍历类型定义序列
			for _, oname := range order {
				typ := types.MustGet(oname)
				defBlock := &TypeDefineBlock{
					Name:    oname,
					IsMixed: typ.isMixed,
				}

				if typ.attrs.Len() > 0 {
					// 给类型添加属性定义
					for _, attr := range typ.attrs.Order() {
						attrInfo := typ.attrs.MustGet(attr)
						attrBlock := &AttributeBlock{
							Name:    attr,
							IsXMLNS: attrInfo.isXMLNS,
						}
						defBlock.Attrs = append(defBlock.Attrs, attrBlock)
					}
				}

				if typ.elements.Len() > 0 {
					// 给类型添加元素定义
					for _, elem := range typ.elements.Order() {
						elemInfo := typ.elements.MustGet(elem)
						elemBlock := &ElementBlock{
							Name:    elem,
							IsArray: elemInfo.counter.MaxCount() > 1,
						}
						defBlock.Elems = append(defBlock.Elems, elemBlock)
					}
				}
				curCodeBlock.Define = append(curCodeBlock.Define, defBlock)
			}
		}
	}

	return baseCode
}

func Render(baseCode *BaseCodeFile, name string) error {
	tplFileName := filepath.Join("templates", name+".tpl")
	outFileName := filepath.Join("output", name)
	bs, err := data.Content.ReadFile(tplFileName)
	if err != nil {
		panic(err)
	}
	tpl := template.Must(template.New(tplFileName).Funcs(template.FuncMap{
		// 首字母大写函数
		"totilecase": func(s string) string {
			return strings.ToUpper(string(s[0])) + s[1:]
		},
	}).Parse(string(bs)))
	fp, err := os.Create(outFileName)
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	return tpl.Execute(fp, &baseCode)
}

func RenderAll(ctx *Context) {
	baseCode := genBaseCodeFile(ctx)

	tpls := []string{
		"ooxml.txt",
		"ooxml.go",
	}

	for _, tplName := range tpls {
		err := Render(baseCode, tplName)
		if err != nil {
			panic(err)
		}
	}
}
