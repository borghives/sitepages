package topic

import (
	"errors"
	"testing"

	"github.com/borghives/entanglement"
	"github.com/borghives/sitepages"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestStatusResponse(t *testing.T) {
	err := errors.New("test error")
	resp := NewStatusError(err, 404)

	if resp.ErrorCode() != 404 {
		t.Errorf("expected 404, got %d", resp.ErrorCode())
	}
	if !resp.(*StatusResponse).HasError() {
		t.Errorf("expected HasError to be true")
	}

	if resp.Error() != "Response Status 404: test error" {
		t.Errorf("expected error string to match, got %s", resp.Error())
	}

	status := resp.(*StatusResponse).GetStatus()
	if status.StatusCode != 404 {
		t.Errorf("expected status code 404, got %d", status.StatusCode)
	}

	// ErrorCode fallback
	emptyResp := StatusResponse{}
	if emptyResp.ErrorCode() != 500 {
		t.Errorf("expected fallback error code 500, got %d", emptyResp.ErrorCode())
	}
}

func TestBaseResponse_GetTarget(t *testing.T) {
	base := BaseResponse{}
	targetID := bson.NewObjectID()

	if base.GetTarget() != nil {
		t.Errorf("expected nil target initially")
	}

	base.SetTargetID(targetID)
	if base.GetTargetID() != targetID {
		t.Errorf("expected %v, got %v", targetID, base.GetTargetID())
	}

	// Target not found
	if base.GetTarget() != nil {
		t.Errorf("expected nil target since it's not in the data arrays")
	}

	// Add to PageData
	page := Page{}
	page.ID = targetID
	base.Append(page)

	target := base.GetTarget()
	if target == nil || target.GetID() != targetID {
		t.Errorf("expected to find target in PageData")
	}
}

func TestBaseResponse_Append(t *testing.T) {
	base := BaseResponse{}
	id1 := bson.NewObjectID()
	id2 := bson.NewObjectID()
	id3 := bson.NewObjectID()
	id4 := bson.NewObjectID()

	// Append values
	page := Page{}
	page.ID = id1
	stanza := Stanza{}
	stanza.ID = id2
	comment := Comment{}
	comment.ID = id3
	bundle := sitepages.Bundle{}
	bundle.ID = id4

	ret1 := base.Append(page)
	ret2 := base.Append(stanza)
	ret3 := base.Append(comment)
	ret4 := base.Append(bundle)

	if ret1 != id1 || ret2 != id2 || ret3 != id3 || ret4 != id4 {
		t.Errorf("expected append to return correct IDs")
	}

	if len(base.PageData) != 1 || len(base.StanzaData) != 1 || len(base.CommentData) != 1 || len(base.BundleData) != 1 {
		t.Errorf("expected data arrays to have 1 element each")
	}

	// Append pointers
	base.Append(&page)
	base.Append(&stanza)
	base.Append(&comment)
	base.Append(&bundle)

	if len(base.PageData) != 2 || len(base.StanzaData) != 2 || len(base.CommentData) != 2 || len(base.BundleData) != 2 {
		t.Errorf("expected data arrays to have 2 elements each after pointer append")
	}

	// Test unknown type fallback
	retUnknown := base.Append("unknown string")
	if !retUnknown.IsZero() {
		t.Errorf("expected zero ID for unknown type")
	}
}

func TestEntangledResponse(t *testing.T) {
	resp := NewResponse().(*EntangledResponse)

	if resp.EntanglementState != nil {
		t.Errorf("expected nil entanglement state initially")
	}

	sess := entanglement.Session{}
	resp.Append(sess)
	if resp.EntanglementState == nil {
		t.Errorf("expected entanglement state to be initialized")
	}

	sessPtr := &entanglement.Session{}
	resp.Append(sessPtr)

	// Test appending normal data delegates to BaseResponse
	id := bson.NewObjectID()
	page := Page{}
	page.ID = id
	ret := resp.Append(page)
	if ret != id {
		t.Errorf("expected entangled response to delegate append to base response correctly")
	}
}

func TestListTopicResponse(t *testing.T) {
	resp := NewListTopicResponse("test_list").(*ListTopicResponse)

	if len(resp.ListData) != 1 || resp.ListData[0].ID != "test_list" {
		t.Errorf("expected list data to be initialized with name test_list")
	}

	id := bson.NewObjectID()
	page := Page{}
	page.ID = id
	ret := resp.Append(page)
	if ret != id {
		t.Errorf("expected append to return correct ID")
	}

	if len(resp.ListData[0].Contents) != 1 || resp.ListData[0].Contents[0] != id {
		t.Errorf("expected appended ID to be added to list contents")
	}
}

func TestRelationTopicResponse(t *testing.T) {
	resp := NewRelationTopicResponse().(*RelationTopicResponse)

	linkDesc := LinkDescription{
		ObjectId: bson.NewObjectID(),
	}
	pageLink := UserToPageLink{
		LinkDescription: linkDesc,
	}

	linkDesc2 := LinkDescription{
		ObjectId: bson.NewObjectID(),
	}
	commentLink := UserToCommentLink{
		LinkDescription: linkDesc2,
	}

	ret1 := resp.Append(pageLink)
	ret2 := resp.Append(&pageLink)
	ret3 := resp.Append(commentLink)
	ret4 := resp.Append(&commentLink)

	if ret1 != linkDesc.ObjectId || ret2 != linkDesc.ObjectId || ret3 != linkDesc2.ObjectId || ret4 != linkDesc2.ObjectId {
		t.Errorf("expected append to return correct ObjectIds")
	}

	if len(resp.LinkDescs) != 4 {
		t.Errorf("expected link descriptions to be appended")
	}

	// Test fallback to BaseResponse
	id := bson.NewObjectID()
	page := Page{}
	page.ID = id
	ret := resp.Append(page)
	if ret != id || len(resp.PageData) != 1 {
		t.Errorf("expected fallback to BaseResponse for unknown types")
	}
}
