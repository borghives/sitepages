package sitepages

import (
	"encoding/xml"
	"errors"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LinkAndIdPageMap map[string]map[string]*SitePage
type LinkPageMap map[string]*SitePage

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

func (t *TemplateServerMux) Handle(pattern string, page string, requireAuth bool) {
	template, exists := t.templates[page]
	if !exists {
		log.Fatal("Page Template doesn't exists", page)
	}
	t.Mux.Handle(pattern, TemplateHandler{template, requireAuth})
}

func (t *TemplateServerMux) HandlePageByLinkAndId(pattern string, page string, mapping LinkAndIdPageMap) {
	template, exists := t.templates[page]
	if !exists {
		log.Fatal("Page Template doesn't exists", page)
	}
	t.Mux.Handle(pattern, PageByIdTemplateHandler{template, mapping})
}

func (t *TemplateServerMux) HandlePageByLink(pattern string, page string, mapping LinkPageMap) {
	template, exists := t.templates[page]
	if !exists {
		log.Fatal("Page Template doesn't exists", page)
	}
	t.Mux.Handle(pattern, PageLinksTemplateHandler{template, mapping})
}

func (t *TemplateServerMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.Mux.ServeHTTP(w, r)
}

type TemplateHandler struct {
	WebTemplates *template.Template
	RequireAuth  bool
}

// ServeHTTP implements the http.Handler interface
func (h TemplateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.WebTemplates == nil {
		log.Printf("instance@%s ERROR page template is nil", GetHostInfo().Id)
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	webSession := RefreshRequestSession(w, r)
	if h.RequireAuth && (webSession.UserName == "" || webSession.UserId.IsZero()) {
		http.Redirect(w, r, getAuthLoginUrl(r.URL.Path), http.StatusFound)
		return
	}

	tData := TemplateData{
		ID:           r.PathValue("id"),
		SessionToken: webSession.GenerateSessionToken(),
		SaltSeed:     GetRandomHexString(),
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
		log.Printf("instance@%s ERROR page template is nil", GetHostInfo().Id)
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	pagelistmarshal, err := xml.MarshalIndent(h.SelectedPages, "", "  ")
	if err != nil {
		log.Printf("instance@%s ERROR marshalling page to xml", GetHostInfo().Id)
	}

	datamarshal, err := xml.MarshalIndent(h.SelectedPages.PageData, "", "  ")
	if err != nil {
		log.Printf("instance@%s ERROR marshalling page content data to xml", GetHostInfo().Id)
	}

	webSession := RefreshRequestSession(w, r)

	tData := TemplateData{
		ID:           h.SelectedPages.ID.Hex(),
		Title:        "",
		Username:     webSession.UserName,
		SessionToken: webSession.GenerateSessionToken(),
		SaltSeed:     GetRandomHexString(),
		Models: []template.HTML{
			template.HTML(pagelistmarshal),
			template.HTML(datamarshal),
		},
	}
	executeTemplateToHttpResponse(w, h.WebTemplates, tData)
}

type PageLinksTemplateHandler struct {
	WebTemplates *template.Template
	Page         map[string]*SitePage
}

// ServeHTTP implements the http.Handler interface
func (h PageLinksTemplateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	title := ""
	id := ""
	pageRoot := ""

	pathKey := r.URL.Path[1:]
	page, exists := h.Page[pathKey]
	if !exists {
		log.Printf("host instance@%s ERROR getting path key from request", GetHostInfo().Id.Hex())
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	title = page.Title
	id = page.ID.Hex()
	pageRoot = page.Root.Hex()
	pagemarshal, err := xml.MarshalIndent(page, "", "  ")
	if err != nil {
		log.Printf("host instance@%s ERROR marshalling page to xml", GetHostInfo().Id.Hex())
	}

	pagedatamarshal, err := xml.MarshalIndent(page.StanzaData, "", "  ")
	if err != nil {
		log.Printf("host instance@%s ERROR marshalling page content data to xml", GetHostInfo().Id.Hex())
	}

	webSession := RefreshRequestSession(w, r)

	tData := TemplateData{
		Title:        title,
		ID:           id,
		RootId:       pageRoot,
		Username:     webSession.UserName,
		SessionToken: webSession.GenerateSessionToken(),
		SaltSeed:     GetRandomHexString(),
		Models: []template.HTML{
			template.HTML(pagemarshal),
			template.HTML(pagedatamarshal),
		},
	}

	SetupToken(webSession, &tData, page.Root)
	executeTemplateToHttpResponse(w, h.WebTemplates, tData)
}

func SetupToken(webSession *WebSession, tData *TemplateData, rootId primitive.ObjectID) {
	if webSession == nil || tData == nil {
		return
	}

	var comment Comment

	comment.Moment = GenerateMomentString(0)
	comment.Root = rootId
	commentToken := GenerateCommentToken(*webSession, rootId.Hex(), comment.Moment)
	comment.ID, _ = primitive.ObjectIDFromHex(commentToken)

	commentmarshal, err := xml.MarshalIndent(comment, "", "  ")
	if err != nil {
		log.Printf("host instance@%s ERROR marshalling comment content data to xml", GetHostInfo().Id.Hex())

	}

	if tData.Models == nil {
		tData.Models = []template.HTML{}
	}

	tData.Models = append(tData.Models, template.HTML(commentmarshal))
	tData.CommentToken = commentToken
	tData.CommentRelationToken = GenerateRelationToken(*webSession, rootId.Hex(), "comment")
	tData.PageRelationToken = GenerateRelationToken(*webSession, rootId.Hex(), "page")
}

type PageByIdTemplateHandler struct {
	WebTemplates    *template.Template
	PageByLinkAndId map[string]map[string]*SitePage
}

// ServeHTTP implements the http.Handler interface
func (h PageByIdTemplateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	link, pageid, err := getPageParamFromRequest(r)
	if err != nil {
		log.Printf("instance@%s ERROR getting pageid from request: %s", GetHostInfo().Id, err)
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
			log.Printf("instance@%s ERROR marshalling page to xml", GetHostInfo().Id)
		}

		pagedatamarshal, err = xml.MarshalIndent(page.StanzaData, "", "  ")
		if err != nil {
			log.Printf("instance@%s ERROR marshalling page content data to xml", GetHostInfo().Id)
		}
	}

	webSession := RefreshRequestSession(w, r)

	tData := TemplateData{
		Title:        title,
		ID:           id,
		RootId:       pageRoot,
		Username:     webSession.UserName,
		SessionToken: webSession.GenerateSessionToken(),
		SaltSeed:     GetRandomHexString(),
		Models: []template.HTML{
			template.HTML(pagemarshal),
			template.HTML(pagedatamarshal),
		},
	}

	if page != nil {
		SetupToken(webSession, &tData, page.Root)
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
	SaltSeed             string
	CommentToken         string
	CommentRelationToken string
	PageRelationToken    string
	Models               []template.HTML
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
	err := webTemplates.Execute(w, tData)
	if err != nil {
		log.Printf("instance@%s ERROR executing template:  %s", GetHostInfo().Id, err)
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}
}

func getAuthLoginUrl(redirectPath string) string {
	domain := GetDomain()
	var retval string

	if (domain == "localhost") || (domain == "127.0.0.1") {
		retval = "http://" + domain + ":8000"
	} else {
		retval = "https://login" + domain
	}

	return retval + "/?redirect=" + url.QueryEscape(redirectPath)
}
