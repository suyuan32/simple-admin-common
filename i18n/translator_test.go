package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTranslator(t *testing.T) {
	l := &Translator{}
	l.NewBundle(LocaleFS)
	l.NewTranslator()
	res := l.Trans("zh", "common.success")
	assert.Equal(t, "成功", res)
}
