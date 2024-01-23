package code

var (
	OK                  = add(200, "成功")
	BadRequest          = add(400, "错误请求")
	Unauthorized        = add(401, "未授权")
	Forbidden           = add(403, "禁止访问")
	NotFound            = add(404, "未找到")
	InternalServerError = add(500, "内部服务器错误")
	ServiceUnavailable  = add(503, "服务不可用")
	StatusNotAvailable  = add(1001, "状态不可用")
	StatusException     = add(1002, "状态异常")
	InvalidParameter    = add(422, "参数无效")
)
