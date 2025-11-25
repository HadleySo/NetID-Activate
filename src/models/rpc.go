package models

type BatchResponse struct {
	Count    int            `json:"count"`
	Messages []Message      `json:"messages"`
	Results  []ResultHolder `json:"results"`
}

type Message struct {
	Code    int               `json:"code"`
	Data    map[string]string `json:"data"`
	Message string            `json:"message"`
	Name    string            `json:"name"`
	Type    string            `json:"type"`
}

type ResultHolder struct {
	Error   any         `json:"error"`
	Result  GroupResult `json:"result"`
	Summary any         `json:"summary"`
	Value   string      `json:"value"`
}

type GroupResult struct {
	CN                []string `json:"cn"`
	Description       []string `json:"description"`
	DN                string   `json:"dn"`
	GIDNumber         []string `json:"gidnumber"`
	MemberGroup       []string `json:"member_group,omitempty"`
	MemberManagerUser []string `json:"membermanager_user,omitempty"`
}
