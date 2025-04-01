module oncokb-annotator

go 1.23.7

require (
	github.mskcc.org/cdsi/cdsi-protobuf/tempo v0.0.0-20250318020142-e6473b3ddb77
	github.mskcc.org/cdsi/tempo-databricks-gateway v0.0.0-20250127192401-f611fccd64d7
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
)

replace github.mskcc.org/cdsi/tempo-databricks-gateway => ./
