module github.com/itsmurugappan/k8s-resource-backup

go 1.13

require (
	contrib.go.opencensus.io/exporter/prometheus v0.1.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.12.8 // indirect
	github.com/google/go-containerregistry v0.0.0-20191218175032-34fb8ff33bed // indirect
	github.com/lib/pq v1.3.0
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
	go.opencensus.io v0.22.2 // indirect
	go.uber.org/zap v1.13.0 // indirect
	golang.org/x/net v0.0.0-20191209160850-c0dbc17a3553 // indirect
	istio.io/client-go v0.0.0-20191218043923-5fad2566daf6
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	knative.dev/pkg v0.0.0-20200102192742-169ef0797c1f
	knative.dev/serving v0.11.1
)

replace contrib.go.opencensus.io/exporter/stackdriver => contrib.go.opencensus.io/exporter/stackdriver v0.12.9-0.20191108183826-59d068f8d8ff

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
