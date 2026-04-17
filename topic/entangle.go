package topic

import (
	"log"
	"time"

	"github.com/borghives/entanglement/concept"
	"github.com/borghives/sitepages"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CorrelationMap map[string]string

type EntangleProperties struct {
	Token        string                    `xml:"-" json:"Token" bson:"-" `
	Correlations map[string]CorrelationMap `xml:"-" json:"Correlations,omitempty" bson:"-" `
}

func (e *EntangleProperties) SetCorrelationProperties(typeName string, properties CorrelationMap) {
	if e.Correlations == nil {
		e.Correlations = make(map[string]CorrelationMap)
	}

	e.Correlations[typeName] = properties
}

type EntangledResponse struct {
	BaseResponse
	EntanglementState *EntangleProperties `xml:"-" json:"EntangleProperties,omitempty" bson:"-" `
}

func (e *EntangledResponse) Append(data any) bson.ObjectID {
	switch data := data.(type) {
	case *concept.Entanglement:
		if e.EntanglementState == nil {
			e.EntanglementState = &EntangleProperties{}
		}
		e.EntanglementState.Token = data.GenerateToken()
		e.EntangleFrame(data)
		return bson.ObjectID{}
	default:
		return e.BaseResponse.Append(data)
	}
}

func (e *EntangledResponse) EntangleFrame(entanglement *concept.Entanglement) {
	switch entanglement.Frame {
	case "page_system":
		var page *sitepages.SitePage
		for _, data := range e.PageData {
			if data.ID.Hex() == e.TargetID.Hex() {
				page = &data
				break
			}
		}
		e.EntanglementState = EntanglePage(e.EntanglementState, entanglement, page)
	}
}

func EntanglePage(state *EntangleProperties, entanglement *concept.Entanglement, page *sitepages.SitePage) *EntangleProperties {
	if page == nil || state == nil {
		return state
	}

	entanglement.SetProperty("pageid", page.ID.Hex())
	entanglement.SetProperty("rootid", page.Root.Hex())

	pageId := entanglement.GenerateCorrelation(page.ID.Hex())
	state.SetCorrelationProperties("page", CorrelationMap{
		page.ID.Hex(): pageId,
	})

	stanzaEntanglement := entanglement.CreatSubFrame("stanza_system")
	stanzaEntanglement.SetProperty("baseid", pageId)

	if len(page.Contents) > 0 {
		stanzaCorrelation := CorrelationMap{}
		for _, content := range page.Contents {
			stanzaCorrelation[content.Hex()] = stanzaEntanglement.GenerateCorrelation(content.Hex())
		}

		state.Correlations["stanza"] = stanzaCorrelation
	}

	return state
}

func EntangleCommentProperties(entanglement *concept.Entanglement, sourceId bson.ObjectID, rootId bson.ObjectID, coolDown time.Duration) CorrelationMap {
	moment := sitepages.GenerateMomentString(coolDown)
	commentEntanglement := entanglement.CreatSubFrame("comment_system")
	commentEntanglement.SetProperty("sourceid", sourceId.Hex())
	commentEntanglement.SetProperty("rootid", rootId.Hex())
	commentEntanglement.SetProperty("moment", moment)
	log.Println("sourceid", sourceId.Hex(), "rootid", rootId.Hex(), "moment", moment)
	return CorrelationMap{
		"--page-comment-creator": commentEntanglement.GenerateCorrelation("--page-comment-creator"),
		"moment":                 moment,
	}
}
