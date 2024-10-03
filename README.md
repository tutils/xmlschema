# xmlschema

## 关于 targetNamespace 和 xmlns

- targetNamespace 是 schema 元素的属性，表示当前 schema 定义的约束所属的命名空间。
- xmlns 是任意元素的属性，表示元素的默认命名空间。即没有前缀（prefix）的元素都是 xmlns 指定的命名空间下的。
- 例如 data/schemas/xml/xml.xsd 中，targetNamespace 指向 [xml](http://www.w3.org/XML/1998/namespace)，xmlns 指向[xhtml](http://www.w3.org/1999/xhtml)，所以该 schema 所属的命名空间为 xml，而当前文件中所使用的<p>、<div>等没有前缀的元素均属于 xhtml 命名空间。
- 再例如下面一段 schema，其中 schema、element、simpleType、restriction 等没有前缀的元素的命名空间为 xmlns="http://www.w3.org/2001/XMLSchema"；其中schema中定义了一个simpleType：aGlobalType，aGlobalType所属targetNamespace="http://example.com/"。引用"http://example.com/" 下的 aGlobalType 时，需要使用前缀 tns。

```xml
<schema
  xmlns="http://www.w3.org/2001/XMLSchema"
  targetNamespace="http://example.com/"
  xmlns:tns="http://example.com/">

  <element name="aGlobalElement" type="tns:aGlobalType"/>

  <simpleType name="aGlobalType">
    <restriction base="string"/>
  </simpleType>
</schema>
```

## 符号表数据结构

```
NamespaceScope
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
```
