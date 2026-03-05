// NOTE: golib and tinyftp are pinned to vendored versions whose APIs differ
// from latest upstream. Always build with -mod=vendor.
module github.com/MG-RAST/Shock

go 1.22

require (
	cloud.google.com/go v0.45.2-0.20190912001407-6ae710637747
	github.com/Azure/azure-pipeline-go v0.2.2
	github.com/Azure/azure-storage-blob-go v0.8.0
	github.com/MG-RAST/Shock/clients/shock-go v0.0.0-00010101000000-000000000000
	github.com/MG-RAST/go-shock-client v0.0.0-20190828185941-363c96852f00
	github.com/MG-RAST/golib v0.0.0-20190510221542-86643de6f9e0
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/go-chi/chi/v5 v5.2.5
	github.com/golang/groupcache v0.0.0-20190702054246-869f871628b6
	github.com/golang/protobuf v1.3.3-0.20190827175835-822fe56949f5
	github.com/googleapis/gax-go/v2 v2.0.6-0.20190829211521-e18837313485
	github.com/jum/tinyftp v1.0.0
	github.com/mattn/go-ieproxy v0.0.0-20190805055040-f9202b1cfdeb
	github.com/stretchr/testify v1.10.0
	go.opencensus.io v0.22.2-0.20190911211948-65310139a05d
	golang.org/x/net v0.33.0
	golang.org/x/oauth2 v0.25.0
	golang.org/x/sys v0.28.0
	golang.org/x/text v0.21.0
	google.golang.org/api v0.10.1-0.20190918000732-634b73c1f50b
	google.golang.org/appengine v1.6.2
	google.golang.org/genproto v0.0.0-20190916214212-f660b8655731
	google.golang.org/grpc v1.21.1
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/MG-RAST/Shock/clients/shock-go => ./clients/shock-go
