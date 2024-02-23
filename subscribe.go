package deepcoin

type SendTopicAction struct {
	Action      string `json:"action"`
	FilterValue string `json:"filterValue"`
	LocalNo     int    `json:"localNo"`
	ResumeNo    int    `json:"resumeNo"`
	TopicID     string `json:"topicID"`
}
