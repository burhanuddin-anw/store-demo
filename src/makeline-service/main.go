package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

// Valid database API types
const (
	AZURE_COSMOS_DB_SQL_API = "cosmosdbsql"
)

// initTracer creates and initializes a new OpenTelemetry tracer provider
func initTracer() (*sdktrace.TracerProvider, error) {
	ctx := context.Background()
	
	// Get the collector endpoint from environment or use default
	collectorEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if collectorEndpoint == "" {
		collectorEndpoint = "localhost:4317" // Default collector endpoint
	}
	
	// Create exporter
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(collectorEndpoint),
	)
	if err != nil {
		return nil, err
	}
	
	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("makeline-service"),
			semconv.ServiceVersionKey.String(os.Getenv("APP_VERSION")),
			attribute.String("environment", os.Getenv("ENVIRONMENT")),
		),
	)
	if err != nil {
		return nil, err
	}
	
	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	
	// Set global tracer provider and propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	
	return tp, nil
}

func main() {
	var orderService *OrderService

	// Initialize OpenTelemetry
	tp, err := initTracer()
	if err != nil {
		log.Printf("Failed to initialize tracer: %s", err)
	} else {
		log.Println("OpenTelemetry tracer initialized successfully")
		// Ensure tracer is shut down properly
		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				log.Printf("Error shutting down tracer provider: %v", err)
			}
		}()
	}

	// Get the database API type
	apiType := os.Getenv("ORDER_DB_API")
	switch apiType {
	case "cosmosdbsql":
		log.Printf("Using Azure CosmosDB SQL API")
	default:
		log.Printf("Using MongoDB API")
	}

	// Initialize the database
	orderService, err = initDatabase(apiType)
	if err != nil {
		log.Printf("Failed to initialize database: %s", err)
		os.Exit(1)
	}

	router := gin.Default()
	
	// Add OpenTelemetry middleware
	router.Use(otelgin.Middleware("makeline-service"))
	router.Use(cors.Default())
	router.Use(OrderMiddleware(orderService))
	
	router.GET("/order/fetch", fetchOrders)
	router.GET("/order/:id", getOrder)
	router.PUT("/order", updateOrder)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": os.Getenv("APP_VERSION"),
		})
	})
	
	// Add Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}
	
	router.Run(":" + port)
}

// OrderMiddleware is a middleware function that injects the order service into the request context
func OrderMiddleware(orderService *OrderService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("orderService", orderService)
		c.Next()
	}
}

