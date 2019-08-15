package location

// Location set of storage locations
type Location struct {
	ID          string `bson:"id" json:"id"`                   // e.g. ANLs3 or local for local store
	Description string `bson:"description" json:"description"` // e.g. ANL official S3 service
	Type        string `bson:"type" json:"type"`               // e.g. S3
	URL         string `bson:"url" json:"url"`                 // e.g. http://s3api.invalid.org/download&id=
	Token       string `bson:"token" json:"-"`                 // e.g.g S3 Key or password
	Prefix      string `bson:"prefix" json:"-"`                // e.g.g S3 Bucket or username
}

// type Location struct {
// 	ID          string      `bson:"id" json:"id"`                   // e.g. ANLs3 or local for local store
// 	Description string      `bson:"description" json:"description"` // e.g. ANL official S3 service
// 	Type        string      `bson:"type" json:"type"`               // e.g. S3
// 	Config      interface{} `bson:"-" json:"-"`
// }

// type LocationS3 struct {
// 	URL    string `bson:"url" json:"url"`  // e.g. http://s3api.invalid.org/download&id=
// 	Token  string `bson:"token" json:"-"`  // e.g.g S3 Key or password
// 	Prefix string `bson:"prefix" json:"-"` // e.g.g S3 Bucket or username
// }
