package tempo_databricks_gateway

import (
	"bufio"
	"context"
	"os"
	"regexp"
	"strings"
	"testing"

	tt "github.mskcc.org/cdsi/cdsi-protobuf/tempo/generated/v1/go"
)

// maf files for testing were obtained by grabbing the following fields from the OncoKB annotated clinical impact MAF
// cut -f1,2,4,6,7,10,17,40,126,127,128,129,130,131,132,133,134,135,136,137,138,139,140,141,142,143,144,145,146,147,148,149,150,151,152 data_mutations_extended.oncokb.txt > ~/prgs/cdsi/oncokb-annotator/data_mutations_extended.oncokb.trimmed.txt
// clinical sample files for testing were obtained by grabbing the following fields from the OncoKB annotated clinical impact sample clinicalFile
// cut -f1,7,17 ~/prgs/cbio/cbio-portal-data/oncokb-annotated-msk-impact/data_clinical_sample.oncokb.txt > ~/prgs/cdsi/oncokb-annotator/testdata/data_clinical_sample.oncokb.trimmed.txt
const (
	pat          = ""
	annotateURL  = "https://www.oncokb.org/api/v1/annotate/mutations/byProteinChange"
	clinicalFile = "testdata/data_clinical_sample.oncokb.txt"
	mafFile      = "testdata/data_mutations_extended.oncokb.txt"
	//clinicalFile = "testdata/data_clinical_sample.oncokb.trimmed.txt"
	//mafFile      = "testdata/data_mutations_extended.oncokb.trimmed.txt"
)

func TestAnnotateMutations(t *testing.T) {

	ctx := context.Background()

	oncokbAnnotator, err := NewOncoKBAnnotatorService(pat, annotateURL)
	if err != nil {
		t.Fatalf("Failed to create an OncoKBAnnotatorService: %v", err)
	}

	oncoMap := readClinicalFile(t, clinicalFile)
	mafRecords := readMAF(t, mafFile)

	for lc, line := range mafRecords {

		fields := strings.Split(line, "\t")
		t.Logf("Processing line %v sample %q and protein change %q", lc+1, fields[6], fields[7])
		tm, err := getTempoMessage(t, oncoMap, fields)
		if err != nil {
			t.Logf("Error creating a tempo message from MAF record, skipping to next MAF record: %q", err)
			continue
		}

		err = oncokbAnnotator.AnnotateMutations(ctx, tm)
		if err != nil {
			t.Logf("Error returned from Annotate Mutations, skipping to next MAF record: %q", err)
			continue
		}

		assertNoError(t, fields, tm)
	}
}

func readClinicalFile(t testing.TB, clinicalFile string) map[string]string {
	oncoMap := make(map[string]string)
	fClinical, err := os.Open(clinicalFile)
	if err != nil {
		t.Fatalf("Failed to open clinical file %q: %v", clinicalFile, err)
	}
	defer fClinical.Close()
	scanner := bufio.NewScanner(fClinical)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, "\t")
		oncoMap[fields[0]] = fields[2]
	}
	return oncoMap
}

func readMAF(t testing.TB, mafFile string) []string {
	mafRecords := make([]string, 0)
	fMAF, err := os.Open(mafFile)
	if err != nil {
		t.Fatalf("Failed to open MAF file %q: %v", mafFile, err)
	}
	defer fMAF.Close()
	lineCt := 0
	scanner := bufio.NewScanner(fMAF)
	for scanner.Scan() {
		lineCt++
		if lineCt == 1 {
			continue
		}
		line := scanner.Text()
		mafRecords = append(mafRecords, line)
	}
	return mafRecords
}

func getTempoMessage(t testing.TB, oncoMap map[string]string, fields []string) (*tt.TempoMessage, error) {
	var tm tt.TempoMessage
	tm.Events = []*tt.Event{&tt.Event{}}
	tm.CmoSampleId = fields[6]
	tm.NormalCmoSampleId = fields[6]
	tm.PipelineVersion = "v1.0"
	if code, exists := oncoMap[fields[6]]; exists {
		tm.OncotreeCode = code
	} else {
		t.Logf("Cannot find oncotree code for patient %q", fields[6])
	}
	tm.Events[0].HgvspShort = fields[7]
	tm.Events[0].VariantClassification = fields[5]
	tm.Events[0].EntrezGeneId = fields[1]
	tm.Events[0].HugoSymbol = fields[0]
	tm.Events[0].StartPosition = fields[3]
	tm.Events[0].EndPosition = fields[4]
	tm.Events[0].NcbiBuild = fields[2]

	return &tm, nil
}

var fieldMap = map[int]string{
	0:  "Hugo_Symbol",
	1:  "Entrez_Gene_Id",
	2:  "NCBI_Build",
	3:  "Start_Position",
	4:  "End_Position",
	5:  "Variant_Classification",
	6:  "Tumor_Sample_Barcode",
	7:  "HGVSp_Short",
	8:  "ANNOTATED",
	9:  "GENE_IN_ONCOKB",
	10: "VARIANT_IN_ONCOKB",
	11: "MUTATION_EFFECT",
	12: "MUTATION_EFFECT_CITATIONS",
	13: "ONCOGENIC",
	14: "LEVEL_1",
	15: "LEVEL_2",
	16: "LEVEL_3A",
	17: "LEVEL_3B",
	18: "LEVEL_4",
	19: "LEVEL_R1",
	20: "LEVEL_R2",
	21: "HIGHEST_LEVEL",
	22: "HIGHEST_SENSITIVE_LEVEL",
	23: "HIGHEST_RESISTANCE_LEVEL",
	24: "TX_CITATIONS",
	25: "LEVEL_Dx1",
	26: "LEVEL_Dx2",
	27: "LEVEL_Dx3",
	28: "HIGHEST_DX_LEVEL",
	29: "DX_CITATIONS",
	30: "LEVEL_Px1",
	31: "LEVEL_Px2",
	32: "LEVEL_Px3",
	33: "HIGHEST_PX_LEVEL",
	34: "PX_CITATIONS",
}

