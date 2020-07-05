package bundle

import (
	"database/sql"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/patrickascher/gofw/locale"
	"github.com/patrickascher/gofw/orm"
	"github.com/patrickascher/gofw/server"
	"github.com/patrickascher/gofw/sqlquery"
	"golang.org/x/text/language"
)

const DB = "dbBundle"

func init() {
	_ = locale.Register(DB, newDB)
}

func convertDBMessage(message Message) *i18n.Message {
	ia8nMessage := i18n.Message{ID: message.MessageID}
	if message.Description.Valid {
		ia8nMessage.Description = message.Description.String
	}
	if message.One.Valid {
		ia8nMessage.One = message.One.String
	}
	if message.Two.Valid {
		ia8nMessage.Two = message.Two.String
	}
	if message.Few.Valid {
		ia8nMessage.Few = message.Few.String
	}
	if message.Many.Valid {
		ia8nMessage.Many = message.Many.String
	}
	if message.Other.Valid {
		ia8nMessage.Other = message.Other.String
	}
	if message.Zero.Valid {
		ia8nMessage.Zero = message.Zero.String
	}
	return &ia8nMessage
}

// New satisfies the config.provider interface.
func newDB() locale.BundleI {
	return &dbBundle{}
}

type dbBundle struct {
	raw         map[string]Message
	defaultLang language.Tag
}

func (d *dbBundle) SetDefaultLanguage(defaultLang language.Tag) {
	d.defaultLang = defaultLang
}
func (d *dbBundle) DefaultLanguage() language.Tag {
	return d.defaultLang
}

func (d *dbBundle) DefaultMessage(id string) *i18n.Message {
	if msg, ok := d.raw[id]; ok {
		return convertDBMessage(msg)
	}
	return &i18n.Message{ID: id}
}

func (d *dbBundle) AddSource(source interface{}) error {

	messages := source.([]Message)
	builder, err := server.Builder(server.DEFAULT)
	if err != nil {
		return err
	}

	var dbMessages []Message
	dbMessage := Message{}
	err = dbMessage.Init(&dbMessage)
	if err != nil {
		return err
	}

	err = dbMessage.All(&dbMessages, sqlquery.NewCondition().Where("lang = ?", locale.RAW))
	if err != nil {
		return err
	}

	var existingIDs []int
	if len(messages) > 0 && d.raw == nil {
		d.raw = make(map[string]Message, len(messages))
	}
	for _, m := range messages {
		d.raw[m.MessageID] = m
		err = m.Init(&m)
		if err != nil {
			return err
		}

		m.Lang = locale.RAW

		foundMessage := Message{}
		for i, existing := range dbMessages {
			if existing.MessageID == m.MessageID {
				foundMessage = existing
				dbMessages = append(dbMessages[:i], dbMessages[i+1:]...)
				break
			}
		}

		if foundMessage.ID == 0 {
			err = m.Create()
			if err != nil {
				return err
			}
		} else {
			//checking for changes
			m.ID = foundMessage.ID
			if m.Zero != foundMessage.Zero ||
				m.Few != foundMessage.Few ||
				m.Many != foundMessage.Many ||
				m.Other != foundMessage.Other ||
				m.One != foundMessage.One ||
				m.Two != foundMessage.Two ||
				m.Plural != foundMessage.Plural ||
				m.Description != foundMessage.Description {
				err = m.Update()
				if err != nil {
					return err
				}
			}
		}
	}

	// delete non existing keys
	if len(dbMessages) > 0 {
		for _, existing := range dbMessages {
			existingIDs = append(existingIDs, existing.ID)
		}
		_, err = builder.Delete("translations").Where("lang = ? AND id IN (?)", locale.RAW, existingIDs).Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *dbBundle) Bundle() (*i18n.Bundle, error) {

	bundle := i18n.NewBundle(d.defaultLang)

	// add database messages
	var messages []Message
	model := &Message{}
	err := model.Init(model)
	if err != nil {
		return nil, err
	}

	err = model.All(&messages, sqlquery.NewCondition().Where("lang != ?", locale.RAW))
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	for _, m := range messages {
		message := i18n.Message{ID: m.MessageID}
		if m.Description.Valid {
			message.Description = m.Description.String
		}
		if m.Zero.Valid {
			message.Zero = m.Zero.String
		}
		if m.One.Valid {
			message.One = m.One.String
		}
		if m.Two.Valid {
			message.Two = m.Two.String
		}
		if m.Few.Valid {
			message.Few = m.Few.String
		}
		if m.Many.Valid {
			message.Many = m.Many.String
		}
		if m.Other.Valid {
			message.Other = m.Other.String
		}

		lang, err := language.Parse(m.Lang)
		if err != nil {
			return nil, err
		}

		err = bundle.AddMessages(lang, &message)
		if err != nil {
			return nil, err
		}
	}

	return bundle, nil
}

type Message struct {
	orm.Model
	ID int

	MessageID   string
	Lang        string
	Description orm.NullString

	Zero  orm.NullString
	One   orm.NullString
	Two   orm.NullString
	Few   orm.NullString
	Many  orm.NullString
	Other orm.NullString

	Plural bool
}

func (m *Message) DefaultTableName() string {
	return "translations"
}
