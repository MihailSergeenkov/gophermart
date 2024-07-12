package common

type ContextValueKey int

const KeyUserID ContextValueKey = iota

var EncRespErrStr = "error encoding response"
var ReadReqErrStr = "failed to read request body"
var ContentTypeHeader = "Content-Type"
var JSONContentType = "application/json"
