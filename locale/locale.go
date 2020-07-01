package locale

import (
	"errors"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var Bundle *i18n.Bundle

type localizer struct {
	localizer *i18n.Localizer
}

// CreateLocalizer create localizer object to generate text messages.
func NewLocalizer(lang string) (*localizer, error) {
	if Bundle == nil || lang == "" {
		return nil, errors.New("global bundle ord language is not defined")
	}

	l := i18n.NewLocalizer(Bundle, lang)
	v := &localizer{localizer: l}
	return v, nil
}

// Translate form and output a message based on messageID and template configuration.
func (v *localizer) Translate(messageID string, template ...map[string]interface{}) string {
	return v.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: getTemplateData(template...)})
}

// TranslatePlural form and output a message based on messageID, template and pluralCount configuration.
func (v *localizer) TranslatePlural(messageID string, pluralCount interface{}, template ...map[string]interface{}) string {
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