// Fetches orders from the order queue and stores them in database
func fetchOrders(c *gin.Context) {
	client, ok := c.MustGet("orderService").(*OrderService)
	if !ok {
		log.Printf("Failed to get order service")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Get orders from the queue
	orders, err := getOrdersFromQueue()
	if err != nil {
		log.Printf("Failed to fetch orders from queue: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Save orders to database
	err = client.repo.InsertOrders(orders)
	if err != nil {
		log.Printf("Failed to save orders to database: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Return the orders to be processed
	orders, err = client.repo.GetPendingOrders()
	if err != nil {
		log.Printf("Failed to get pending orders from database: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.IndentedJSON(http.StatusOK, orders)
}

// Gets a single order from database by order ID
func getOrder(c *gin.Context) {
	client, ok := c.MustGet("orderService").(*OrderService)
	if !ok {
		log.Printf("Failed to get order service")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Failed to convert order id to int: %s", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	sanitizedOrderId := strconv.FormatInt(int64(id), 10)

	order, err := client.repo.GetOrder(sanitizedOrderId)
	if err != nil {
		log.Printf("Failed to get order from database: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.IndentedJSON(http.StatusOK, order)
}

// Updates the status of an order
func updateOrder(c *gin.Context) {
	client, ok := c.MustGet("orderService").(*OrderService)
	if !ok {
		log.Printf("Failed to get order service")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// unmarsal the order from the request body
	var order Order
	if err := c.BindJSON(&order); err != nil {
		log.Printf("Failed to unmarshal order: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	id, err := strconv.Atoi(order.OrderID)
	if err != nil {
		log.Printf("Failed to convert order id to int: %s", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	sanitizedOrderId := strconv.FormatInt(int64(id), 10)

	sanitizedOrder := Order{
		OrderID:    sanitizedOrderId,
		CustomerID: order.CustomerID,
		Items:      order.Items,
		Status:     order.Status,
	}

	err = client.repo.UpdateOrder(sanitizedOrder)
	if err != nil {
		log.Printf("Failed to update order status: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.SetAccepted("202")
}

// Gets an environment variable or exits if it is not set
func getEnvVar(varName string, fallbackVarNames ...string) string {
	value := os.Getenv(varName)
	if value == "" {
		for _, fallbackVarName := range fallbackVarNames {
			value = os.Getenv(fallbackVarName)
			if value == "" {
				break
			}
		}
		if value == "" {
			log.Printf("%s is not set", varName)
			if len(fallbackVarNames) > 0 {
				log.Printf("Tried fallback variables: %v", fallbackVarNames)
			}
			os.Exit(1)
		}
	}
	return value
}

// Initializes the database based on the API type
func initDatabase(apiType string) (*OrderService, error) {
	dbName := getEnvVar("ORDER_DB_NAME")

	switch apiType {
	case AZURE_COSMOS_DB_SQL_API:
		dbURI := getEnvVar("AZURE_COSMOS_RESOURCEENDPOINT", "ORDER_DB_URI")
		containerName := getEnvVar("ORDER_DB_CONTAINER_NAME")
		dbPartitionKey := getEnvVar("ORDER_DB_PARTITION_KEY")
		dbPartitionValue := getEnvVar("ORDER_DB_PARTITION_VALUE")

		// check if USE_WORKLOAD_IDENTITY_AUTH is set
		useWorkloadIdentityAuth := os.Getenv("USE_WORKLOAD_IDENTITY_AUTH")
		if useWorkloadIdentityAuth == "" {
			useWorkloadIdentityAuth = "false"
		}

		if useWorkloadIdentityAuth == "true" {
			cosmosRepo, err := NewCosmosDBOrderRepoWithManagedIdentity(dbURI, dbName, containerName, PartitionKey{dbPartitionKey, dbPartitionValue})
			if err != nil {
				return nil, err
			}
			return NewOrderService(cosmosRepo), nil
		} else {
			dbPassword := os.Getenv("ORDER_DB_PASSWORD")
			cosmosRepo, err := NewCosmosDBOrderRepo(dbURI, dbName, containerName, dbPassword, PartitionKey{dbPartitionKey, dbPartitionValue})
			if err != nil {
				return nil, err
			}
			return NewOrderService(cosmosRepo), nil
		}
	default:
		collectionName := getEnvVar("ORDER_DB_COLLECTION_NAME")

		// check if USE_WORKLOAD_IDENTITY_AUTH is set
		useWorkloadIdentityAuth := os.Getenv("USE_WORKLOAD_IDENTITY_AUTH")
		if useWorkloadIdentityAuth == "" {
			useWorkloadIdentityAuth = "false"
		}

		if useWorkloadIdentityAuth == "true" {
			log.Printf("Authenticating with Workload Identity")
			dbListConnStringsURL := getEnvVar("ORDER_DB_LIST_CONNECTION_STRING_URL")
			mongoRepo, err := NewMongoDBOrderRepoWithManagedIdentity(dbListConnStringsURL, dbName, collectionName)
			if err != nil {
				return nil, err
			}
			return NewOrderService(mongoRepo), nil
		} else {
			log.Printf("Authenticating with username and password")
			dbURI := getEnvVar("AZURE_COSMOS_RESOURCEENDPOINT", "ORDER_DB_URI")
			dbUsername := os.Getenv("ORDER_DB_USERNAME")
			dbPassword := os.Getenv("ORDER_DB_PASSWORD")
			mongoRepo, err := NewMongoDBOrderRepo(dbURI, dbName, collectionName, dbUsername, dbPassword)
			if err != nil {
				return nil, err
			}
			return NewOrderService(mongoRepo), nil
		}
	}
}
