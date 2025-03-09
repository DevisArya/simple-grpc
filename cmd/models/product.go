package models

import (
	productPb "simple_mongo_grpc/pb/product"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Product struct {
	ID       bson.ObjectID `bson:"_id,omitempty"`
	Name     string        `bson:"name"`
	Price    float64       `bson:"price"`
	Stock    uint32        `bson:"stock"`
	Category Category      `bson:"category"`
}

type Category struct {
	Name string `bson:"name"`
}

func (p *Product) ToProto() *productPb.Product {
	return &productPb.Product{
		Id:    p.ID.Hex(),
		Name:  p.Name,
		Price: p.Price,
		Stock: p.Stock,
		Category: &productPb.Category{
			Name: p.Category.Name,
		},
	}
}

func ProductFromProto(pb *productPb.Product) (*Product, error) {

	ObjectId, err := bson.ObjectIDFromHex(pb.Id)
	if err != nil {
		return nil, err
	}

	return &Product{
		ID:    ObjectId,
		Name:  pb.Name,
		Price: pb.Price,
		Stock: pb.Stock,
		Category: Category{
			Name: pb.Category.Name,
		},
	}, nil
}
