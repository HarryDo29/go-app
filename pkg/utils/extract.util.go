package utils

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ExtractCount(data bson.M, key string) int64 {
	arr, ok := data[key].(primitive.A) // step 1: unwrap the array
	if !ok || len(arr) == 0 {
		return 0 // facet returns [] when no docs match
	}
	doc, ok := arr[0].(bson.M) // step 2: get the first (only) doc
	if !ok {
		return 0
	}
	count, _ := doc["count"].(int32) // step 3: $count outputs int32
	return int64(count)              // step 4: widen to int64
}
