package main

const (
	linkToExplorer      = "https://ton.live/accounts/accountDetails?id="
	linkToSubmissionGov = "https://gov.freeton.org/submission?proposalAddress=%s&submissionId="
	gas                 = 0.04
)

type TomlConfig struct {
	Rewards []int64 `toml:"rewardsPlace"`
}
type mainDats struct {
	TitleContext   string
	LinkToContext  string
	Contenders     []Contenders
	Jurys          []Jury
	RewardsSumCont int64
}
type Contenders struct {
	IDS          int64
	Address      string
	AverageScore float64
	GovermentD   *goverment
	Reject       int64
	Ranking      int64
	Reward       int64
}

type Jury struct {
	Address   string
	PublicKey string
	Reward    float64
}

type goverment struct {
	CommentsAbstained []string `json:"commentsAbstained"`
	CommentsAgainst   []string `json:"commentsAgainst"`
	CommentsFor       []string `json:"commentsFor"`
	JurorsAbstained   []string `json:"jurorsAbstained"`
	JurorsAgainst     []string `json:"jurorsAgainst"`
	JurorsFor         []string `json:"jurorsFor"`
	Marks             []string `json:"marks"`
}

type votes struct {
	JuryFor       int64
	JuryAbstained int64
	JuryAgainst   int64
}

//temp structs
type resultContenders struct {
	Addresses []string `json:"addresses"`
	Ids       []string `json:"ids"`
}

type resContestInfo struct {
	JuryAddresses []string `json:"juryAddresses"`
	JuryKeys      []string `json:"juryKeys"`
	Link          string   `json:"link"`
	Title         string   `json:"title"`
	Hash          string   `json:"hash"`
}

type req struct {
	ID int64 `json:"id"`
}

type T struct{}
