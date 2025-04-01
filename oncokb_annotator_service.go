package tempo_databricks_gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	tt "github.mskcc.org/cdsi/cdsi-protobuf/tempo/generated/v1/go"
)

type OncoKBAnnotatorService struct {
	pat       string
	oncokbURL string
}

func NewOncoKBAnnotatorService(token, oncokbURL string) (*OncoKBAnnotatorService, error) {
	if len(token) == 0 || len(oncokbURL) == 0 {
		return nil, fmt.Errorf("Both token: %q and oncokbURL: %q need to be valid", token, oncokbURL)
	}
	return &OncoKBAnnotatorService{pat: token, oncokbURL: oncokbURL}, nil
}

func (o OncoKBAnnotatorService) AnnotateMutations(ctx context.Context, message *tt.TempoMessage) error {
	// we need to strip p. from change
	jsonData, err := getOncoKBRequestJSON(strings.Contains(o.oncokbURL, "byProteinChange"), message)
	if err != nil {
		return fmt.Errorf("Error creating OncoKB request body %s", err)
	}
	req, err := http.NewRequest(http.MethodPost, o.oncokbURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("Error creating http request: %s", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.pat))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error creating http client: %s", err)
	}

	oncoKBResponse, err := getOncoKBResponse(resp)
	if err != nil {
		return err
	}

	mapResponseToEvents(message.Events, oncoKBResponse)
	return nil
}

var variantClassToConsequence = map[string][]string{
	"3'flank":                 []string{"any"},
	"3'utr":                   []string{"any"},
	"5'flank":                 []string{"any"},
	"5'utr":                   []string{"any"},
	"intron":                  []string{"any"},
	"intronic":                []string{"any"},
	"igr":                     []string{"any"},
	"targeted_region":         []string{"inframe_deletion", "inframe_insertion"},
	"complex_indel":           []string{"inframe_deletion", "inframe_insertion"},
	"essential_splice_site":   []string{"feature_truncation"},
	"exon skipping":           []string{"inframe_deletion"},
	"frameshift deletion":     []string{"frameshift_variant"},
	"frameshift_deletion":     []string{"frameshift_variant"},
	"frameshift insertion":    []string{"frameshift_variant"},
	"frameshift_insertion":    []string{"frameshift_variant"},
	"nonframeshift_deletion":  []string{"any"},
	"nonframeshift_insertion": []string{"any"},
	"frameshift_coding":       []string{"frameshift_variant"},
	"frame_shift_del":         []string{"frameshift_variant"},
	"frame_shift_ins":         []string{"frameshift_variant"},
	"fusion":                  []string{"fusion"},
	"indel":                   []string{"frameshift_variant", "inframe_deletion", "inframe_insertion"},
	"in_frame_del":            []string{"inframe_deletion"},
	"in_frame_ins":            []string{"inframe_insertion"},
	"missense":                []string{"missense_variant"},
	"missense_mutation":       []string{"missense_variant"},
	"nonsynonymous_snv":       []string{"missense_variant"},
	"nonsense_mutation":       []string{"stop_gained"},
	"stopgain_snv":            []string{"stop_gained"},
	"nonstop_mutation":        []string{"stop_lost"},
	"stoploss_snv":            []string{"stop_lost"},
	"silent":                  []string{"silent"},
	"splice_site":             []string{"splice_region_variant"},
	"splice_site_del":         []string{"splice_region_variant"},
	"splice_site_snp":         []string{"splice_region_variant"},
	"splicing":                []string{"splice_region_variant"},
	"splice_region":           []string{"splice_region_variant"},
	"rna":                     []string{"splice_region_variant"},
	"translation_start_site":  []string{"start_lost"},
	"viii deletion":           []string{"any"},
}

