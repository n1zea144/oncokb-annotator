module oncokb-annotator

go 1.23.7

require (
	github.mskcc.org/cdsi/cdsi-protobuf/tempo v0.0.0-20250402191850-afb43daaf8d9
	github.mskcc.org/cdsi/tempo-databricks-gateway v0.0.0-00010101000000-000000000000
)

require google.golang.org/protobuf v1.36.6 // indirect

replace github.mskcc.org/cdsi/tempo-databricks-gateway => ./
