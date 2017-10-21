package payment

var (
	//提现参数序列化错误
	WithdrawParamsSerializeFail = &WithdrawResult{
		Status:   FAIL,
		FailCode: "PARAMS_SERIALIZE_FAIL",
		FailMsg:  "参数序列化错误",
	}
	//提现请求失败
	WithdrawRequestFail = &WithdrawResult{
		Status:   FAIL,
		FailCode: "REQUEST_FAIL",
		FailMsg:  "请求失败",
	}
	//提现请求结果读取异常，状态属于处理中，需要通过检测接口验证是否请求成功
	WithdrawResponseReadFail = &WithdrawResult{
		Status:   DEALING,
		FailCode: "RESPONSE_READ_FAIL",
		FailMsg:  "请求结果读取异常",
	}
	//请求结果解析异常，状态属于处理中,需要通过检测接口验证请求是否成功
	WithdrawResponseUnserializeFail = &WithdrawResult{
		Status:   DEALING,
		FailCode: "RESPONSE_UNSERIALIZE_FAIL",
		FailMsg:  "请求结果解析异常",
	}
	//请求结果签名验证失败，状态属于处理中,需要通过检测接口验证请求是否成功
	WithdrawResponseVerifyFail = &WithdrawResult{
		Status:   DEALING,
		FailCode: "RESPONSE_UNSERIALIZE_FAIL",
		FailMsg:  "请求结果签名验证失败",
	}
)
