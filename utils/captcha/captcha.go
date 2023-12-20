// Copyright 2023 The Ryan SU Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package captcha

import (
	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
	redis2 "github.com/zeromicro/go-zero/core/stores/redis"
	"image/color"
)

const Prefix = "CAPTCHA:"

// MustNewRedisCaptcha returns the captcha using redis, it will exit when error occur
func MustNewRedisCaptcha(c Conf, r *redis2.Redis) *base64Captcha.Captcha {
	driver := NewDriver(c)

	store := NewRedisStore(r)

	return base64Captcha.NewCaptcha(driver, store)
}

// MustNewOriginalRedisCaptcha returns the captcha using original go redis, it will exit when error occur
func MustNewOriginalRedisCaptcha(c Conf, r *redis.Client) *base64Captcha.Captcha {
	driver := NewDriver(c)

	store := NewOriginalRedisStore(r)

	return base64Captcha.NewCaptcha(driver, store)
}

func NewDriver(c Conf) base64Captcha.Driver {
	var driver base64Captcha.Driver

	bgColor := &color.RGBA{
		R: 254,
		G: 254,
		B: 254,
		A: 254,
	}

	fonts := []string{
		"ApothecaryFont.ttf",
		"DENNEthree-dee.ttf",
		"Flim-Flam.ttf",
		"RitaSmith.ttf",
		"actionj.ttf",
		"chromohv.ttf",
	}

	switch c.Driver {
	case "digit":
		driver = base64Captcha.NewDriverDigit(c.ImgHeight, c.ImgWidth,
			c.KeyLong, 0.7, 80)
	case "string":
		driver = base64Captcha.NewDriverString(c.ImgHeight, c.ImgWidth, 12, 3, c.KeyLong,
			"qwertyupasdfghjkzxcvbnm23456789",
			bgColor, nil, fonts)
	case "math":
		driver = base64Captcha.NewDriverMath(c.ImgHeight, c.ImgWidth, 12, 3, bgColor,
			nil, fonts)
	case "chinese":
		driver = base64Captcha.NewDriverChinese(c.ImgHeight, c.ImgWidth, 10, 4, c.KeyLong,
			"天地玄黄宇宙洪荒日月盈辰宿列张寒来暑往秋收冬藏闰余成岁律吕调阳夫大水浸灌草木破坡石右云虫师军舰流浪数据速度", bgColor, nil,
			[]string{"wqy-microhei.ttc"})
	default:
		driver = base64Captcha.NewDriverDigit(c.ImgHeight, c.ImgWidth,
			c.KeyLong, 0.7, 80)
	}

	return driver
}
