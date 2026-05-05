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
		FromID: nodeID,
		ToID:   tagID,
		Type:   "tag",
	})
}
