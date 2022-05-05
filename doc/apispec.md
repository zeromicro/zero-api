[TOC]

# 介绍(Introduction)
zero-api 是一种声明 HTTP API 的语言，他可以通过 goctl 翻译成基于 go-zero 的服务代码。

# 标记(Notation)
语法使用 Extended Backus-Naur Form (EBNF)指定, 实例如下:

```EBNF
Production  = production_name "=" [ Expression ] "." .
Expression  = Alternative { "|" Alternative } .
Alternative = Term { Term } .
Term        = production_name | token [ "…" token ] | Group | Option | Repetition .
Group       = "(" Expression ")" .
Option      = "[" Expression "]" .
Repetition  = "{" Expression "}" .
```

结果是由术语和以下操作符构造的表达式，优先级递增:

```EBNF
|   交替
()  分组
[]  可选（0 或 1 次）
{}  重复（0 到 n 次）
```

# 源码表示(Source code representation)
原代码需要使用 UTF-8 编码，我们会区分大小写。

## 字符(characters)
以下术语用来表示特定的 Unicode 字符类：

```EBNF
newline        = /* Unicode 码位U+000A */ .
unicode_char   = /* 除了换行符的其他任意 Unicode 码位 */ .
unicode_letter = /* 分类为 “Letter” 的 Unicode 码位 */ .
unicode_digit  = /* 分类为“Number, decimal digit”的Unicode码位 */ .
```

## 字母与数字(Letters and digits)
下划线 _(U+005F) 被认为字母。

```EBNF
letter        = unicode_letter | "_" .
decimal_digit = "0" … "9" .
binary_digit  = "0" | "1" .
octal_digit   = "0" … "7" .
hex_digit     = "0" … "9" | "A" … "F" | "a" … "f" .
```

# 词法元素
## 注释
注释作为文档，有 1 种格式：

1. 单行注释以 **//** 开头，并且在行尾停止。

注释不能在 **rune** 和 **string literal** 中出现、

## Tokens
Tokens 组成了 API 语言的词汇表。有四个分类： 标识符 、 关键字 、 运算符和标点以及字面值 。 空白 是由空格（U+0020）、水平制表（U+0009）、回车（U+000D）和新行（U+000A）所组成的，空白一般会被忽略，除非它分隔了组合在一起会形成单一 token 的 tokens. 并且，新行或者文件结尾可能会触发 分号 的插入。当把输入的内容区分为 tokens 时，每一个 token 都是可组成有效 token 的最长字符序列。

## 分号
正式的语法使用分号 **;** 作为一定数量的语句终结符。

zero-api 会有一定规则来省略分号：
1. 输入的内容被分为 tokens 时，当每一行最后一个 token 为以下 token 时，一个分号会自动插入到其后面
2. 为了使复杂的语句可以占据在单一一行上，分号也可以在关闭的 ) 或者 } 前被省略

为了反应出惯用的使用习惯，本文档中的代码示例将参照这些规则来省略掉分号。

## 标识符(Identifiers)
标识符用于命名程序中的实体——比如变量和类型。它是一个或者多个字母和数字的序列组合。标识符的第一个字符必须是一个字母。

```EBNT
identifier = letter { letter | unicode_digit } .
```

```EBNT_DEMO
a
_x9
ThisVariableIsExported
αβ
```

有一些标识符已经被 预先声明 了。

## 字符串字面值(String literals)
字符串字面值代表了通过串联字符序列而获得的字符串 常量 。它有两种形式： 原始 raw 字符串字面值和 解释型 interpreted 字符串字面值。

```EBNT
string_lit             = raw_string_lit | interpreted_string_lit | identifier .
raw_string_lit         = "`" { unicode_char | newline } "`" .
interpreted_string_lit = `"` { unicode_value | byte_value } `"` .
```

```EBNT_DEMO
`abc`                // 同 "abc"
`\n
\n`                  // 同 "\\n\n\\n"
"\n"
"\""                 // 同 `"`
"Hello, world!\n"
"日本語"
"\u65e5本\U00008a9e"
"\xff\u00FF"
"\uD800"             // 非法: 代理了一半
"\U00110000"         // 非法: 无效的 Unicode 码位
```

## zero-api 特殊字符定义
zero-api 支持一批特殊字符变量，可以以特殊字符开头，数字等开头。
```EBNT
value_string_lit = unicode_char { unicode_char } .
```

```EBNT_DEMO
3s
abc
/api/user/:info
```

TODO:
针对
/api/user/:info(req)

这种需要解析成 **/api/user/:info**，**(**，**)** 和 **req** 这样的token。

## 预声明的标识符
以下是 zero-api 支持的预声明的标识符

```EBNT
Types:
bool float32 float64 int int8 int16 int32 int64
string uint uint8 uint16 uint32 uint64

