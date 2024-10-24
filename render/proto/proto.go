package proto

import (
	"fmt"
	"os"

	"github.com/beevik/etree"
	"github.com/tutils/xmlschema/data"
)

/*
重点解决的问题：
1. 如何解决重名元素/属性
2. 如何搜寻命名空间定义
3. 对于group、attributeGroup的处理方式
4. choice和sequence，打平后，元素之间互斥判断

1. 成员前缀解决跨空间引用group和attributeGroup时，展开token后名字相同问题
2. xmlns成员，前缀与命名空间的映射

元素，引用和非引用，前缀规则一致。

属性则遵循以下规则：
1. 非引用属性，必须没有前缀
2. 引用属性，必须有前缀，哪怕引用的为当前空间的属性

组、属性组，对于引用的和非引用的，先将其展开，规则与单一元素和单一属性一致。
1. 对于属性组的引用，完全等价于直接展开；无论是本空间的还是其他空间的，均视作本空间的属性
2. 如果属性组的引用，属性组里包含引用属性，则直接视为普通引用属性

group定义、group/sequence、group/choice 不允许出现minOccurs或maxOccurs属性
group引用允许出现minOccurs或maxOccurs属性

对于sequence引用的多个group中存在同名的元素的情况，对于可能造成歧义的可选元素，一律按照本轮排序最前的group算，不会回溯判断

如果sequence.maxOccurs > 1 全部使用Add方法，否则使用Set方法

group直接展开后，是否有重名元素
	1. 无：每个元素渲染set/get方法
	2. 有：使用兜底结构（xml基本操作方法）

sequence
	1. maxOccurs<=1：每个元素渲染set/get方法
	2. maxOccurs>1
		2.1 子元素只有一个：等价于一个元素的数组
		2.2 子元素不止一个：使用兜底结构（xml基本操作方法）

choice
	1. 只定义类型下的第一级choice，并按序号命名
	2. 元素要标明属于哪个一级choice
	2. 如果父级choice为unbounded，递归子元素全部映射Add方法

Set() {
	root choice1
		A
		B(unbounded)
		choice2 (unbounded)
			C
			choice3
				D
		choice4
			E
			sequence1
				F
				G
				choice5
					H
					I
		
	c1/c2
	c1/c2/c3
	c1/c4
	c1/c4/s1
	c1/c4/s1/c5
	c1/c2/c3/D
	c1/c4/E
	c1/c4/s1/F
	c1/c4/s1/G
	c1/c4/s1/c5/H
	c1/c4/s1/c5/I

	c1/c2/c3
	c1/c4/s1/G
	c1/c4/s1/c5/H

	有一个多叉树状的数据结构，其中叶子节点为实际数据使用大写字母表示；非叶子节为容器节点用小写字母表示，非叶子节点有两种：c和s。其中c下的子节点均为互斥关系，s下的子节点为非互斥关系。节点可以通过路径表示，例如：
	c1/c2
	c1/c2/c3
	c1/c4
	c1/c4/s1
	c1/c4/s1/c5
	c1/c2/c3/D
	c1/c4/E
	c1/c4/s1/F
	c1/c4/s1/G
	c1/c4/s1/c5/H
	c1/c4/s1/c5/I
	如果想判断任意叶子节点之间是否互斥，应该怎么做？
}

TODO:
1. 对any和"lax"的处理
2. 对sequence下有多个any情况下做特殊处理，比如把sequence做个临时命名

识别sequence只有一个元素的数组类型
识别元素重复导致的降级
测试生成
识别choice数组

*/

// Unmarshal 步骤
// 1. 解析属性
// 1.1 解析命名空间定义

// 1.2 解析其他属性值
// 1.2.1 其他属性值初始化

// 1.2.2 解析其他属性值
// 2.解析元素
// 2.1 元素初始化
// 2.2 解析元素
// [3.解析文本]

func TestMarshal() {
	eperson := NewT_CT_Person(nil)

	e_t_name := NewT_CT_Name(nil)
	e_t_name.Base().AddXMLNS("x", "http://tutils.com")
	e_t_name.SetAttrEn("t5w0rd")
	eperson.SetElemName(e_t_name)

	e_t_remark := NewT_CT_Doc(nil)
	e_t_remark.SetText("注释")
	eperson.SetElemRemark(e_t_remark)

	var e_t_arrName *T_CT_Name
	e_t_arrName = NewT_CT_Name(nil)
	e_t_arrName.SetAttrEn("xxx")
	eperson.AddElemArrName(e_t_arrName)
	e_t_arrName = NewT_CT_Name(nil)
	e_t_arrName.SetAttrEn("yyy")
	eperson.AddElemArrName(e_t_arrName)

	e_t_defbyseq := NewT_CT_DefBySeq(nil)
	var doc *T_CT_Doc
	doc = NewT_CT_Doc(nil)
	doc.SetText("aaa")
	e_t_defbyseq.AddElemA(doc)
	doc = NewT_CT_Doc(nil)
	doc.SetText("ccc")
	e_t_defbyseq.AddElemC(doc)
	doc = NewT_CT_Doc(nil)
	doc.SetText("bbb")
	e_t_defbyseq.AddElemB(doc)
	doc = NewT_CT_Doc(nil)
	doc.SetText("ddd")
	e_t_defbyseq.AddElemD(doc)
	eperson.SetElemDefbyseq(e_t_defbyseq)

	eperson.SetAttrSex("male")
	eperson.SetAttrAge("18")
	eperson.SetAttrTURL("https://www.baidu.com")

	// xmlns
	eperson.Base().AddXMLNS("", "http://example.org")
	eperson.Base().AddXMLNS("e", "http://example.org")
	eperson.Base().AddXMLNS("t", "http://tutils.com")
	root := eperson.MarshalXML("http://example.org", "person")

	e_per_doc := etree.NewDocumentWithRoot(root)
	e_per_doc.Indent(4)
	e_per_doc.WriteTo(os.Stdout)

	fmt.Println("exit")
}

func TestUnmarshal() {
	e_per_doc := etree.NewDocument()
	fmt.Println(string(data.EPerson))
	fmt.Println("--------------")
	e_per_doc.ReadFromBytes(data.EPerson)

	eperson := NewT_CT_Person(nil)
	eperson.UnmarshalXML(e_per_doc.Root())

	root := eperson.MarshalXML("http://example.org", "person")
	e_per_doc_2 := etree.NewDocumentWithRoot(root)
	e_per_doc_2.Indent(4)
	e_per_doc_2.WriteTo(os.Stdout)

	fmt.Println("exit", eperson)
}
