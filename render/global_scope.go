package render

import (
	"io/fs"
	"net/url"
	"path/filepath"

	"github.com/tutils/xmlschema/proto"
	"github.com/tutils/xmlschema/tplcontainer"
)

/*

NamespaceMap {
	Namespace -> SymbolMap {
		SymbolName -> SymbolInfo {
			proto.Symbol
			FileName
		}
	}
}

FileMap {
    FileName -> FileScope {
        Namespace
        PrefixMap {
            Prefix -> Namespace
        }
        SymbolList
    }
}
*/

type loadState int

const (
	unloaded loadState = iota // 未加载
	loading                   // 加载中
	loaded                    // 已加载
)

type GlobalScope struct {
	namespaceMap *tplcontainer.SequenceMap[string, *SymbolMap]
	fileMap      *tplcontainer.SequenceMap[string, *FileScope]

	// [file, record]
	cache FileCache
}

func NewGlobalScope() *GlobalScope {
	gs := &GlobalScope{
		namespaceMap: tplcontainer.NewSequenceMap[string, *SymbolMap](),
		fileMap:      tplcontainer.NewSequenceMap[string, *FileScope](),

		cache: make(FileCache),
	}

	for _, fs := range []*FileScope{
		xmlFs,
		xsFs,
	} {
		gs.cache[fs.name] = fs.schema

		// 添加文件映射
		gs.fileMap.Set(fs.name, fs)

		// 将文件域的符号表合并至全局符号表中
		for _, ns := range fs.namespaceMap.Order() {
			fileSymbs := fs.namespaceMap.MustGet(ns)
			globalSymbs, ok := gs.namespaceMap.Get(ns)
			if !ok {
				globalSymbs = NewSymbolMap()
				gs.namespaceMap.Set(ns, globalSymbs)
			}
			globalSymbs.Merge(fileSymbs, false)
		}
	}

	return gs
}

func (gs *GlobalScope) LoadSchema(name string) *FileScope {
	u, err := url.Parse(name)
	if err != nil {
		panic(err)
	}
	if len(u.Scheme) == 0 {
		// 文件名标准化
		dir := filepath.Dir(name)
		name = filepath.Join(dir, filepath.Base(name))
	}

	if fs, ok := gs.fileMap.Get(name); ok {
		// 已经加载过该文件
		return fs
	}

	fs := newFileScope()
	fs.loadSchema(name, gs.cache)

	// 添加文件映射
	gs.fileMap.Set(name, fs)

	// 将文件域的符号表合并至全局符号表中
	for _, ns := range fs.namespaceMap.Order() {
		fileSymbs := fs.namespaceMap.MustGet(ns)
		globalSymbs, ok := gs.namespaceMap.Get(ns)
		if !ok {
			globalSymbs = NewSymbolMap()
			gs.namespaceMap.Set(ns, globalSymbs)
		}
		globalSymbs.Merge(fileSymbs, false)
	}

	return fs
}

func (gs *GlobalScope) LoadSchemaFilesFromDirectory(dir string) {
	filepath.Walk(dir, func(path string, info fs.FileInfo, _ error) error {
		if info.IsDir() {
			return nil
		}

		gs.LoadSchema(path)

		return nil
	})
}

func (gs *GlobalScope) GetSimpleType(namespace string, name string) (*Symbol[*proto.SimpleType], bool) {
	symbs, ok := gs.namespaceMap.Get(namespace)
	if !ok {
		return nil, false
	}
	return symbs.GetSimpleType(name)
}

func (gs *GlobalScope) GetComplexType(namespace string, name string) (*Symbol[*proto.ComplexType], bool) {
	symbs, ok := gs.namespaceMap.Get(namespace)
	if !ok {
		return nil, false
	}
	return symbs.GetComplexType(name)
}

func (gs *GlobalScope) GetType(namespace string, name string) (*Symbol[any], bool) {
	symbs, ok := gs.namespaceMap.Get(namespace)
	if !ok {
		return nil, false
	}
	return symbs.GetType(name)
}

func (gs *GlobalScope) GetElement(namespace string, name string) (*Symbol[*proto.Element], bool) {
	symbs, ok := gs.namespaceMap.Get(namespace)
	if !ok {
		return nil, false
	}
	return symbs.GetElement(name)
}

func (gs *GlobalScope) GetGroup(namespace string, name string) (*Symbol[*proto.Group], bool) {
	symbs, ok := gs.namespaceMap.Get(namespace)
	if !ok {
		return nil, false
	}
	return symbs.GetGroup(name)
}

func (gs *GlobalScope) GetAttribute(namespace string, name string) (*Symbol[*proto.Attribute], bool) {
	symbs, ok := gs.namespaceMap.Get(namespace)
	if !ok {
		return nil, false
	}
	return symbs.GetAttribute(name)
}

func (gs *GlobalScope) GetAttributeGroup(namespace string, name string) (*Symbol[*proto.AttributeGroup], bool) {
	symbs, ok := gs.namespaceMap.Get(namespace)
	if !ok {
		return nil, false
	}
	return symbs.GetAttributeGroup(name)
}
