package model

const (
	DOMAIN_ID  = ""
	URL_PRE    = "https://api.mycactus.dev" //api 服务ip
	API_KEY    = ""                         //custody 会发送给客户 api key
	AK_ID      = ""                         //从 custody 获取的 api akid ，管理员在人员管理页面指定人员上传步骤一生成的公钥，对应人员会收到akid邮件
	KEY_NAME   = ""                         //密钥库里的私钥别名
	KEY_TYPE   = "pkcs12"                   //密钥的类型
	KEY_PASS   = ""                         //密钥的密码
	STORE_PASS = ""                         //密钥库的密码

	Bid               = ""                              //业务线编号
	SOLWallet         = ""                              //SOL钱包编号
	TronWallet        = ""                              //Tron钱包编号
	ETHWallet         = ""                              //ETH钱包编号
	SIGN_PIRVATE_PATH = ""                              //私钥库的路径
	TimeFormat        = "Mon, 02 Jan 2006 15:04:05 GMT" //时间格式
)
