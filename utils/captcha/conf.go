package captcha

// Conf is the captcha configuration structure
type Conf struct {
	KeyLong   int    `json:",optional,default=5"`                                 // captcha length
	ImgWidth  int    `json:",optional,default=240"`                               // captcha width
	ImgHeight int    `json:",optional,default=80"`                                // captcha height
	Driver    string `json:",optional,default=digit,options=[digit,string,math]"` // captcha type
}
