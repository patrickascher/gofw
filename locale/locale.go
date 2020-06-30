package locale

import (
	"errors"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"os"
	"strings"
	"sync"
)

var GlobalBundle *i18n.Bundle

type Localizer struct {
	sync.Mutex
	localizer *i18n.Localizer
}

func NewBundle() {
	// default language
	bundle := i18n.NewBundle(language.English)
	// TODO add db messages
	GlobalBundle = bundle
}

// CreateLocalizer create localizer object to generate text messages.
func NewLocalizer(lang string) (*Localizer, error) {
	if GlobalBundle == nil {
		return nil, errors.New("bundles is not defined")
	}

	// fallback lang?
	if lang == "" {
		lang = os.Getenv("LANG")
		// remove ".UTF-8" suffix from language if found, as "en-US.UTF-8"
		if i := strings.Index(lang, ".UTF-8"); i != -1 {
			lang = lang[:i]
		}
	}

	localizer := i18n.NewLocalizer(GlobalBundle, lang)
	v := &Localizer{localizer: localizer}

	v.Translate("aaa")
	v.TranslatePlural("aaa", 2)
	return v, nil
}

// Translate form and output a message based on messageID and template configuration.
func (v *Localizer) Translate(messageID string, template ...map[string]interface{}) string {
	return v.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: getTemplateData(template...)})
}

// TranslatePlural form and output a message based on messageID, template and pluralCount configuration.
func (v *Localizer) TranslatePlural(messageID string, pluralCount interface{}, template ...map[string]interface{}) string {
	return v.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: getTemplateData(template...),
		PluralCount:  pluralCount})
}

func getTemplateData(template ...map[string]interface{}) map[string]interface{} {
	if len(template) > 0 {
		return template[0]
	}
	return make(map[string]interface{}, 0)
}
