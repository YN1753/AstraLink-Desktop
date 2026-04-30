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

func (r *RelationService) CreateRelation(req model.Relation) error {
	return r.base.CreateRelation(req)
}

func (r *RelationService) GetAllRelations() (*[]model.Relation, error) {
	return r.base.GetAllRelations()
}

func (r *RelationService) GetOutgoingLinks(noteID string) (*[]model.Relation, error) {
	return r.base.GetRelationsByFromID(noteID)
}

func (r *RelationService) GetBacklinks(noteID string) (*[]model.Relation, error) {
	return r.base.GetRelationsByToID(noteID)
}
