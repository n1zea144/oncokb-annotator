package main

import (
	"context"
	"fmt"
	"os"

	tt "github.mskcc.org/cdsi/cdsi-protobuf/tempo/generated/v1/go"
	tdg "github.mskcc.org/cdsi/tempo-databricks-gateway"
)

func main() {

	ctx := context.Background()

	oncokbAnnotator, err := tdg.NewOncoKBAnnotatorService("",
		"https://www.oncokb.org/api/v1/annotate/mutations/byProteinChange")
	if err != nil {
		fmt.Errorf("Failed to create a OncoKBAnnotatorService: %v", err)
		os.Exit(1)
	}

	tempoMessage := tt.TempoMessage{
		CmoSampleId:       "P-0086668-T16-IH4",
		NormalCmoSampleId: "P-0086668-T16-IH4",
		PipelineVersion:   "v1.0",
		OncotreeCode:      "AMLNPM1",
		Events: []*tt.Event{
			&tt.Event{
				HgvspShort:            "p.W288Cfs*12",
				VariantClassification: "Frame_Shift_Ins",
				EntrezGeneId:          "4869",
				HugoSymbol:            "NPM1",
				StartPosition:         "170837543",
				EndPosition:           "170837544",
				NcbiBuild:             "GRCh37",
			},
		},
	}

	err = oncokbAnnotator.AnnotateMutations(ctx, &tempoMessage)
	if err != nil {
		fmt.Errorf("Error annotating mutations: %v", err)
	}
	fmt.Println("Annotated Tempo Message\n%v", tempoMessage)
}
