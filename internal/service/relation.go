package service

import (
	"astralink/internal/model"
	"astralink/internal/repo"
)

type RelationService struct {
	base *repo.BaseRepo
}

func NewRelationService(baseRepo *repo.BaseRepo) *RelationService {
	return &RelationService{
		base: baseRepo,
	}
}

func (r *RelationService) CreateLink(fromID, toID string) error {
	return r.base.UpsertRelation(model.Relation{
		FromID: fromID,
		ToID:   toID,
		Type:   "link",
	})
}

func (r *RelationService) LinkTagToNode(tagID string, nodeID string) error {
	return r.base.UpsertRelation(model.Relation{
		FromID: tagID,
		ToID:   nodeID,
		Type:   "tag",
	})
}

func (r *RelationService) GetTagsByNodeID(nodeID string) (*[]model.TagMessage, error) {
	relation, err := r.base.GetTagsByNodeID(nodeID)
	if err != nil {
		return nil, err
	}
	nodes := make([]string, 0)
	for _, value := range *relation {
		nodes = append(nodes, value.FromID)
	}
	nodeMessage, err := r.base.GetTagById(nodes)
	return &nodeMessage, err
}

func (r *RelationService) DeleteRelation(req model.DeleteRelationReq) error {
	return r.base.DeleteRelation(req)
}
