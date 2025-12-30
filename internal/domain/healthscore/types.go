package healthscore

type HealthScoreRange struct {
	Min    int
	Max    int
	Label  string
	Color  string
	Status string
}

var HealthScoreRanges = []HealthScoreRange{
	{Min: 90, Max: 100, Label: "Excelente", Color: "#4CAF50", Status: "excellent"},
	{Min: 70, Max: 89, Label: "Bom", Color: "#8BC34A", Status: "good"},
	{Min: 50, Max: 69, Label: "Regular", Color: "#FF9800", Status: "regular"},
	{Min: 30, Max: 49, Label: "Atenção", Color: "#FF5722", Status: "attention"},
	{Min: 0, Max: 29, Label: "Crítico", Color: "#F44336", Status: "critical"},
}

func GetHealthScoreRange(score int) HealthScoreRange {
	for _, r := range HealthScoreRanges {
		if score >= r.Min && score <= r.Max {
			return r
		}
	}
	return HealthScoreRanges[len(HealthScoreRanges)-1]
}
