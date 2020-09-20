package locale

import (
	"errors"
	"fmt"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var bundle BundleI
var bundleData *i18n.Bundle

// registry for all config providers.
var registry = make(map[string]provider)

const RAW = "raw"

// Error messages.
var (
	ErrNoProvider            = errors.New("config: empty config-name or config-provider is nil")
	ErrUnknownProvider       = errors.New("config: unknown config-provider %q")
	ErrProviderAlreadyExists = errors.New("config: config-provider %#v is already registered")
)

type LocalizerI interface {
	Translate(messageID string, template ...map[string]interface{}) (string, error)
	TranslatePlural(messageID string, pluralCount interface{}, template ...map[string]interface{}) (string, error)
}

type BundleI interface {
	Bundle() (*i18n.Bundle, error)
	AddSource(interface{}) error
	DefaultMessage(id string) *i18n.Message
	DefaultLanguage() language.Tag
	SetDefaultLanguage(language.Tag)
}

type localizer struct {
	localizer *i18n.Localizer
}

// provider is a function which returns the config interface.
// Like this the config provider is getting initialized only when its called.
type provider func() BundleI

// Register the config provider. This should be called in the init() of the providers.
// If the config provider/name is empty or is already registered, an error will return.
func Register(provider string, fn provider) error {
	if fn == nil || provider == "" {
		return ErrNoProvider
	}
	if _, exists := registry[provider]; exists {
		return fmt.Errorf(ErrProviderAlreadyExists.Error(), provider)
	}
	registry[provider] = fn
	return nil
}

func NewBundle(provider string, defaultLanguage string) (BundleI, error) {
	instanceFn, ok := registry[provider]
	if !ok {
		return nil, fmt.Errorf(ErrUnknownProvider.Error(), provider)
	}
	i := instanceFn()
	t, err := language.Parse(defaultLanguage)
	if err != nil {
		return nil, err
	}
	i.SetDefaultLanguage(t)
	return i, nil
}

func SetBundle(p BundleI) (err error) {
	bundle = p
	bundleData, err = p.Bundle()
	return
}

func Reload() (err error) {
	if bundle == nil || bundleData == nil {
		return errors.New("global bundle ord language is not defined")
	}
	bundleData, err = bundle.Bundle()
	return err
}

// CreateLocalizer create localizer object to generate text messages.
func NewLocalizer(lang ...string) (LocalizerI, error) {
	if bundle == nil || len(lang) == 0 || lang[0] == "" || bundleData == nil {
		return nil, errors.New("global bundle or language is not defined")
	}
	l := i18n.NewLocalizer(bundleData, lang...)
	v := &localizer{localizer: l}
	return v, nil
}

// Translate form and output a message based on messageID and template configuration.
func (v *localizer) Translate(messageID string, template ...map[string]interface{}) (string, error) {
	return v.localizer.Localize(&i18n.LocalizeConfig{
		MessageID:      messageID,
		DefaultMessage: bundle.DefaultMessage(messageID),
		TemplateData:   getTemplateData(template...)})
}

// TranslatePlural form and output a message based on messageID, template and pluralCount configuration.
func (v *localizer) TranslatePlural(messageID string, pluralCount interface{}, template ...map[string]interface{}) (string, error) {
	return v.localizer.Localize(&i18n.LocalizeConfig{
		MessageID:      messageID,
		DefaultMessage: bundle.DefaultMessage(messageID),
		TemplateData:   getTemplateData(template...),
		PluralCount:    pluralCount})
}

func getTemplateData(template ...map[string]interface{}) map[string]interface{} {
	if len(template) > 0 {
		return template[0]
	}
	return make(map[string]interface{}, 0)
}
