package code

var (
	OK                  = add(200, "成功")
	BadRequest          = add(400, "错误请求")
	Unauthorized        = add(401, "未授权")
	Forbidden           = add(403, "禁止访问")
	NotFound            = add(404, "未找到")
	InternalServerError = add(500, "内部服务器错误")
	ServiceUnavailable  = add(503, "服务不可用")
	InvalidParameter    = add(422, "参数无效")
	DuplicateOperation  = add(409, "重复操作")
	StatusNotAvailable  = add(1001, "状态不可用")
	StatusException     = add(1002, "状态异常")
	MyCustomErrorCode   = add(1003, "自定义错误码")
	Expired             = add(1004, "已过期")
)
