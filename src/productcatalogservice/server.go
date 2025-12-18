package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/GoogleCloudPlatform/microservices-demo/src/productcatalogservice/genproto"
	"github.com/sirupsen/logrus"

	// AWS Imports
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

var (
	log          *logrus.Logger
	extraLatency time.Duration
	port         = "3550"
	dynamoClient *dynamodb.Client // Global DynamoDB Client
)

func init() {
	log = logrus.New()
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}
	log.Out = os.Stdout

	// Initialize AWS DynamoDB Client
	// AWS Academy Learner Lab credentials are auto-loaded from environment (LabRole)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	dynamoClient = dynamodb.NewFromConfig(cfg)
	log.Info("AWS DynamoDB Client Initialized")
}

func main() {
	// Simplified Main: Removed complicated Tracing/Profiling to prevent Learner Lab crashes
	log.Info("Tracing disabled (Simplified for AWS Lab).")
	log.Info("Profiling disabled (Simplified for AWS Lab).")

	flag.Parse()

	if s := os.Getenv("EXTRA_LATENCY"); s != "" {
		v, err := time.ParseDuration(s)
		if err != nil {
			log.Fatalf("failed to parse EXTRA_LATENCY (%s) as time.Duration: %+v", v, err)
		}
		extraLatency = v
		log.Infof("extra latency enabled (duration: %v)", extraLatency)
	}

	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	// Graceful Shutdown Handler
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Printf("Received signal: %s", sig)
		os.Exit(0)
	}()

	log.Infof("starting grpc server at :%s", port)
	run(port)
}

func run(port string) string {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal(err)
	}

	srv := grpc.NewServer() // Basic Server, no OpenTelemetry interceptors to reduce complexity

	svc := &productCatalog{}

	pb.RegisterProductCatalogServiceServer(srv, svc)
	healthpb.RegisterHealthServer(srv, svc)

	// Blocking call
	err = srv.Serve(listener)
	if err != nil {
		log.Fatal(err)
	}

	return listener.Addr().String()
}
