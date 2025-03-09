package service

import (
	"context"
	"math"
	"simple_mongo_grpc/cmd/models"
	PaggingPb "simple_mongo_grpc/pb/pagination"
	productPb "simple_mongo_grpc/pb/product"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductService struct {
	productPb.UnimplementedProductServiceServer
	DB *mongo.Database
}

func (ps *ProductService) CreateProduct(ctx context.Context, request *productPb.Product) (*productPb.Id, error) {

	var Response productPb.Id

	collection := ps.DB.Collection("product")

	product := bson.M{
		"name":  request.GetName(),
		"price": request.GetStock(),
		"stock": request.GetStock(),
		"category": bson.M{
			"name": request.GetCategory().GetName(),
		},
	}

	result, err := collection.InsertOne(ctx, product)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	insertedID, ok := result.InsertedID.(bson.ObjectID)
	if !ok {
		return nil, status.Error(codes.Internal, "InsertedID is not an ObjectID")
	}
	Response.Id = insertedID.Hex()

	return &Response, nil
}

func (ps *ProductService) GetProducts(ctx context.Context, req *PaggingPb.Pagination) (*productPb.Products, error) {

	collection := ps.DB.Collection("product")

	limit := req.GetLimit()
	if limit < 1 {
		limit = 10
	}
	page := req.GetCurrentPage()
	if page < 1 {
		page = 1
	}

	filter := bson.M{}

	//count total data
	totalCount, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	totalPage := uint32(math.Ceil(float64(totalCount) / float64(limit)))

	//find with pagination
	cursor, err := collection.Find(ctx, filter, options.Find().
		SetSkip(int64((page-1)*limit)).
		SetLimit(int64(limit)))

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	defer cursor.Close(ctx)

	var products []*productPb.Product

	for cursor.Next(ctx) {
		var product models.Product

		if err := cursor.Decode(&product); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		products = append(products, product.ToProto())
	}

	if err := cursor.Err(); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &productPb.Products{
		Pagination: &PaggingPb.Pagination{
			CurrentPage: page,
			Limit:       limit,
			TotalRecord: uint32(totalCount),
			TotalPage:   totalPage,
		},
		Data: products,
	}, nil

}

func (ps *ProductService) GetProduct(ctx context.Context, req *productPb.Id) (*productPb.Product, error) {

	collection := ps.DB.Collection("product")

	var product models.Product

	objectId, err := bson.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := collection.FindOne(ctx, bson.M{"_id": objectId}).Decode(&product); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return product.ToProto(), nil
}

func (ps *ProductService) UpdateProduct(ctx context.Context, req *productPb.Product) (*productPb.Status, error) {

	collection := ps.DB.Collection("product")

	objectid, err := bson.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	updateData := models.Product{
		Name:  req.GetName(),
		Price: req.GetPrice(),
		Stock: req.GetStock(),
		Category: models.Category{
			Name: req.GetCategory().GetName(),
		},
	}

	filter := bson.M{"_id": objectid}

	update := bson.M{"$set": updateData}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if result.MatchedCount == 0 {
		return nil, status.Error(codes.NotFound, "Product not found")
	}

	return &productPb.Status{Status: "update product succesfully"}, nil

}

func (ps *ProductService) DeleteProduct(ctx context.Context, req *productPb.Id) (*productPb.Status, error) {

	collection := ps.DB.Collection("product")

	objectId, err := bson.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	filter := bson.M{"_id": objectId}

	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if result.DeletedCount == 0 {
		return nil, status.Error(codes.NotFound, "Product not found")
	}

	return &productPb.Status{Status: "Success delete product"}, nil
}
