package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/worker/core"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBNode handles MongoDB database operations
type MongoDBNode struct{}

func (n *MongoDBNode) Type() string {
	return "integration.mongodb"
}

func (n *MongoDBNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	operation := getString(config, "operation", "find")

	// Get credential
	credIDStr := getString(config, "credentialId", "")
	if credIDStr == "" {
		return nil, fmt.Errorf("MongoDB credential is required")
	}

	credID, err := uuid.Parse(credIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid credential ID")
	}

	cred, err := execCtx.GetCredential(credID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	client, err := n.connect(ctx, cred.Data)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer func() { _ = client.Disconnect(ctx) }()

	database := getString(config, "database", getStringFromMap(cred.Data, "database", ""))
	collection := getString(config, "collection", "")

	if database == "" || collection == "" {
		return nil, fmt.Errorf("database and collection are required")
	}

	coll := client.Database(database).Collection(collection)

	switch operation {
	case "find":
		return n.find(ctx, coll, config)
	case "findOne":
		return n.findOne(ctx, coll, config)
	case "insertOne":
		return n.insertOne(ctx, coll, config, execCtx.Input)
	case "insertMany":
		return n.insertMany(ctx, coll, config, execCtx.Input)
	case "updateOne":
		return n.updateOne(ctx, coll, config, execCtx.Input)
	case "updateMany":
		return n.updateMany(ctx, coll, config, execCtx.Input)
	case "deleteOne":
		return n.deleteOne(ctx, coll, config)
	case "deleteMany":
		return n.deleteMany(ctx, coll, config)
	case "aggregate":
		return n.aggregate(ctx, coll, config)
	case "count":
		return n.count(ctx, coll, config)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

func (n *MongoDBNode) connect(ctx context.Context, creds map[string]interface{}) (*mongo.Client, error) {
	uri := getString(creds, "connectionString", "")
	if uri == "" {
		host := getString(creds, "host", "localhost")
		port := getInt(creds, "port", 27017)
		user := getString(creds, "user", "")
		password := getString(creds, "password", "")

		if user != "" && password != "" {
			uri = fmt.Sprintf("mongodb://%s:%s@%s:%d", user, password, host, port)
		} else {
			uri = fmt.Sprintf("mongodb://%s:%d", host, port)
		}
	}

	clientOptions := options.Client().ApplyURI(uri)
	clientOptions.SetConnectTimeout(10 * time.Second)
	clientOptions.SetMaxPoolSize(5)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client, nil
}

func (n *MongoDBNode) find(ctx context.Context, coll *mongo.Collection, config map[string]interface{}) (map[string]interface{}, error) {
	filter := n.parseFilter(config)
	opts := options.Find()

	if limit := getInt(config, "limit", 0); limit > 0 {
		opts.SetLimit(int64(limit))
	}
	if skip := getInt(config, "skip", 0); skip > 0 {
		opts.SetSkip(int64(skip))
	}
	if sort := getMap(config, "sort"); len(sort) > 0 {
		opts.SetSort(sort)
	}
	if projection := getMap(config, "projection"); len(projection) > 0 {
		opts.SetProjection(projection)
	}

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("find failed: %w", err)
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Convert ObjectIDs to strings for JSON compatibility
	for i, doc := range results {
		results[i] = n.convertObjectIDs(doc)
	}

	return map[string]interface{}{
		"documents": results,
		"count":     len(results),
	}, nil
}

func (n *MongoDBNode) findOne(ctx context.Context, coll *mongo.Collection, config map[string]interface{}) (map[string]interface{}, error) {
	filter := n.parseFilter(config)
	opts := options.FindOne()

	if projection := getMap(config, "projection"); len(projection) > 0 {
		opts.SetProjection(projection)
	}

	var result map[string]interface{}
	err := coll.FindOne(ctx, filter, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return map[string]interface{}{
				"document": nil,
				"found":    false,
			}, nil
		}
		return nil, fmt.Errorf("findOne failed: %w", err)
	}

	return map[string]interface{}{
		"document": n.convertObjectIDs(result),
		"found":    true,
	}, nil
}

func (n *MongoDBNode) insertOne(ctx context.Context, coll *mongo.Collection, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	document := getMap(config, "document")
	if len(document) == 0 {
		if inputDoc, ok := input["document"].(map[string]interface{}); ok {
			document = inputDoc
		}
	}

	if len(document) == 0 {
		return nil, fmt.Errorf("document is required")
	}

	result, err := coll.InsertOne(ctx, document)
	if err != nil {
		return nil, fmt.Errorf("insertOne failed: %w", err)
	}

	insertedID := ""
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		insertedID = oid.Hex()
	}

	return map[string]interface{}{
		"success":    true,
		"insertedId": insertedID,
	}, nil
}

func (n *MongoDBNode) insertMany(ctx context.Context, coll *mongo.Collection, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	documents := getArray(config, "documents")
	if len(documents) == 0 {
		if inputDocs, ok := input["documents"].([]interface{}); ok {
			documents = inputDocs
		}
	}

	if len(documents) == 0 {
		return nil, fmt.Errorf("documents are required")
	}

	result, err := coll.InsertMany(ctx, documents)
	if err != nil {
		return nil, fmt.Errorf("insertMany failed: %w", err)
	}

	insertedIDs := make([]string, len(result.InsertedIDs))
	for i, id := range result.InsertedIDs {
		if oid, ok := id.(primitive.ObjectID); ok {
			insertedIDs[i] = oid.Hex()
		}
	}

	return map[string]interface{}{
		"success":      true,
		"insertedIds":  insertedIDs,
		"insertedCount": len(insertedIDs),
	}, nil
}

func (n *MongoDBNode) updateOne(ctx context.Context, coll *mongo.Collection, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	filter := n.parseFilter(config)
	update := getMap(config, "update")

	if len(update) == 0 {
		if inputUpdate, ok := input["update"].(map[string]interface{}); ok {
			update = inputUpdate
		}
	}

	if len(update) == 0 {
		return nil, fmt.Errorf("update is required")
	}

	// Wrap update in $set if not already wrapped
	if _, hasOperator := update["$set"]; !hasOperator {
		update = map[string]interface{}{"$set": update}
	}

	opts := options.Update()
	if getBool(config, "upsert", false) {
		opts.SetUpsert(true)
	}

	result, err := coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, fmt.Errorf("updateOne failed: %w", err)
	}

	return map[string]interface{}{
		"success":       true,
		"matchedCount":  result.MatchedCount,
		"modifiedCount": result.ModifiedCount,
		"upsertedCount": result.UpsertedCount,
	}, nil
}

