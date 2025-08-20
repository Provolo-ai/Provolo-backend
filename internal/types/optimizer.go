package types

// Portfolio Optimizer Response expected from Gemini
type OptimizerResponse struct {
	FullAnalysis              string `json:"fullAnalysis" validate:"required,min=1"`
	WeaknessesAndOptimization string `json:"weaknessesAndOptimization" validate:"required,min=1"`
	OptimizedProfileOverview  string `json:"optimizedProfileOverview" validate:"required,min=1"`
	SuggestedProjectTitles    string `json:"suggestedProjectTitles" validate:"required,min=1"`
	RecommendedVisuals        string `json:"recommendedVisuals" validate:"required,min=1"`
	BeforeAfterComparison     string `json:"beforeAfterComparison" validate:"required,min=1"`
}