```

```EBNT_DEMO
type User {
    Id string
    Age int
    Name string
}
```

# 源文件组织(Source file organization)
每个源文件都是由以下的组成：

```EBNT
SourceFile = SyntaxDecl ";" { ImportDecl ";" } [ InfoDecl ";" ] { TypeDecl | ServiceDecl ":"}
```

## 语法版本(syntax)
语法版本控制 API 的语法版本。

```EBNT
SyntaxDecl = "syntax" "=" SyntaxName .
SyntaxName = string_lit .
```

语法版本示例
```EBNT_DEMO
syntax = "v1"
```

## 导入声明(Import declarations)
导出声明用于 当前 API 文件导入其他 API 的时候使用。

```EBNT
ImportDecl       = "import" ( ImportPath | "(" { ImportPath ";" } ")" ) .
ImportPath       = string_lit .
```
## 信息声明(Info declaration)
信息声明用于声明 API 的一些额外信息。
```EBNT
InfoDecl    = "info" "(" { InfoElement ";"} ")" .
InfoElement = identifier ":" ( string_lit | identifier) .
```

```EBNT_DEMO
info (
    auth: dylan
    desc: `abc
def`
)
```

## 类型声明(Type declarations)
一个类型声明绑定了一个标识符（也就是 类型名 ）到一个 类型 。目前 API 只支持类型定义。

```EBNT
TypeDecl = "type" ( TypeDef | "(" { TypeDef ";" } ")" ) .
TypeDef = identifier Type .
```

# 类型(Types)
类型确定一组值。
```EBNT
Type      = TypeName | TypeLit | "(" Type ")" .
TypeName  = identifier . 
TypeLit   = ArrayType | StructType | PointerType | SliceType | MapType .
```

语言本身 预先声明 了一些特定的类型名。其它的命名类型则使用 类型声明 或者 类型形参列表 引入。 复合类型 ——数组、结构体、指针、分片、映射——可以由类型字面值构成。

## 布尔类型(Boolean types)
布尔类型 代表以预先声明的常量 true 和 false 所表示的布尔真值的集合。预先声明的布尔类型为 **bool** ，这是一个 定义类型 。

### 数字类型(Numeric types)
整数 、 浮点数 或 复数 类型分别代表整数、浮点数或复数值的集合。 它们被统称为 数字类型 。

```EBNT
uint8       无符号的  8 位整数集合（0 到 255）
uint16      无符号的 16 位整数集合（0 到 65535）
uint32      无符号的 32 位整数集合（0 到 4294967295）
uint64      无符号的 64 位整数集合（0 到 18446744073709551615）

int8        带符号的  8 位整数集合（-128 到 127）
int16       带符号的 16 位整数集合（-32768 到 32767）
int32       带符号的 32 位整数集合（-2147483648 到 2147483647）
int64       带符号的 64 位整数集合（-9223372036854775808 到 9223372036854775807）

float32     所有 IEEE-754 标准的 32 位浮点数数字集合
float64     所有 IEEE-754 标准的 64 位浮点数数字集合
```

还有一部分定一下是与预先声明架构相关
```EBNT
uint     可以是 32 或 64 位
int      和 uint 大小相同
```

## 结构体类型(Struct types)
结构是命名元素的序列，称为字段，每个字段有一个名称和一个类型。字段名可以显式指定(IdentifierList)或隐式指定(EmbeddedField)。在结构中，非空字段名必须是唯一的。
```EBNT
StructType    = "{" { FieldDecl ";" } "}" .
FieldDecl     = (identifier Type | EmbeddedField) [ Tag ] .
EmbeddedField = [ "*" ] TypeName .
Tag           = string_lit .
```

// TODO: EmbeddedField 定义是不够明确

```EBNT_DEMO
type Foo {
    Stude {
        Name string
    }
    Foo Foo 
}
```

## 指针类型(Pointer Type)
指针类型表示指向一给定类型的 变量 的所有指针的集合，这个给定类型称为该指针的 基础类型 。
```EBNT
PointerType = "*" Type .
```

## 数组类型(ArrayType)
数组是单一类型元素的有序序列，该单一类型称为元素类型。元素的个数被称为数组长度，并且不能为负值。

```EBNT
ArrayType   = "[" ArrayLength, "]" Type .
ArrayLength = Expression .
```

## 分片类型(SliceType)
分片是针对一个底层数组的连续段的描述符，它提供了对该数组内有序序列元素的访问。
分片类型表示其元素类型的数组的所有分片的集合。元素的数量被称为分片长度，且不能为负。未初始化的分片的值为 nil 。
```EBNT
SliceType = "[", "]", Type .
```

```EBNT_DEMO
type Foo {
    Arr []string 
}
```

## 映射类型(MapType)
映射是由一种类型的元素所组成的无序组，这个类型被称为元素类型， 其元素被一组另一种类型的唯一 键 索引，这个类型被称为键类型。

```EBNT
MapType = "map" "[" Type "]" Type .
```

# 服务定义
服务为 API 定义的路由服务，一个 API 可以有多个 Service， 但是必须 ServiceName 必须同名。
```EBNT
ServiceDecl = (ServiceExtDecl) ServiceBody .

ServiceBody = "service" value_string_lit "{" { RouteDecl } "}" .
RouteDecl = (Doc) ";" Handler ";" Method Path (Request) (Response) .
Doc = "@doc" string_lit .
Handler = "@handler" identifier .
Method = "get" | "post" | "put" | "head"  | "otions" | "delete" | "patch" .
Path   = value_string_lit .
Request = "(" identifier ")" .
Response = "returns" "(" identifier ")" .
```

```EBNT_DEMO
service user-api {
    @doc ""
    @handler GetUserInfo
    get /api/user/info(req) returns (resp)
}
```

TODO: Response 括号

## 服务额外扩展定义
服务 @server 扩展信息定义。
```EBNT
ServiceExtDecl = "@server" "(" { ServiceExtElement } ")" .
ServiceExtElement = identifier ":" (value_string_lit | string_lit | identifier) .
```

```EBNT_DEMO
@server (
    jwt: auth
    timeout: 3s
)
```

