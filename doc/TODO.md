# 1. req resp 的定义
背景： 目前我们 req 与 resp 的定义是使用 go struct 的定义规则，但是存在的问题是 json key 是使用 tag 中的定义，
这样非 go 相关的同学可能不太熟悉这个规则，同时这个 key 编写起来太麻烦了，需要定义2边，一次 field，一次 tag 中。

```go
type User {
    Id int64 `json:"id"`
}

type User {
	id int64 
}
```

# 2. resp 相关定义
目前我们的 response 可以用过 httpx.Ok 进行返回，也有部分会改下 response 的方式，例如会返回 **{"code": 0, "data": {}}**
但是这种格式无法在 API 中提现出来。


# 3. api 中 error 相关的返回。
目前针对 API 的 error 无法体现出来。 
post /user/login(req) returns (resp)
返回的错误码，错误格式，错误描述等信息无法通过 API 体现

同时一种可能的格式
```api
post /foo(req) returns (resp, error)
```
 
# 4. type 是否需要支持 interface{} any
部分同学在使用的过程中，想要通过 interface{} 或者 any 表示任意对象

这种在 zero-api 中是不被支持的，我们推荐使用严格标准的表示方法，将具体的信息表示出来。


# 5. handler 同名支持

```api
@handler foo
get /foo(req)

@handler foo
post /foo(req)


// one
@handler foo
get|post /foo(req)


@handler foo
get /foo(req)
post /foo(req)

```
这种不会支持，建议直接定义不同的 handler 自行处理

# 6. type 定义 group
refer： https://github.com/zeromicro/go-zero/issues/1854
 