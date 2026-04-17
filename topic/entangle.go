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
	Entanglement *EntangleProperties `xml:"-" json:"EntangleProperties,omitempty" bson:"-" `
}

func (e *EntangledResponse) Append(data any) bson.ObjectID {
	switch data := data.(type) {
	case *concept.Entanglement:
		if e.Entanglement == nil {
			e.Entanglement = &EntangleProperties{}
		}
		e.Entanglement.Token = data.GenerateToken()
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
		if page != nil {
			entanglement.SetProperty("pageid", page.ID.Hex())
			entanglement.SetProperty("rootid", page.Root.Hex())

			pageId := entanglement.GenerateCorrelation(page.ID.Hex())
			e.Entanglement.SetCorrelationProperties("page", CorrelationMap{
				page.ID.Hex(): pageId,
			})

			stanzaEntanglement := entanglement.CreatSubFrame("stanza_system")
			stanzaEntanglement.SetProperty("baseid", pageId)

			if page.Contents != nil {
				stanzaCorrelation := CorrelationMap{}
				for _, content := range page.Contents {
					stanzaCorrelation[content.Hex()] = stanzaEntanglement.GenerateCorrelation(content.Hex())
				}

				e.Entanglement.Correlations["stanza"] = stanzaCorrelation
			}

			e.Entanglement.SetCorrelationProperties("comment", EntangleCommentProperties(entanglement, page.ID, page.Root, 0))
		}
	}
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
