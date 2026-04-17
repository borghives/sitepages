package topic

import (
	"net/http"
)

type Handler interface {
	ServeTopic(response Response, request *http.Request)
}

type ResponseFactory func() Response

type InitializeResponse func(r *http.Request) Response

type ServePipe struct {
	Chains         []Handler
	CreateResponse ResponseFactory
	Initialize     InitializeResponse
	BodyLimit      int64
}

func NewServePipe(createResponse ResponseFactory) *ServePipe {
	return &ServePipe{
		CreateResponse: createResponse,
	}
}

func NewServePipeInitializer(initialize InitializeResponse) *ServePipe {
	return &ServePipe{
		Initialize: initialize,
	}
}

func (h *ServePipe) SetBodyLimit(limit int64) *ServePipe {
	h.BodyLimit = limit
	return h
}

func (h *ServePipe) Chain(chains ...Handler) *ServePipe {
	h.Chains = append(h.Chains, chains...)
	return h
}

func (h ServePipe) ServeTopic(response Response, r *http.Request) {
	for _, chain := range h.Chains {
		chain.ServeTopic(response, r)
		if response.HasError() {
			break
		}
	}
}

func (h ServePipe) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var topicResponse Response
	if h.Initialize != nil {
		topicResponse = h.Initialize(r)
	} else {
		topicResponse = h.CreateResponse()
	}

	if r.Body != nil {
		r.Body = http.MaxBytesReader(w, r.Body, h.BodyLimit)
	}

	h.ServeTopic(topicResponse, r)
	MarshalResponse(topicResponse, w)
}

func Handle(chain ...Handler) *ServePipe {
	return NewServePipe(NewResponse).SetBodyLimit(1048576).Chain(chain...)
}

func HandleRelation(chain ...Handler) *ServePipe {
	return NewServePipe(NewRelationTopicResponse).SetBodyLimit(1048576).Chain(chain...)
}

func HandleList(name string, servers ...Handler) *ServePipe {
	return NewServePipe(func() Response {
		listName := name
		var list *List
		if listName != "" {
			list = &List{ID: listName}
		}

		var listData []List
		if list != nil {
			listData = []List{*list}
		}

		return &ListTopicResponse{
			ListData: listData,
		}
	}).SetBodyLimit(1048576).Chain(servers...)
}