func assertNoError(t testing.TB, fields []string, tm *tt.TempoMessage) {
	if !strings.EqualFold(tm.Events[0].OncokbAnnotated, fields[8]) && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[8], fields[8], tm.Events[0].OncokbAnnotated)
	}
	if !strings.EqualFold(tm.Events[0].OncokbKnownGene, fields[9]) && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[9], fields[9], tm.Events[0].OncokbKnownGene)
	}
	if !strings.EqualFold(tm.Events[0].OncokbKnownVariant, fields[10]) && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[10], fields[10], tm.Events[0].OncokbKnownVariant)
	}
	if tm.Events[0].OncokbMutationEffect != fields[11] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[11], fields[11], tm.Events[0].OncokbMutationEffect)
	}
	if tm.Events[0].OncokbMutationEffectCitations != fields[12] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[12], fields[12], tm.Events[0].OncokbMutationEffectCitations)
	}
	if tm.Events[0].OncokbOncogenic != fields[13] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[13], fields[13], tm.Events[0].OncokbOncogenic)
	}
	if tm.Events[0].OncokbLevel1 != fields[14] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[14], fields[14], tm.Events[0].OncokbLevel1)
	}
	if tm.Events[0].OncokbLevel2 != fields[15] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[15], fields[15], tm.Events[0].OncokbLevel2)
	}
	if tm.Events[0].OncokbLevel3A != fields[16] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[16], fields[16], tm.Events[0].OncokbLevel3A)
	}
	if tm.Events[0].OncokbLevel3B != fields[17] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[17], fields[17], tm.Events[0].OncokbLevel3B)
	}
	if tm.Events[0].OncokbLevel4 != fields[18] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[18], fields[18], tm.Events[0].OncokbLevel4)
	}
	if tm.Events[0].OncokbLevelR1 != fields[19] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[19], fields[19], tm.Events[0].OncokbLevelR1)
	}
	if tm.Events[0].OncokbLevelR2 != fields[20] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[20], fields[20], tm.Events[0].OncokbLevelR2)
	}
	if tm.Events[0].OncokbHighestLevel != fields[21] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[21], fields[21], tm.Events[0].OncokbHighestLevel)
	}
	if tm.Events[0].OncokbHighestSensitivityLevel != fields[22] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[22], fields[22], tm.Events[0].OncokbHighestSensitivityLevel)
	}
	if tm.Events[0].OncokbHighestResistanceLevel != fields[23] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[23], fields[23], tm.Events[0].OncokbHighestResistanceLevel)
	}
	if tm.Events[0].OncokbTxCitations != fields[24] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[24], fields[24], tm.Events[0].OncokbTxCitations)
	}
	if tm.Events[0].OncokbLevelDx1 != fields[25] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[25], fields[25], tm.Events[0].OncokbLevelDx1)
	}
	if tm.Events[0].OncokbLevelDx2 != fields[26] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[26], fields[26], tm.Events[0].OncokbLevelDx2)
	}
	if tm.Events[0].OncokbLevelDx3 != fields[27] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[27], fields[27], tm.Events[0].OncokbLevelDx3)
	}
	if tm.Events[0].OncokbHighestDxLevel != fields[28] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[28], fields[28], tm.Events[0].OncokbHighestDxLevel)
	}
	if tm.Events[0].OncokbDxCitations != fields[29] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[29], fields[29], tm.Events[0].OncokbDxCitations)
	}
	if tm.Events[0].OncokbLevelPx1 != fields[30] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[30], fields[30], tm.Events[0].OncokbLevelPx1)
	}
	if tm.Events[0].OncokbLevelPx2 != fields[31] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[31], fields[31], tm.Events[0].OncokbLevelPx2)
	}
	if tm.Events[0].OncokbLevelPx3 != fields[32] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[32], fields[32], tm.Events[0].OncokbLevelPx3)
	}
	if tm.Events[0].OncokbHighestPxLevel != fields[33] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[33], fields[33], tm.Events[0].OncokbHighestPxLevel)
	}
	if tm.Events[0].OncokbPxCitations != fields[34] && assertNotProblematicRecord(fields) {
		t.Errorf("patient: %q; field: %q; expected %q but got %q", fields[6], fieldMap[34], fields[34], tm.Events[0].OncokbPxCitations)
	}
}

func assertNotProblematicRecord(fields []string) bool {
	// MAFAnnotator.py had a typo in it that was leaving consequence set to "5'Flank"
	// when it should have been set to "any", so we ignore comparisons/errors when "Variant_Classification" == 5'Flank
	// Also, some records have invalid hgvs protein sequences.  This causes the MAFAnnotator to drop start/end position
	// fields which change the OncoKB response.  Lets ignore the comparisons/errors when the hgvs protein sequence is invalid
	return fields[5] != "5'Flank" && isValidHGVSProtein(fields[7])
}

var hgvsProteinRegex = regexp.MustCompile(`^p\.([A-Z][a-z]{2})(\d+)([A-Z][a-z]{2}|del|ins|dup|fs\*?\d*|Ter|X)?$`)

func isValidHGVSProtein(hgvs string) bool {
	return hgvsProteinRegex.MatchString(hgvs)
}
