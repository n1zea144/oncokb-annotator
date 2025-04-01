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

	// 838323
	pChange := "p.W288Cfs*12"
	vc := "Frame_Shift_Ins"
	eId := "4869"
	hSymbol := "NPM1"
	build := "GRCh37"
	tempoMessage := tt.TempoMessage{
		CmoSampleId:       "P-0086668-T16-IH4",
		NormalCmoSampleId: "P-0086668-T16-IH4",
		PipelineVersion:   "v1.0",
		OncotreeCode:      "AMLNPM1",
		Events: []*tt.Event{
			&tt.Event{
				HgvspShort:            &pChange,
				VariantClassification: &vc,
				EntrezGeneId:          &eId,
				HugoSymbol:            &hSymbol,
				StartPosition:         170837543,
				EndPosition:           170837544,
				NcbiBuild:             &build,
			},
		},
	}

	// pChange := "p.S428F"
	// vc := "Missense_Mutation"
	// eId := "11200"
	// hSymbol := "CHEK2"
	// build := "GRCh37"
	// tempoMessage := tt.TempoMessage{
	// 	CmoSampleId:       "P-0041863-T01-IM6",
	// 	NormalCmoSampleId: "P-0041863-N01-IM6",
	// 	PipelineVersion:   "v1.0",
	// 	OncotreeCode:      "CCRCC",
	// 	Events: []*tt.Event{
	// 		&tt.Event{
	// 			HgvspShort:            &pChange,
	// 			VariantClassification: &vc,
	// 			EntrezGeneId:          &eId,
	// 			HugoSymbol:            &hSymbol,
	// 			StartPosition:         428,
	// 			EndPosition:           428,
	// 			NcbiBuild:             &build,
	// 		},
	// 	},
	// }

	oncokbAnnotator.AnnotateMutations(ctx, tempoMessage)
}
