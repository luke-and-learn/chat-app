type Message struct {
	Username	string	`json:"username"`
	Text		string	`json:"text"`
	Time		string	`json:"time"`
	Type		string	`json:"type"`
	UserCount	int		`json:"user_count"`
}
