package captcha

// Conf is the captcha configuration structure
type Conf struct {
	KeyLong   int    `json:",optional,default=5,env=CAPTCHA_KEY_LONG"`                               // captcha length
	ImgWidth  int    `json:",optional,default=240,env=CAPTCHA_IMG_WIDTH"`                            // captcha width
	ImgHeight int    `json:",optional,default=80,env=CAPTCHA_IMG_HEIGHT"`                            // captcha height
	Driver    string `json:",optional,default=digit,options=[digit,string,math],env=CAPTCHA_DRIVER"` // captcha type
}
