package sitepages

import (
	"html/template"

	"github.com/borghives/entanglement"
	"github.com/borghives/entanglement/concept"
	"github.com/borghives/websession"
)

func CreateTemplateData(id string, rid string, session *websession.Session) TemplateData {
	return TemplateData{
		ID:           id,
		RootId:       rid,
		Entanglement: entanglement.CreateEntanglement(session),
		Username:     session.UserName,
	}
}

// TemplateData is the data passed to the template
type TemplateData struct {
	ID           string
	RootId       string
	Title        string
	Username     string
	LinkName     string
	Entanglement *concept.Entanglement
	Models       []template.HTML
}

func (d TemplateData) MakeTemplateFunc() template.FuncMap {
	return template.FuncMap{
		"gettopic": func() string {
			return "hello"
		},
	}
}
