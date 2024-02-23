package deepcoin

type SendTopicAction struct {
	Action      string `json:"action"`
	FilterValue string `json:"filterValue"`
	LocalNo     int64  `json:"localNo"`
	ResumeNo    int64  `json:"resumeNo"`
	TopicID     string `json:"topicID"`
}
