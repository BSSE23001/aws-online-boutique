// package main

// import (
// 	"context"
// 	"strings"
// 	"time"

// 	pb "github.com/GoogleCloudPlatform/microservices-demo/src/productcatalogservice/genproto"
// 	"google.golang.org/grpc/codes"
// 	healthpb "google.golang.org/grpc/health/grpc_health_v1"
// 	"google.golang.org/grpc/status"
// )

// type productCatalog struct {
// 	catalog pb.ListProductsResponse
// }

// func (p *productCatalog) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
// 	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
// }

// func (p *productCatalog) Watch(req *healthpb.HealthCheckRequest, ws healthpb.Health_WatchServer) error {
// 	return status.Errorf(codes.Unimplemented, "health check via Watch not implemented")
// }

// func (p *productCatalog) ListProducts(context.Context, *pb.Empty) (*pb.ListProductsResponse, error) {
// 	time.Sleep(extraLatency)

// 	return &pb.ListProductsResponse{Products: p.parseCatalog()}, nil
// }

// func (p *productCatalog) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
// 	time.Sleep(extraLatency)

// 	var found *pb.Product
// 	for i := 0; i < len(p.parseCatalog()); i++ {
// 		if req.Id == p.parseCatalog()[i].Id {
// 			found = p.parseCatalog()[i]
// 		}
// 	}

// 	if found == nil {
// 		return nil, status.Errorf(codes.NotFound, "no product with ID %s", req.Id)
// 	}
// 	return found, nil
// }

// func (p *productCatalog) SearchProducts(ctx context.Context, req *pb.SearchProductsRequest) (*pb.SearchProductsResponse, error) {
// 	time.Sleep(extraLatency)

// 	var ps []*pb.Product
// 	for _, product := range p.parseCatalog() {
// 		if strings.Contains(strings.ToLower(product.Name), strings.ToLower(req.Query)) ||
// 			strings.Contains(strings.ToLower(product.Description), strings.ToLower(req.Query)) {
// 			ps = append(ps, product)
// 		}
// 	}

// 	return &pb.SearchProductsResponse{Results: ps}, nil
// }

// func (p *productCatalog) parseCatalog() []*pb.Product {
// 	if reloadCatalog || len(p.catalog.Products) == 0 {
// 		err := readCatalogFile(&p.catalog)
// 		if err != nil {
// 			return []*pb.Product{}
// 		}
// 	}

// 	return p.catalog.Products
// }

package main

import (
	"context"
	"strings"

	pb "github.com/GoogleCloudPlatform/microservices-demo/src/productcatalogservice/genproto"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type productCatalog struct {
	// No longer need to store catalog in memory
}

// Struct to match your DynamoDB Item exactly
type ProductItem struct {
	ID          string   `dynamodbav:"id"`
	Name        string   `dynamodbav:"name"`
	Description string   `dynamodbav:"description"`
	Picture     string   `dynamodbav:"picture"`
	PriceUsd    Price    `dynamodbav:"priceUsd"`
	Categories  []string `dynamodbav:"categories"`
}

type Price struct {
	CurrencyCode string `dynamodbav:"currencyCode"`
	Units        int64  `dynamodbav:"units"`
	Nanos        int32  `dynamodbav:"nanos"`
}

func (p *productCatalog) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (p *productCatalog) Watch(req *healthpb.HealthCheckRequest, ws healthpb.Health_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "health check via Watch not implemented")
}

func (p *productCatalog) ListProducts(ctx context.Context, _ *pb.Empty) (*pb.ListProductsResponse, error) {
	// Scan is okay for small catalogs (like this demo). For production, use Query.
	out, err := dynamoClient.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String("Products"),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to scan dynamodb: %v", err)
	}

	var items []ProductItem
	err = attributevalue.UnmarshalListOfMaps(out.Items, &items)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmarshal dynamodb items: %v", err)
	}

	var products []*pb.Product
	for _, item := range items {
		products = append(products, p.convertToProto(item))
	}

	return &pb.ListProductsResponse{Products: products}, nil
}

func (p *productCatalog) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	out, err := dynamoClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("Products"),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: req.Id},
		},
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "dynamodb error: %v", err)
	}

	if out.Item == nil {
		return nil, status.Errorf(codes.NotFound, "no product with ID %s", req.Id)
	}

	var item ProductItem
	err = attributevalue.UnmarshalMap(out.Item, &item)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmarshal item: %v", err)
	}

	return p.convertToProto(item), nil
}

func (p *productCatalog) SearchProducts(ctx context.Context, req *pb.SearchProductsRequest) (*pb.SearchProductsResponse, error) {
	// Note: Implementing true search in DynamoDB requires logic or ElasticSearch.
	// For this demo, we will SCAN all and filter in memory (inefficient but works for small demo).

	productsResponse, err := p.ListProducts(ctx, nil)
	if err != nil {
		return nil, err
	}

	var results []*pb.Product
	query := strings.ToLower(req.Query)
	for _, product := range productsResponse.Products {
		if strings.Contains(strings.ToLower(product.Name), query) ||
			strings.Contains(strings.ToLower(product.Description), query) {
			results = append(results, product)
		}
	}

	return &pb.SearchProductsResponse{Results: results}, nil
}

// Helper to convert internal Struct to GRPC Proto
func (p *productCatalog) convertToProto(item ProductItem) *pb.Product {
	return &pb.Product{
		Id:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Picture:     item.Picture,
		PriceUsd: &pb.Money{
			CurrencyCode: item.PriceUsd.CurrencyCode,
			Units:        item.PriceUsd.Units,
			Nanos:        item.PriceUsd.Nanos,
		},
		Categories: item.Categories,
	}
}
