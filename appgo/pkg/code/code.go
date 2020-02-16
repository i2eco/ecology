package code

const (
	MsgOk       = 0
	MsgErr      = 1
	MsgParamErr = 3

	// 10000 ~ 10199
	UserAccountLengthErr   = 10001
	UserAccountNicknameErr = 10002
	UserAccountErr3        = 10003
	UserAccountErr4        = 10004
	UserAccountErr5        = 10005
	UserAccountErr6        = 10006
	UserAccountErr7        = 10007

	UserUpdateErr1 = 10100
	UserUpdateErr2 = 10101
	UserUpdateErr3 = 10102
	UserUpdateErr4 = 10103
	UserUpdateErr5 = 10104
	UserUpdateErr6 = 10105
	UserUpdateErr7 = 10106

	BookCreateErr0  = 10200
	BookCreateErr1  = 10201
	BookCreateErr2  = 10202
	BookCreateErr3  = 10203
	BookCreateErr4  = 10204
	BookCreateErr5  = 10205
	BookCreateErr6  = 10206
	BookCreateErr7  = 10207
	BookCreateErr8  = 10208
	BookCreateErr9  = 10209
	BookCreateErr10 = 10210
	BookCreateErr11 = 10211
	BookCreateErr12 = 10212

	DocumentContentAuto   = 10301
	DocumentContentTrue   = 10302
	DocumentContentPost3  = 10303
	DocumentContentPost4  = 10304
	DocumentContentPost5  = 10305
	DocumentContentPost6  = 10306
	DocumentContentPost7  = 10307
	DocumentContentPost8  = 10308
	DocumentContentPost9  = 10309
	DocumentContentPost10 = 10310

	BookReleaseErr1 = 20001

	UploadCoverErr0  = 30000
	UploadCoverErr1  = 30001
	UploadCoverErr2  = 30002
	UploadCoverErr3  = 30003
	UploadCoverErr4  = 30004
	UploadCoverErr5  = 30005
	UploadCoverErr6  = 30006
	UploadCoverErr7  = 30007
	UploadCoverErr8  = 30008
	UploadCoverErr9  = 30009
	UploadCoverErr10 = 30010

	AccountBindErr1  = 40001
	AccountBindErr2  = 40002
	AccountBindErr3  = 40003
	AccountBindErr4  = 40004
	AccountBindErr5  = 40005
	AccountBindErr6  = 40006
	AccountBindErr7  = 40007
	AccountBindErr8  = 40008
	AccountBindErr9  = 40009
	AccountBindErr10 = 40010

	DodUploadErr1  = 50001
	DodUploadErr2  = 50002
	DodUploadErr3  = 50003
	DodUploadErr4  = 50004
	DodUploadErr5  = 50005
	DodUploadErr6  = 50006
	DodUploadErr7  = 50007
	DodUploadErr8  = 50008
	DodUploadErr9  = 50009
	DodUploadErr10 = 50010
	DodUploadErr11 = 50011
	DodUploadErr12 = 50011

	DeleteMemberErr1 = 60001
	DeleteMemberErr2 = 60002
	DeleteMemberErr3 = 60003
	DeleteMemberErr4 = 60004
	DeleteMemberErr5 = 60005
	DeleteMemberErr6 = 60006
	DeleteMemberErr7 = 60007

	DocUpdateErr1  = 70001
	DocUpdateErr2  = 70002
	DocUpdateErr3  = 70003
	DocUpdateErr4  = 70004
	DocUpdateErr5  = 70005
	DocUpdateErr6  = 70006
	DocUpdateErr7  = 70007
	DocUpdateErr8  = 70008
	DocUpdateErr9  = 70009
	DocUpdateErr10 = 70010
	DocUpdateErr11 = 70011
	DocUpdateErr12 = 70012

	DocContentPostErr1 = 70101
	DocContentPostErr2 = 70102
	DocContentPostErr3 = 70103
	DocContentPostErr4 = 70104
	DocContentPostErr5 = 70105
	DocContentPostErr6 = 70106
	DocContentPostErr7 = 70107
	DocContentPostErr8 = 70108
	DocContentPostErr9 = 70109
)

var CodeMap = map[int]string{
	0:           "成功",
	MsgParamErr: "参数错误",
	10001:       "登录微信失败",
	10112:       "不是待支付商品",
	10201:       "商品不存在",
	10202:       "系统错误",
	10203:       "分享后,可以下载",
	10204:       "积分购买后,可以下载",
	10205:       "现金购买后,可以下载",
	10206:       "积分或者现金购买后,可以下载",
	10207:       "类型错误",
	10208:       "系统错误",
	10209:       "已经有未支付订单,请及时支付",
	10251:       "系统错误",

	BookReleaseErr1: "书的标识符存在问题",
	UploadCoverErr1: "参数错误",
	UploadCoverErr2: "",

	AccountBindErr1: "参数错误",
	AccountBindErr2: "绑定用户失败，用户名或密码不正确",
	AccountBindErr3: "登录密码与确认密码不一致",
	AccountBindErr4: "用户名只能由英文字母数字组成，且在3-50个字符",
	AccountBindErr5: "密码必须在6-50个字符之间",
	AccountBindErr6: "邮箱格式不正确",
	AccountBindErr7: "用户昵称限制在2-20个字符",
	DodUploadErr1:   "参数错误",
	DodUploadErr2:   "没有发现需要上传的文件",
	DodUploadErr3:   "无法解析文件的格式",

	DeleteMemberErr1: "参数错误",
}
