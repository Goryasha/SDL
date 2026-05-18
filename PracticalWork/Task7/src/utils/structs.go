package utils

type FilterPair struct {
    Key   string `json:"key" binding:"required"`
    Value string `json:"value" binding:"required"`
}

type RequestSelectBody struct {
    Scenario int          `json:"scenario" binding:"required"`
    Filters  []FilterPair `json:"filters"`
}

type ChangesPair struct {
    Key   string `json:"key" binding:"required"`
    Value string `json:"value" binding:"required"`
}

type RequestUpdateBody struct {
    Scenario int          `json:"scenario" binding:"required"`
    Ids  []int            `json:"ids"`
    Changes  []ChangesPair `json:"changes"`
}

type InputMainPair struct {
    Key   string `json:"key" binding:"required"`
    Value string `json:"value" binding:"required"`
}

type InputConnectedPair struct {
    Key   string `json:"key" binding:"required"`
    Value string `json:"value" binding:"required"`
}

type RequestInputOneBody struct {
    Scenario int                         `json:"scenario" binding:"required"`
    InputMain  []InputMainPair           `json:"mainRecordData" binding:"required"`
    InputConnected  []InputConnectedPair `json:"connectionData"`
    TableName    string                  `json:"connectedTable"`
    ConnectedField    string             `json:"ConnectedField"`
}

type RequestInputManyBody struct {
    Scenario int                         `json:"scenario" binding:"required"`
    InputMain  [][]InputMainPair           `json:"mainRecordData" binding:"required"`
    InputConnected  [][]InputConnectedPair `json:"connectionData" binding:"required"`
    TableName    string                  `json:"connectedTable"`
    ConnectedField    string             `json:"ConnectedField"`
}

type TableViewModel struct {
    TableName string
    Headers   []string
    Rows      [][]string
}

type TableInfo struct {
    TableName    string
    Columns      []string
    NotNullColumns      []string
    ConnectedField string
}

type Result struct {
    Tables []TableInfo
}