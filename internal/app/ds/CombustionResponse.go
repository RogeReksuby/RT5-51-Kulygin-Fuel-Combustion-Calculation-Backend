package ds

type CombustionResponse struct {
	ID             uint    `json:"id"`
	Status         string  `json:"status"`
	DateCreate     string  `json:"date_create"`
	DateUpdate     string  `json:"date_update"`
	DateFinish     string  `json:"date_finish,omitempty"`
	CreatorLogin   string  `json:"creator_login"`
	ModeratorLogin string  `json:"moderator_login,omitempty"`
	MolarVolume    float64 `json:"molar_volume"`
	FinalResult    float64 `json:"final_result"`
}
