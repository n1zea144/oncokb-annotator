package tempo_databricks_gateway

type OncoKBMutationRequest struct {
	Alteration      string   `json:"alteration"`
	Consequence     string   `json:"consequence"`
	EvidenceTypes   []string `json:"evidenceTypes,omitempty"`
	Gene            Gene     `json:"gene"`
	ID              string   `json:"id"`
	ProteinEnd      int      `json:"proteinEnd"`
	ProteinStart    int      `json:"proteinStart"`
	ReferenceGenome string   `json:"referenceGenome"`
	TumorType       string   `json:"tumorType"`
}

type Gene struct {
	EntrezGeneID int    `json:"entrezGeneId"`
	HugoSymbol   string `json:"hugoSymbol"`
}

type OncoKBErrorResponse struct {
	Detail  string `json:"detail"`
	Message string `json:"message"`
	Path    string `json:"path"`
	Status  int    `json:"status"`
	Title   string `json:"title"`
}

type OncoKBResponse struct {
	AlleleExist                       bool                     `json:"alleleExist"`
	DataVersion                       string                   `json:"dataVersion"`
	DiagnosticImplications            []DiagnosticImplications `json:"diagnosticImplications"`
	DiagnosticSummary                 string                   `json:"diagnosticSummary"`
	GeneExist                         bool                     `json:"geneExist"`
	GeneSummary                       string                   `json:"geneSummary"`
	HighestDiagnosticImplicationLevel string                   `json:"highestDiagnosticImplicationLevel"`
	HighestFdaLevel                   string                   `json:"highestFdaLevel"`
	HighestPrognosticImplicationLevel string                   `json:"highestPrognosticImplicationLevel"`
	HighestResistanceLevel            string                   `json:"highestResistanceLevel"`
	HighestSensitiveLevel             string                   `json:"highestSensitiveLevel"`
	Hotspot                           bool                     `json:"hotspot"`
	LastUpdate                        string                   `json:"lastUpdate"`
	MutationEffect                    MutationEffect           `json:"mutationEffect"`
	Oncogenic                         string                   `json:"oncogenic"`
	OtherSignificantResistanceLevels  []string                 `json:"otherSignificantResistanceLevels"`
	OtherSignificantSensitiveLevels   []string                 `json:"otherSignificantSensitiveLevels"`
	PrognosticImplications            []PrognosticImplications `json:"prognosticImplications"`
	PrognosticSummary                 string                   `json:"prognosticSummary"`
	Query                             Query                    `json:"query"`
	Treatments                        []Treatments             `json:"treatments"`
	TumorTypeSummary                  string                   `json:"tumorTypeSummary"`
	VariantExist                      bool                     `json:"variantExist"`
	VariantSummary                    string                   `json:"variantSummary"`
	Vus                               bool                     `json:"vus"`
}

type Abstracts struct {
	Abstract string `json:"abstract"`
	Link     string `json:"link"`
}

type Children struct {
	AdditionalProp1 string `json:"additionalProp1"`
	AdditionalProp2 string `json:"additionalProp2"`
	AdditionalProp3 string `json:"additionalProp3"`
}

type MainType struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	TumorForm string `json:"tumorForm"`
}

type TumorType struct {
	Children  Children `json:"children"`
	Code      string   `json:"code"`
	Color     string   `json:"color"`
	ID        int      `json:"id"`
	Level     int      `json:"level"`
	MainType  MainType `json:"mainType"`
	Name      string   `json:"name"`
	Parent    string   `json:"parent"`
	Tissue    string   `json:"tissue"`
	TumorForm string   `json:"tumorForm"`
}

type DiagnosticImplications struct {
	Abstracts       []Abstracts `json:"abstracts"`
	Alterations     []string    `json:"alterations"`
	Description     string      `json:"description"`
	LevelOfEvidence string      `json:"levelOfEvidence"`
	Pmids           []string    `json:"pmids"`
	TumorType       TumorType   `json:"tumorType"`
}

type Citations struct {
	Abstracts []Abstracts `json:"abstracts"`
	Pmids     []string    `json:"pmids"`
}

type MutationEffect struct {
	Citations   Citations `json:"citations"`
	Description string    `json:"description"`
	KnownEffect string    `json:"knownEffect"`
}

type PrognosticImplications struct {
	Abstracts       []Abstracts `json:"abstracts"`
	Alterations     []string    `json:"alterations"`
	Description     string      `json:"description"`
	LevelOfEvidence string      `json:"levelOfEvidence"`
	Pmids           []string    `json:"pmids"`
	TumorType       TumorType   `json:"tumorType"`
}

type Query struct {
	Alteration          string `json:"alteration"`
	AlterationType      string `json:"alterationType"`
	CanonicalTranscript string `json:"canonicalTranscript"`
	Consequence         string `json:"consequence"`
	EntrezGeneID        int    `json:"entrezGeneId"`
	Hgvs                string `json:"hgvs"`
	HgvsInfo            string `json:"hgvsInfo"`
	HugoSymbol          string `json:"hugoSymbol"`
	ID                  string `json:"id"`
	ProteinEnd          int    `json:"proteinEnd"`
	ProteinStart        int    `json:"proteinStart"`
	ReferenceGenome     string `json:"referenceGenome"`
	SvType              string `json:"svType"`
	TumorType           string `json:"tumorType"`
}

type Drugs struct {
	DrugName string `json:"drugName"`
	NcitCode string `json:"ncitCode"`
}

type LevelAssociatedCancerType struct {
	Children  Children `json:"children"`
	Code      string   `json:"code"`
	Color     string   `json:"color"`
	ID        int      `json:"id"`
	Level     int      `json:"level"`
	MainType  MainType `json:"mainType"`
	Name      string   `json:"name"`
	Parent    string   `json:"parent"`
	Tissue    string   `json:"tissue"`
	TumorForm string   `json:"tumorForm"`
}

type LevelExcludedCancerTypes struct {
	Children  Children `json:"children"`
	Code      string   `json:"code"`
	Color     string   `json:"color"`
	ID        int      `json:"id"`
	Level     int      `json:"level"`
	MainType  MainType `json:"mainType"`
	Name      string   `json:"name"`
	Parent    string   `json:"parent"`
	Tissue    string   `json:"tissue"`
	TumorForm string   `json:"tumorForm"`
}

type Treatments struct {
	Abstracts                 []Abstracts                `json:"abstracts"`
	Alterations               []string                   `json:"alterations"`
	ApprovedIndications       []string                   `json:"approvedIndications"`
	Description               string                     `json:"description"`
	Drugs                     []Drugs                    `json:"drugs"`
	FdaLevel                  string                     `json:"fdaLevel"`
	Level                     string                     `json:"level"`
	LevelAssociatedCancerType LevelAssociatedCancerType  `json:"levelAssociatedCancerType"`
	LevelExcludedCancerTypes  []LevelExcludedCancerTypes `json:"levelExcludedCancerTypes"`
	Pmids                     []string                   `json:"pmids"`
}
