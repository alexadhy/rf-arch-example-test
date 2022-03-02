// TODO: create a formatter to the stdout / console without encoding it with JSON
package main

//import (
//	"bytes"
//	"fmt"
//	"github.com/newrelic/go-agent/v3/integrations/logcontext"
//	"github.com/newrelic/go-agent/v3/newrelic"
//	log "github.com/sirupsen/logrus"
//)
//
//type logFields map[string]interface{}
//
//// ContextFormatter
//type contextFormatter struct{}
//
//// Format renders a single log entry.
//func (f contextFormatter) Format(e *log.Entry) ([]byte, error) {
//	// 12 = 6 from GetLinkingMetadata + 6 more below
//	data := make(logFields, len(e.Data)+12)
//	for k, v := range e.Data {
//		data[k] = v
//	}
//
//	if ctx := e.Context; nil != ctx {
//		if txn := newrelic.FromContext(ctx); nil != txn {
//			logcontext.AddLinkingMetadata(data, txn.GetLinkingMetadata())
//		}
//	}
//
//	data[logcontext.KeyTimestamp] = uint64(e.Time.UnixNano()) / uint64(1000*1000)
//	data[logcontext.KeyMessage] = e.Message
//	data[logcontext.KeyLevel] = e.Level
//
//	if e.HasCaller() {
//		data[logcontext.KeyFile] = e.Caller.File
//		data[logcontext.KeyLine] = e.Caller.Line
//		data[logcontext.KeyMethod] = e.Caller.Function
//	}
//	var b *bytes.Buffer
//	if e.Buffer != nil {
//		b = e.Buffer
//	} else {
//		b = &bytes.Buffer{}
//	}
//	f.writeData(b, data)
//	return b.Bytes(), nil
//}
//
//func (f *contextFormatter) writeData(buf *bytes.Buffer, data logFields) {
//	var needsComma bool
//	buf.WriteByte('{')
//	for k, v := range data {
//		if needsComma {
//			buf.WriteByte(',')
//		} else {
//			needsComma = true
//		}
//		buf.WriteString(k)
//		buf.WriteByte(':')
//		buf.WriteByte(' ')
//		f.appendValue(buf, v)
//	}
//	buf.WriteByte('}')
//	buf.WriteByte('\n')
//}
//
//func needsQuoting(text string) bool {
//	for _, ch := range text {
//		if !((ch >= 'a' && ch <= 'z') ||
//			(ch >= 'A' && ch <= 'Z') ||
//			(ch >= '0' && ch <= '9') ||
//			ch == '-' || ch == '.' || ch == '_' || ch == '/' || ch == '@' || ch == '^' || ch == '+') {
//			return true
//		}
//	}
//	return false
//}
//
//func (f *contextFormatter) appendValue(buffer *bytes.Buffer, value interface{}) (n int) {
//	switch value.(type) {
//	case string:
//		//buffer.WriteByte('"')
//		stringVal := fmt.Sprintf(`%s`, value)
//		n = writeString(buffer, stringVal)
//		//buffer.WriteByte('"')
//	default:
//		stringVal := fmt.Sprint(value)
//		n = writeString(buffer, stringVal)
//	}
//
//	return
//}
//
//func writeString(buffer *bytes.Buffer, stringVal string) (n int) {
//	if !needsQuoting(stringVal) {
//		n, _ = buffer.WriteString(stringVal)
//	} else {
//		n, _ = buffer.WriteString(fmt.Sprintf("%q", stringVal))
//	}
//	return
//}
