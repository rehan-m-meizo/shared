package pipes

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// PipelineBuilder provides a fluent API for building MongoDB aggregation pipelines.
type PipelineBuilder struct {
	stages mongo.Pipeline
}

// NewPipeline initializes a new pipes builder.
func NewPipeline() *PipelineBuilder {
	return &PipelineBuilder{
		stages: mongo.Pipeline{},
	}
}

// Match adds a $match stage.
func (p *PipelineBuilder) Match(filter interface{}) *PipelineBuilder {
	p.stages = append(p.stages, bson.D{{Key: "$match", Value: filter}})
	return p
}

// Lookup adds a $lookup stage.
func (p *PipelineBuilder) Lookup(from, localField, foreignField, as string) *PipelineBuilder {
	stage := bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: from},
		{Key: "localField", Value: localField},
		{Key: "foreignField", Value: foreignField},
		{Key: "as", Value: as},
	}}}
	p.stages = append(p.stages, stage)
	return p
}

// Unwind adds a $unwind stage.
func (p *PipelineBuilder) Unwind(path string, preserveEmpty bool) *PipelineBuilder {
	stage := bson.D{{Key: "$unwind", Value: bson.D{
		{Key: "path", Value: path},
		{Key: "preserveNullAndEmptyArrays", Value: preserveEmpty},
	}}}
	p.stages = append(p.stages, stage)
	return p
}

// AddFields adds a $addFields stage.
func (p *PipelineBuilder) AddFields(fields interface{}) *PipelineBuilder {
	p.stages = append(p.stages, bson.D{{Key: "$addFields", Value: fields}})
	return p
}

// Project adds a $project stage.
func (p *PipelineBuilder) Project(fields interface{}) *PipelineBuilder {
	p.stages = append(p.stages, bson.D{{Key: "$project", Value: fields}})
	return p
}

// Sort adds a $sort stage.
func (p *PipelineBuilder) Sort(sort interface{}) *PipelineBuilder {
	p.stages = append(p.stages, bson.D{{Key: "$sort", Value: sort}})
	return p
}

// Limit adds a $limit stage.
func (p *PipelineBuilder) Limit(limit int64) *PipelineBuilder {
	p.stages = append(p.stages, bson.D{{Key: "$limit", Value: limit}})
	return p
}

// Skip adds a $skip stage.
func (p *PipelineBuilder) Skip(skip int64) *PipelineBuilder {
	p.stages = append(p.stages, bson.D{{Key: "$skip", Value: skip}})
	return p
}

// Build returns the final pipes.
func (p *PipelineBuilder) Build() mongo.Pipeline {
	return p.stages
}
