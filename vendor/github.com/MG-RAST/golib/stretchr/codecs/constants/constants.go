package constants

/*
	Supported Content Types
*/

const (
	ContentTypeJSON      string = "application/json"
	FileExtensionJSON    string = ".json"
	ContentTypeJSONP     string = "text/javascript"
	FileExtensionJSONP   string = ".js"
	ContentTypeBSON      string = "application/bson"
	FileExtensionBSON    string = ".bson"
	ContentTypeMsgpack   string = "application/x-msgpack"
	FileExtensionMsgpack string = ".msgpack"
	ContentTypeCSV       string = "text/csv"
	FileExtensionCSV     string = ".csv"
	ContentTypeXML       string = "text/xml"
	FileExtensionXML     string = ".xml"
)

const (
	OptionKeyClientCallback string = "options.client.callback"
	OptionKeyClientContext  string = "options.client.context"
)
