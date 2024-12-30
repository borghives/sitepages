package sitepages

import (
	"encoding/xml"
	"errors"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/borghives/websession"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LinkAndIdPageMap map[string]map[string]*SitePage
type LinkPageMap map[string]*SitePage

type TemplateHandler struct {
	WebTemplates *template.Template
	RequireAuth  bool
}

// ServeHTTP implements the http.Handler interface
func (h TemplateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.WebTemplates == nil {
		log.Printf("instance@%s ERROR page template is nil", websession.GetHostInfo().Id)
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	webSession := websession.RefreshRequestSession(w, r)
	if h.RequireAuth && (webSession.UserName == "" || webSession.UserId.IsZero()) {
		http.Redirect(w, r, getAuthLoginUrl(r.URL.Path), http.StatusFound)
		return
	}

	tData := TemplateData{
		ID:           r.PathValue("id"),
		SessionToken: webSession.GenerateSessionToken(),
		Nonce:        websession.GetRandomHexString(),
		RootId:       r.PathValue("rid"),
		RelType:      CastRelationType(r.PathValue("relationtype")),
		Username:     webSession.UserName,
	}
	executeTemplateToHttpResponse(w, h.WebTemplates, tData)
}

type PageListTemplateHandler struct {
	WebTemplates  *template.Template
	SelectedPages *PageList
}

// ServeHTTP implements the http.Handler interface
func (h PageListTemplateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.WebTemplates == nil {
		log.Printf("instance@%s ERROR page template is nil", websession.GetHostInfo().Id)
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	pagelistmarshal, err := xml.MarshalIndent(h.SelectedPages, "", "  ")
	if err != nil {
		log.Printf("instance@%s ERROR marshalling page to xml", websession.GetHostInfo().Id)
	}

	datamarshal, err := xml.MarshalIndent(h.SelectedPages.PageData, "", "  ")
	if err != nil {
		log.Printf("instance@%s ERROR marshalling page content data to xml", websession.GetHostInfo().Id)
	}

	webSession := websession.RefreshRequestSession(w, r)

	tData := TemplateData{
		ID:           h.SelectedPages.ID.Hex(),
		Title:        "",
		Username:     webSession.UserName,
		SessionToken: webSession.GenerateSessionToken(),
		Nonce:        websession.GetRandomHexString(),
		Models: []template.HTML{
			template.HTML(pagelistmarshal),
			template.HTML(datamarshal),
		},
	}
	executeTemplateToHttpResponse(w, h.WebTemplates, tData)
}

type PageLinksTemplateHandler struct {
	WebTemplates *template.Template
	Page         LinkPageMap
}

// ServeHTTP implements the http.Handler interface
func (h PageLinksTemplateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	title := ""
	id := ""
	pageRoot := ""

	pathKey := r.URL.Path[1:]
	page, exists := h.Page[pathKey]
	if !exists {
		log.Printf("host instance@%s ERROR getting path key from request", websession.GetHostInfo().Id.Hex())
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	title = page.Title
	id = page.ID.Hex()
	pageRoot = page.Root.Hex()
	pagemarshal, err := xml.MarshalIndent(page, "", "  ")
	if err != nil {
		log.Printf("host instance@%s ERROR marshalling page to xml", websession.GetHostInfo().Id.Hex())
	}

	pagedatamarshal, err := xml.MarshalIndent(page.StanzaData, "", "  ")
	if err != nil {
		log.Printf("host instance@%s ERROR marshalling page content data to xml", websession.GetHostInfo().Id.Hex())
	}

	webSession := websession.RefreshRequestSession(w, r)

	tData := TemplateData{
		Title:        title,
		ID:           id,
		RootId:       pageRoot,
		Username:     webSession.UserName,
		SessionToken: webSession.GenerateSessionToken(),
		Nonce:        websession.GetRandomHexString(),
		Models: []template.HTML{
			template.HTML(pagemarshal),
			template.HTML(pagedatamarshal),
		},
	}

	executeTemplateToHttpResponse(w, h.WebTemplates, tData)
}

type PageByIdTemplateHandler struct {
	WebTemplates    *template.Template
	PageByLinkAndId LinkAndIdPageMap
}

// ServeHTTP implements the http.Handler interface
func (h PageByIdTemplateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	link, pageid, err := getPageParamFromRequest(r)
	if err != nil {
		log.Printf("instance@%s ERROR getting pageid from request: %s", websession.GetHostInfo().Id, err)
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	title := ""
	id := pageid
	pageRoot := ""
	pagemarshal := []byte{}
	pagedatamarshal := []byte{}

	page, _ := h.GetPage(link, pageid)
	if page != nil {
		title = page.Title
		id = page.ID.Hex()
		pageRoot = page.Root.Hex()
		pagemarshal, err = xml.MarshalIndent(page, "", "  ")
		if err != nil {
			log.Printf("instance@%s ERROR marshalling page to xml", websession.GetHostInfo().Id)
		}

		pagedatamarshal, err = xml.MarshalIndent(page.StanzaData, "", "  ")
		if err != nil {
			log.Printf("instance@%s ERROR marshalling page content data to xml", websession.GetHostInfo().Id)
		}
	}

	webSession := websession.RefreshRequestSession(w, r)

	tData := TemplateData{
		Title:        title,
		ID:           id,
		RootId:       pageRoot,
		Username:     webSession.UserName,
		SessionToken: webSession.GenerateSessionToken(),
		Nonce:        websession.GetRandomHexString(),
		LinkName:     link,
		Models: []template.HTML{
			template.HTML(pagemarshal),
			template.HTML(pagedatamarshal),
		},
	}

	executeTemplateToHttpResponse(w, h.WebTemplates, tData)
}

func (h PageByIdTemplateHandler) GetPage(link string, pageid string) (*SitePage, bool) {
	pageMap, exists := h.PageByLinkAndId[link]
	if exists {
		page, exists := pageMap[pageid]
		if exists {
			return page, true
		}
	}

	// Page not found
	return nil, false
}

// TemplateData is the data passed to the template
type TemplateData struct {
	ID                   string
	RootId               string
	Title                string
	Username             string
	RelType              RelationType
	SessionToken         string
	Nonce                string
	CommentToken         string
	CommentRelationToken string
	PageRelationToken    string
	LinkName             string
	Models               []template.HTML
}

func (d TemplateData) MakeTemplateFunc() template.FuncMap {
	return template.FuncMap{
		"gettopic": func() string {
			return "hello"
		},
	}
}

func getPageParamFromRequest(r *http.Request) (string, string, error) {
	// Get the last path segment from the request URL

	pagelink := r.PathValue("link")
	if len(pagelink) > MAX_LINK_LENGTH {
		return "", "", errors.New("link size is greater than limit")
	}

	pageid := r.PathValue("id")
	if pageid == "" {
		return pagelink, "", nil
	}

	objId, err := primitive.ObjectIDFromHex(pageid)
	if err != nil {
		return "", "", err
	}

	return pagelink, objId.Hex(), nil
}

func executeTemplateToHttpResponse(w http.ResponseWriter, webTemplates *template.Template, tData TemplateData) {
	err := webTemplates.Funcs(tData.MakeTemplateFunc()).Execute(w, tData)
	if err != nil {
		log.Printf("instance@%s ERROR executing template:  %s", websession.GetHostInfo().Id, err)
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}
}

func getAuthLoginUrl(redirectPath string) string {
	domain := websession.GetDomain()
	var retval string

	if (domain == "localhost") || (domain == "127.0.0.1") {
		retval = "http://" + domain + ":8000"
	} else {
		retval = "https://login" + domain
	}

	return retval + "/?redirect=" + url.QueryEscape(redirectPath)
}
