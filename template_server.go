package sitepages

import (
	"html/template"
	"log"
	"net/http"
)

type TemplateServerMux struct {
	Mux       *http.ServeMux
	templates map[string]*template.Template
}

func NewTemplateServerMux(frontPagesFolder string, templateComponentFolder string) *TemplateServerMux {
	return &TemplateServerMux{
		Mux:       http.NewServeMux(),
		templates: LoadAllTemplatePages(frontPagesFolder, templateComponentFolder),
	}
}

func (t *TemplateServerMux) Handle(pattern string, handler http.Handler) {
	t.Mux.Handle(pattern, handler)
}

func (t *TemplateServerMux) HandlePage(pattern string, page string, requireAuth bool) {
	template, exists := t.templates[page]
	if !exists {
		log.Fatal("Page Template doesn't exists", page)
	}
	t.Mux.Handle(pattern, TemplateHandler{template, requireAuth})
}

func (t *TemplateServerMux) HandlePageByLinkAndId(pattern string, page string, mapping LinkAndIdPageMap) {
	template, exists := t.templates[page]
	if !exists {
		log.Fatal("Page Template doesn't exists ", page)
	}
	t.Mux.Handle(pattern, PageByIdTemplateHandler{template, mapping})
}

func (t *TemplateServerMux) HandlePageByLink(pattern string, page string, mapping LinkPageMap) {
	template, exists := t.templates[page]
	if !exists {
		log.Fatal("Page Template doesn't exists ", page)
	}
	t.Mux.Handle(pattern, PageLinksTemplateHandler{template, mapping})
}

func (t *TemplateServerMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.Mux.ServeHTTP(w, r)
}
