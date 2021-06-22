package repo

import (
	"context"
	"fmt"

	"github.com/imdario/mergo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AuthzRepo interface {
	UpsertPolicy(ctx context.Context, sub, resourceType, resourceID, action string) error
	ListPolicy(ctx context.Context, sub, resourceType string) []string
	RemovePolicy(ctx context.Context, sub, resourceType, resourceID, action string) error
	RemovePolicyBySub(ctx context.Context, sub string) error
}

type basicAuthzRepo struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewAuthzRepo(client *mongo.Client) AuthzRepo {
	if client == nil {
		return nil
	}
	return &basicAuthzRepo{client: client, db: client.Database("authzdb")}
}

type policyDoc struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Sub      string             `json:"sub" bson:"sub"`
	Policies resourceTypes      `json:"policies" bson:"policies"`
}

type actions map[string][]string
type resourceTypes map[string]actions

func newPolicy(sub, resourceType, resourceID, action string) *policyDoc {
	return &policyDoc{
		Sub: sub,
		Policies: resourceTypes{
			resourceType: actions{
				action: []string{resourceID},
			},
		},
	}
}

func (b *basicAuthzRepo) UpsertPolicy(ctx context.Context, sub, resourceType, resourceID, action string) error {
	policiesCollection := b.db.Collection("policies")

	// policy to be created/merged
	newPolicy := newPolicy(sub, resourceType, resourceID, action)

	doc := policiesCollection.FindOne(ctx, bson.M{"sub": sub})

	// if no document found for the given sub, create one
	if err := doc.Err(); err != nil {
		if doc.Err() == mongo.ErrNoDocuments {
			if _, err := policiesCollection.InsertOne(ctx, newPolicy); err != nil {
				return err
			}
			return nil
		}
		return err
	}

	// else merge the new policy with existing one
	var existingPolicy *policyDoc
	if err := doc.Decode(&existingPolicy); err != nil {
		return err
	}

	// remove duplicates
	for removePolicy(existingPolicy, resourceType, resourceID, action) {
	}

	if err := mergo.MapWithOverwrite(newPolicy, existingPolicy, mergo.WithAppendSlice); err != nil {
		return err
	}
	_, err := policiesCollection.ReplaceOne(ctx, bson.M{"_id": existingPolicy.ID}, newPolicy)
	return err
}

func sliceIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

// removePolicy would remove newPolicy from existing policy
func removePolicy(existingPolicy *policyDoc, resourceTypes, resourceID, action string) bool {
	a, ok := existingPolicy.Policies[resourceTypes]
	if !ok { // means resource type/policy does not exist
		return ok // resourcetype/policy does not exist error can also be returned
	}

	resourceIDs, ok := a[action]
	if !ok { // means action/policy does not exist
		return ok
	}

	idx := sliceIndex(len(resourceIDs), func(i int) bool { return resourceIDs[i] == resourceID })
	if idx < 0 {
		return false
	}

	resourceIDs[idx], resourceIDs[len(resourceIDs)-1] = resourceIDs[len(resourceIDs)-1], resourceIDs[idx]
	resourceIDs = resourceIDs[:len(resourceIDs)-1]

	a[action] = resourceIDs
	return true
}

func (b *basicAuthzRepo) RemovePolicy(ctx context.Context, sub, resourceType, resourceID, action string) error {
	policiesCollection := b.db.Collection("policies")

	doc := policiesCollection.FindOne(ctx, bson.M{"sub": sub})
	if doc.Err() != nil {
		return doc.Err()
	}

	var existingPolicy *policyDoc
	if err := doc.Decode(&existingPolicy); err != nil {
		return err
	}

	// it would remove duplications as well
	for removePolicy(existingPolicy, resourceType, resourceID, action) {
	}

	_, err := policiesCollection.ReplaceOne(ctx, bson.M{"_id": existingPolicy.ID}, existingPolicy)
	return err
}

func (b *basicAuthzRepo) RemovePolicyBySub(ctx context.Context, sub string) error {
	policiesCollection := b.db.Collection("policies")

	result, err := policiesCollection.DeleteOne(ctx, bson.M{"sub": sub})
	if result.DeletedCount <= 0 {
		if err != nil {
			return err
		}
		err = fmt.Errorf("mongo: policy for sub-%s not found", sub)
		return err
	}
	return nil
}

func (b *basicAuthzRepo) ListPolicy(ctx context.Context, sub, resourceType string) (policies []string) {
	policiesCollection := b.db.Collection("policies")

	doc := policiesCollection.FindOne(
		ctx, bson.M{"sub": sub},
		&options.FindOneOptions{Projection: bson.M{fmt.Sprintf("policies.%s", resourceType): 1}},
	)

	if doc.Err() != nil {
		fmt.Println(doc.Err())
		return
	}

	var existingPolicy *policyDoc
	if err := doc.Decode(&existingPolicy); err != nil {
		return
	}

	existingPolicy.Sub = sub
	return buildPolicies(existingPolicy)
}

func buildPolicies(p *policyDoc) (policies []string) {
	for resourceType, actions := range p.Policies {
		for action, resourceIDs := range actions {
			for _, rid := range resourceIDs {
				policies = append(
					policies, fmt.Sprintf("%s:%s:%s:%s", p.Sub, resourceType, action, rid))
			}
		}
	}
	return
}