func getOncoKBRequestJSON(byProteinChangeURL bool, message *tt.TempoMessage) ([]byte, error) {
	var oncoKBMutations []OncoKBMutationRequest
	var proteinStart, proteinEnd int
	for lc, ev := range message.Events {
		var gID int
		if len(*ev.HugoSymbol) == 0 {
			gID, _ = strconv.Atoi(*ev.EntrezGeneId) // this should be an integer in protobuf def
		}
		eIndex := strconv.Itoa(lc)
		var consequence string
		if consList, ok := variantClassToConsequence[strings.ToLower(*ev.VariantClassification)]; !ok {
			return nil, fmt.Errorf("An unknown variant classification has been encountered: %s", *ev.VariantClassification)
		} else {
			consequence = strings.Join(consList, "+")
		}
		// protein start/end are not set by AnnotatorCore.py:process_alteration when
		// the query type is byProteinChange.  This is the query type used when HGVSP_SHORT
		// is present and this is the mode used when annotating the nightly clinical IMPACT files.
		// When we set start/end, we get differing results from the script, so lets not set them
		// when the query is by protein change
		if !byProteinChangeURL {
			proteinStart = int(ev.StartPosition)
			proteinEnd = int(ev.EndPosition)
		}
		if len(*ev.HgvspShort) == 0 {
			return nil, fmt.Errorf("HGVS_Short is missing, cannot proceed")
		}
		request := OncoKBMutationRequest{
			Alteration:  strings.TrimPrefix(*ev.HgvspShort, "p."), // strip leading "p.'
			Consequence: consequence,
			Gene: Gene{
				EntrezGeneID: gID,
				HugoSymbol:   *ev.HugoSymbol,
			},
			ID:              eIndex,
			ProteinStart:    proteinStart,
			ProteinEnd:      proteinEnd,
			ReferenceGenome: *ev.NcbiBuild,
			TumorType:       message.OncotreeCode,
		}
		oncoKBMutations = append(oncoKBMutations, request)
	}
	jsonData, err := json.Marshal(oncoKBMutations)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func getOncoKBResponse(resp *http.Response) ([]OncoKBResponse, error) {
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		errResp, err := unMarshal[OncoKBErrorResponse](string(body))
		if err != nil {
			return nil, fmt.Errorf("Error making OncoKB API request: %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("Error making OncoKB API request: %s", errResp.Message)
	}

	toReturn, err := unMarshal[[]OncoKBResponse](string(body))
	if err != nil {
		return nil, fmt.Errorf("Error reading OncoKB API response: %s", err)
	}
	return toReturn, nil
}

func mapResponseToEvents(events []*tt.Event, resp []OncoKBResponse) {

	for _, r := range resp {
		ind, _ := strconv.Atoi(r.Query.ID)
		e := events[ind]
		e.OncokbAnnotated = "true"
		e.OncokbKnownGene = strconv.FormatBool(r.GeneExist)
		e.OncokbKnownVariant = strconv.FormatBool(r.VariantExist)
		e.OncokbMutationEffect = r.MutationEffect.KnownEffect
		mutationEffectCitations := []Citations{r.MutationEffect.Citations}
		e.OncokbMutationEffectCitations = getCitations[Citations](mutationEffectCitations)
		e.OncokbOncogenic = r.Oncogenic
		setTherapeuticLevels(e, r.Treatments)
		e.OncokbTxCitations = getCitations[Treatments](r.Treatments)
		e.OncokbHighestLevel = getHighestTherapeuticLevel(r.Treatments)
		e.OncokbHighestSensitivityLevel = r.HighestSensitiveLevel
		e.OncokbHighestResistanceLevel = r.HighestResistanceLevel
		setDiagnosticLevels(e, r.DiagnosticImplications)
		e.OncokbDxCitations = getCitations[DiagnosticImplications](r.DiagnosticImplications)
		e.OncokbHighestDxLevel = r.HighestDiagnosticImplicationLevel
		setPrognosticLevels(e, r.PrognosticImplications)
		e.OncokbPxCitations = getCitations[PrognosticImplications](r.PrognosticImplications)
		e.OncokbHighestPxLevel = r.HighestPrognosticImplicationLevel
	}
}

func setTherapeuticLevels(e *tt.Event, treatments []Treatments) {
	// therapeutic levels [1,2,3A,3B,4,R1,R2]
	for _, t := range treatments {
		switch t.Level {
		case "LEVEL_R1":
			e.OncokbLevelR1 = getDrugs(e.OncokbLevelR1, t.Drugs)
		case "LEVEL_1":
			e.OncokbLevel1 = getDrugs(e.OncokbLevel1, t.Drugs)
		case "LEVEL_2":
			e.OncokbLevel2 = getDrugs(e.OncokbLevel2, t.Drugs)
		case "LEVEL_3A":
			e.OncokbLevel3A = getDrugs(e.OncokbLevel3A, t.Drugs)
		case "LEVEL_3B":
			e.OncokbLevel3B = getDrugs(e.OncokbLevel3B, t.Drugs)
		case "LEVEL_R2":
			e.OncokbLevelR2 = getDrugs(e.OncokbLevelR2, t.Drugs)
		case "LEVEL_4":
			e.OncokbLevel4 = getDrugs(e.OncokbLevel4, t.Drugs)
		}
	}
}

func getDrugs(existingDrugList string, newDrugs []Drugs) string {
	var sb strings.Builder
	for _, d := range newDrugs {
		sb.WriteString(d.DrugName)
		sb.WriteString("+")
	}
	newDrugList := strings.TrimSuffix(sb.String(), "+")
	if len(existingDrugList) > 0 {
		if len(newDrugList) > 0 {
			return existingDrugList + "," + newDrugList
		}
		return existingDrugList
	}
	return newDrugList
}

func getCitations[T any](items []T) string {
	existingPmids := make(map[string]bool)
	existingAbstracts := make(map[string]bool)
	var sb strings.Builder
	for _, item := range items {
		var pmids, abstracts string
		// this is ugly!
		if c, ok := any(item).(Citations); ok {
			pmids = getPmids(existingPmids, c.Pmids)
			abstracts = getAbstracts(existingAbstracts, c.Abstracts)
		} else if t, ok := any(item).(Treatments); ok {
			pmids = getPmids(existingPmids, t.Pmids)
			abstracts = getAbstracts(existingAbstracts, t.Abstracts)
		} else if d, ok := any(item).(DiagnosticImplications); ok {
			pmids = getPmids(existingPmids, d.Pmids)
			abstracts = getAbstracts(existingAbstracts, d.Abstracts)
		} else if p, ok := any(item).(PrognosticImplications); ok {
			pmids = getPmids(existingPmids, p.Pmids)
			abstracts = getAbstracts(existingAbstracts, p.Abstracts)
		}
		if sb.Len() > 0 && len(pmids) > 0 {
			sb.WriteString(fmt.Sprintf(";%s", pmids))
		} else if len(pmids) > 0 {
			sb.WriteString(pmids)
		}
		if sb.Len() > 0 && len(abstracts) > 0 {
			sb.WriteString(fmt.Sprintf(";%s", abstracts))
		} else if len(abstracts) > 0 {
			sb.WriteString(abstracts)
		}
	}
	return sb.String()
}

func getPmids(existingPmids map[string]bool, pmids []string) string {
	var sb strings.Builder
	if len(pmids) > 0 {
		for _, pmid := range pmids {
			if _, exist := existingPmids[pmid]; !exist {
				sb.WriteString(pmid)
				sb.WriteString(";")
				existingPmids[pmid] = true
			}
		}
	}
	return strings.TrimSuffix(sb.String(), ";")
}

func getAbstracts(existingAbstracts map[string]bool, abstracts []Abstracts) string {
	var sb strings.Builder
	if len(abstracts) > 0 {
		for _, a := range abstracts {
			abstract := fmt.Sprintf("%s(%s)", a.Abstract, a.Link)
			if _, exist := existingAbstracts[abstract]; !exist {
				sb.WriteString(fmt.Sprintf("%s;", abstract))
				existingAbstracts[abstract] = true
			}
		}
	}
	return strings.TrimSuffix(sb.String(), ";")
}

var therapeuticLevels = [7]string{"LEVEL_R1", "LEVEL_1", "LEVEL_2", "LEVEL_3A", "LEVEL_3B", "LEVEL_4", "LEVEL_R2"}

func getHighestTherapeuticLevel(treatments []Treatments) string {
	for _, l := range therapeuticLevels {
		// the first therapeutic level we find in the treatment levels is the highest level
		for _, t := range treatments {
			if t.Level == l {
				return l
			}
		}
	}
	return ""
}

func setDiagnosticLevels(e *tt.Event, diagnosticImplications []DiagnosticImplications) {
	// diagnostic levels [Dx1, Dx2, Dx3]
	var sbDx1, sbDx2, sbDx3 strings.Builder
	for _, d := range diagnosticImplications {
		tumorType := d.TumorType.Code
		if len(tumorType) == 0 {
			tumorType = d.TumorType.MainType.Name
		}
		switch d.LevelOfEvidence {
		case "LEVEL_Dx1":
			sbDx1.WriteString(fmt.Sprintf("%s,", tumorType))
		case "LEVEL_Dx2":
			sbDx2.WriteString(fmt.Sprintf("%s,", tumorType))
		case "LEVEL_Dx3":
			sbDx3.WriteString(fmt.Sprintf("%s,", tumorType))
		}
	}
	e.OncokbLevelDx1 = strings.TrimSuffix(sbDx1.String(), ",")
	e.OncokbLevelDx2 = strings.TrimSuffix(sbDx2.String(), ",")
	e.OncokbLevelDx3 = strings.TrimSuffix(sbDx3.String(), ",")
}

func setPrognosticLevels(e *tt.Event, prognosticImplications []PrognosticImplications) {
	// prognostic levels [Px1, Px2, Px3]
	var sbPx1, sbPx2, sbPx3 strings.Builder
	for _, p := range prognosticImplications {
		tumorType := p.TumorType.Code
		if len(tumorType) == 0 {
			tumorType = p.TumorType.MainType.Name
		}
		switch p.LevelOfEvidence {
		case "LEVEL_Px1":
			sbPx1.WriteString(fmt.Sprintf("%s,", tumorType))
		case "LEVEL_Px2":
			sbPx2.WriteString(fmt.Sprintf("%s,", tumorType))
		case "LEVEL_Px3":
			sbPx3.WriteString(fmt.Sprintf("%s,", tumorType))
		}
	}
	e.OncokbLevelPx1 = strings.TrimSuffix(sbPx1.String(), ",")
	e.OncokbLevelPx2 = strings.TrimSuffix(sbPx2.String(), ",")
	e.OncokbLevelPx3 = strings.TrimSuffix(sbPx3.String(), ",")
}

func unMarshal[T any](msgData string) (T, error) {
	var target T
	if err := json.Unmarshal([]byte(msgData), &target); err != nil {
		return target, err
	}
	return target, nil
}
