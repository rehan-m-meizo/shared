package ledgerdb

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

func ComputeHash(data bson.M, previousHash string) string {
	combined := bson.M{
		"data":         data,
		"previousHash": previousHash,
	}

	bytes, _ := json.Marshal(combined)
	sum := sha256.Sum256(bytes)
	return fmt.Sprintf("%x", sum)
}
