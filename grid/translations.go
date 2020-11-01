package grid

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/patrickascher/gofw/slices"
	"strings"
)

const TPrefix = "GRID§§"

func Translations(ids ...string) []i18n.Message {
	var messages []i18n.Message

	messages = append(messages, i18n.Message{ID: TPrefix + "Action", Description: "Grid action column (create,edit)", Other: "Action"})
	messages = append(messages, i18n.Message{ID: TPrefix + "RowsPerPage", Description: "Grid pagination rows per page text", Other: "Rows per page"})
	messages = append(messages, i18n.Message{ID: TPrefix + "XofY", Description: "Grid the text in the pagination between X of Y", Other: "of"})
	messages = append(messages, i18n.Message{ID: TPrefix + "NoData", Description: "Grid no data available text", Other: "No data available"})
	messages = append(messages, i18n.Message{ID: TPrefix + "LoadingData", Description: "Grid loading data text", Other: "Loading data..."})

	messages = append(messages, i18n.Message{ID: TPrefix + "AddEdit", Description: "Grid filter add/edit text", Other: "Add / Edit"})
	messages = append(messages, i18n.Message{ID: TPrefix + "Required", Description: "Grid indicator for the * symbol (* required fields)", Other: "required field"})

	messages = append(messages, i18n.Message{ID: TPrefix + "QuickFilter", Description: "Filter text", Other: "Quick Filter"})
	messages = append(messages, i18n.Message{ID: TPrefix + "BtnFilter", Description: "Filter button", Other: "Filter"})
	messages = append(messages, i18n.Message{ID: TPrefix + "Filter", Description: "Filter text", Other: "Filter"})
	messages = append(messages, i18n.Message{ID: TPrefix + "Sort", Description: "Filter text", Other: "Sort"})
	messages = append(messages, i18n.Message{ID: TPrefix + "Desc", Description: "Sorting Desc", Other: "Desc"})

	messages = append(messages, i18n.Message{ID: TPrefix + "Fields", Description: "Filter text", Other: "Fields"})
	messages = append(messages, i18n.Message{ID: TPrefix + "Name", Description: "Filter text", Other: "Name"})
	messages = append(messages, i18n.Message{ID: TPrefix + "GroupBy", Description: "Filter text", Other: "Group by"})
	messages = append(messages, i18n.Message{ID: TPrefix + "EditFilter", Description: "Filter text", Other: "Edit filter"})
	messages = append(messages, i18n.Message{ID: TPrefix + "Show", Description: "Filter text", Other: "Show"})
	messages = append(messages, i18n.Message{ID: TPrefix + "Hide", Description: "Filter text", Other: "Hide"})

	messages = append(messages, i18n.Message{ID: TPrefix + "Save", Description: "Save text - used on buttons", Other: "Save"})
	messages = append(messages, i18n.Message{ID: TPrefix + "Delete", Description: "Save text - used on buttons", Other: "Delete"})
	messages = append(messages, i18n.Message{ID: TPrefix + "Back", Description: "Save text - used on buttons", Other: "Back"})

	messages = append(messages, i18n.Message{ID: TPrefix + "Close", Description: "Close text - used on buttons", Other: "Close"})
	messages = append(messages, i18n.Message{ID: TPrefix + "Add", Description: "Add text - used on buttons", Other: "Add"})
	messages = append(messages, i18n.Message{ID: TPrefix + "Export", Description: "Export text - used on buttons", Other: "Export"})

	messages = append(messages, i18n.Message{ID: TPrefix + "NoChanges", Description: "Export text - used on buttons", Other: "The form has no changes yet!"})
	messages = append(messages, i18n.Message{ID: TPrefix + "NotValid", Description: "Export text - used on buttons", Other: "The form is not valid!"})

	messages = append(messages, i18n.Message{ID: TPrefix + "Filter_=", Description: "Filter equal description", Other: "Field must be equal"})
	messages = append(messages, i18n.Message{ID: TPrefix + "Filter_!=", Description: "Filter not equal description", Other: "Field must be not equal"})
	messages = append(messages, i18n.Message{ID: TPrefix + "Filter_>=", Description: "Filter greater than description", Other: "Field must be greater than"})
	messages = append(messages, i18n.Message{ID: TPrefix + "Filter_<=", Description: "Filter less than description", Other: "Field must be less than"})
	messages = append(messages, i18n.Message{ID: TPrefix + "Filter_IN", Description: "Filter in description", Other: "Field must be in"})
	messages = append(messages, i18n.Message{ID: TPrefix + "Filter_NOTIN", Description: "Filter not in description", Other: "Field must be not in"})
	messages = append(messages, i18n.Message{ID: TPrefix + "Filter_LIKE", Description: "Filter like description", Other: "Field must be %like%"})
	messages = append(messages, i18n.Message{ID: TPrefix + "Filter_RLIKE", Description: "Filter right like description", Other: "Field must be like%"})
	messages = append(messages, i18n.Message{ID: TPrefix + "Filter_LLIKE", Description: "Filter left like description", Other: "Field must be %like"})
	messages = append(messages, i18n.Message{ID: TPrefix + "Filter_NULL", Description: "Filter null description", Other: "Field must be null"})
	messages = append(messages, i18n.Message{ID: TPrefix + "Filter_NOTNULL", Description: "Filter not null description", Other: "Field must be not null"})

	if len(ids) > 0 {
		var custom []i18n.Message
		for _, message := range messages {
			if _, exists := slices.Exists(ids, strings.Replace(message.ID, TPrefix, "", -1)); exists {
				custom = append(custom, message)
			}
		}
		return custom
	}

	return messages
}
