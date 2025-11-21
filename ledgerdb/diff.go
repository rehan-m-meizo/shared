package ledgerdb

import "go.mongodb.org/mongo-driver/bson"

func Diff(oldData, newData bson.M) map[string][2]any {
	diff := make(map[string][2]any)

	for k, oldV := range oldData {
		if newV, ok := newData[k]; ok {
			if oldV != newV {
				diff[k] = [2]any{oldV, newV}
			}
		} else {
			diff[k] = [2]any{oldV, nil}
		}
	}

	for k, newV := range newData {
		if _, ok := oldData[k]; !ok {
			diff[k] = [2]any{nil, newV}
		}
	}
	return diff
}