func (n *MongoDBNode) updateMany(ctx context.Context, coll *mongo.Collection, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	filter := n.parseFilter(config)
	update := getMap(config, "update")

	if len(update) == 0 {
		if inputUpdate, ok := input["update"].(map[string]interface{}); ok {
			update = inputUpdate
		}
	}

	if len(update) == 0 {
		return nil, fmt.Errorf("update is required")
	}

	if _, hasOperator := update["$set"]; !hasOperator {
		update = map[string]interface{}{"$set": update}
	}

	result, err := coll.UpdateMany(ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("updateMany failed: %w", err)
	}

	return map[string]interface{}{
		"success":       true,
		"matchedCount":  result.MatchedCount,
		"modifiedCount": result.ModifiedCount,
	}, nil
}

func (n *MongoDBNode) deleteOne(ctx context.Context, coll *mongo.Collection, config map[string]interface{}) (map[string]interface{}, error) {
	filter := n.parseFilter(config)

	result, err := coll.DeleteOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("deleteOne failed: %w", err)
	}

	return map[string]interface{}{
		"success":      true,
		"deletedCount": result.DeletedCount,
	}, nil
}

func (n *MongoDBNode) deleteMany(ctx context.Context, coll *mongo.Collection, config map[string]interface{}) (map[string]interface{}, error) {
	filter := n.parseFilter(config)

	result, err := coll.DeleteMany(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("deleteMany failed: %w", err)
	}

	return map[string]interface{}{
		"success":      true,
		"deletedCount": result.DeletedCount,
	}, nil
}

func (n *MongoDBNode) aggregate(ctx context.Context, coll *mongo.Collection, config map[string]interface{}) (map[string]interface{}, error) {
	pipeline := getArray(config, "pipeline")
	if len(pipeline) == 0 {
		return nil, fmt.Errorf("pipeline is required")
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregate failed: %w", err)
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	for i, doc := range results {
		results[i] = n.convertObjectIDs(doc)
	}

	return map[string]interface{}{
		"results": results,
		"count":   len(results),
	}, nil
}

func (n *MongoDBNode) count(ctx context.Context, coll *mongo.Collection, config map[string]interface{}) (map[string]interface{}, error) {
	filter := n.parseFilter(config)

	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("count failed: %w", err)
	}

	return map[string]interface{}{
		"count": count,
	}, nil
}

func (n *MongoDBNode) parseFilter(config map[string]interface{}) bson.M {
	filter := getMap(config, "filter")
	if len(filter) == 0 {
		// Try parsing from JSON string
		if filterStr := getString(config, "filterJson", ""); filterStr != "" {
			var f map[string]interface{}
			if err := json.Unmarshal([]byte(filterStr), &f); err == nil {
				filter = f
			}
		}
	}

	if len(filter) == 0 {
		return bson.M{}
	}

	// Convert _id strings to ObjectIDs
	if id, ok := filter["_id"].(string); ok {
		if oid, err := primitive.ObjectIDFromHex(id); err == nil {
			filter["_id"] = oid
		}
	}

	return filter
}

func (n *MongoDBNode) convertObjectIDs(doc map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range doc {
		switch val := v.(type) {
		case primitive.ObjectID:
			result[k] = val.Hex()
		case primitive.DateTime:
			result[k] = val.Time().Format(time.RFC3339)
		case map[string]interface{}:
			result[k] = n.convertObjectIDs(val)
		case []interface{}:
			arr := make([]interface{}, len(val))
			for i, item := range val {
				if m, ok := item.(map[string]interface{}); ok {
					arr[i] = n.convertObjectIDs(m)
				} else if oid, ok := item.(primitive.ObjectID); ok {
					arr[i] = oid.Hex()
				} else {
					arr[i] = item
				}
			}
			result[k] = arr
		default:
			result[k] = v
		}
	}
	return result
}
