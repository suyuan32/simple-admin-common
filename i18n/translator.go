// Copyright 2023 The Ryan SU Authors (https://github.com/suyuan32). All Rights Reserved.
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

package i18n

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/text/language"
	"google.golang.org/grpc/status"

	"github.com/suyuan32/simple-admin-common/utils/errcode"
	"github.com/suyuan32/simple-admin-common/utils/parse"
)

//go:embed locale/*.json
var LocaleFS embed.FS

// Translator is a struct storing translating data.
type Translator struct {
	bundle       *i18n.Bundle
	localizer    map[language.Tag]*i18n.Localizer
	supportLangs []language.Tag
}

// NewBundle returns a bundle from FS.
func (l *Translator) NewBundle(file embed.FS) {
	bundle := i18n.NewBundle(language.Chinese)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	_, err := bundle.LoadMessageFileFS(file, "locale/zh.json")
	logx.Must(err)
	_, err = bundle.LoadMessageFileFS(file, "locale/en.json")
	logx.Must(err)

	l.bundle = bundle
}

// AddBundleFromEmbeddedFS adds new bundle into translator from embedded file system
func (l *Translator) AddBundleFromEmbeddedFS(file embed.FS, path string) error {
	if _, err := l.bundle.LoadMessageFileFS(file, path); err != nil {
		return err
	}
	return nil
}

// NewBundleFromFile returns a bundle from a directory which contains i18n files.
func (l *Translator) NewBundleFromFile(conf Conf) {
	bundle := i18n.NewBundle(language.Chinese)
	filePath, err := filepath.Abs(conf.Dir)
	logx.Must(err)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	_, err = bundle.LoadMessageFile(filepath.Join(filePath, "locale/zh.json"))
	logx.Must(err)
	_, err = bundle.LoadMessageFile(filepath.Join(filePath, "locale/en.json"))
	logx.Must(err)

	l.bundle = bundle
}

// AddBundleFromFile adds new bundle into translator from file path.
func (l *Translator) AddBundleFromFile(path string) error {
	if _, err := l.bundle.LoadMessageFile(path); err != nil {
		return err
	}
	return nil
}

// NewTranslator sets localize for translator.
func (l *Translator) NewTranslator() {
	l.supportLangs = append(l.supportLangs, language.Chinese)
	l.supportLangs = append(l.supportLangs, language.English)
	l.localizer = make(map[language.Tag]*i18n.Localizer)
	l.localizer[language.Chinese] = i18n.NewLocalizer(l.bundle, language.Chinese.String())
	l.localizer[language.English] = i18n.NewLocalizer(l.bundle, language.English.String())
}

// AddLanguageSupport adds supports for new language
func (l *Translator) AddLanguageSupport(lang language.Tag) {
	l.supportLangs = append(l.supportLangs, lang)
	l.localizer[lang] = i18n.NewLocalizer(l.bundle, lang.String())
}

// AddLanguagesByConf adds multiple languages from file system by i18n Conf.
// If Conf.Dir is empty, it will load paths in embedded FS.
// If Conf.Dir is not empty, it will load paths joined with Dir path.
func (l *Translator) AddLanguagesByConf(conf Conf, fs embed.FS) {
	if len(conf.SupportLanguages) > 0 {
		if len(conf.SupportLanguages) != len(conf.BundleFilePaths) {
			logx.Must(errors.New("the i18n config of SupportLanguages is not the same as BundleFilePaths, please check the configuration"))
		} else {
			for i, v := range conf.SupportLanguages {
				l.AddLanguageSupport(parse.ParseTags(v)[0])
				if conf.Dir == "" {
					err := l.AddBundleFromEmbeddedFS(fs, conf.BundleFilePaths[i])
					if err != nil {
						logx.Must(fmt.Errorf("failed to load files from %s for i18n, please check the "+
							"configuration, error: %s", conf.BundleFilePaths[i], err.Error()))
					}
				} else {
					err := l.AddBundleFromFile(filepath.Join(conf.Dir, conf.BundleFilePaths[i]))
					if err != nil {
						logx.Must(fmt.Errorf("failed to load files from %s for i18n, please check the "+
							"configuration, error: %s", filepath.Join(conf.Dir, conf.BundleFilePaths[i]), err.Error()))
					}
				}
			}
		}
	} else {
		logx.Must(errors.New("the i18n config of SupportLanguages is empty, please check the configuration"))
	}
}

// Trans used to translate any i18n string.
func (l *Translator) Trans(ctx context.Context, msgId string) string {
	message, err := l.MatchLocalizer(ctx.Value("lang").(string)).LocalizeMessage(&i18n.Message{ID: msgId})
	if err != nil {
		return msgId
	}

	if message == "" {
		return msgId
	}

	return message
}

// TransError translates the error message
func (l *Translator) TransError(ctx context.Context, err error) error {
	lang := ctx.Value("lang").(string)
	if errcode.IsGrpcError(err) {
		message, e := l.MatchLocalizer(lang).LocalizeMessage(&i18n.Message{ID: strings.Split(err.Error(), "desc = ")[1]})
		if e != nil || message == "" {
			message = err.Error()
		}
		return status.Error(status.Code(err), message)
	} else if codeErr, ok := err.(*errorx.CodeError); ok {
		message, e := l.MatchLocalizer(lang).LocalizeMessage(&i18n.Message{ID: codeErr.Error()})
		if e != nil || message == "" {
			message = codeErr.Error()
		}
		return errorx.NewCodeError(codeErr.Code, message)
	} else if apiErr, ok := err.(*errorx.ApiError); ok {
		message, e := l.MatchLocalizer(lang).LocalizeMessage(&i18n.Message{ID: apiErr.Error()})
		if e != nil {
			message = apiErr.Error()
		}
		return errorx.NewApiError(apiErr.Code, message)
	} else {
		return errorx.NewApiError(http.StatusInternalServerError, err.Error())
	}
}

// MatchLocalizer used to matcher the localizer in map
func (l *Translator) MatchLocalizer(lang string) *i18n.Localizer {
	tags := parse.ParseTags(lang)
	for _, v := range tags {
		if val, ok := l.localizer[v]; ok {
			return val
		}
	}

	return l.localizer[language.Chinese]
}

// NewTranslator returns a translator by FS.
func NewTranslator(file embed.FS) *Translator {
	trans := &Translator{}
	trans.NewBundle(file)
	trans.NewTranslator()
	return trans
}

// NewTranslatorFromFile returns a translator by FS.
func NewTranslatorFromFile(conf Conf) *Translator {
	trans := &Translator{}
	trans.NewBundleFromFile(conf)
	trans.NewTranslator()
	return trans
}
