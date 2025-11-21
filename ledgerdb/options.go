package ledgerdb

import "go.mongodb.org/mongo-driver/mongo/options"

func mongoOptionsLatest() *options.FindOneOptions {
	return options.FindOne().SetSort(map[string]int{"version": -1})
}

func mongoOptionsAllVersions() *options.FindOptions {
	return options.Find().SetSort(map[string]int{"version": 1})
}
