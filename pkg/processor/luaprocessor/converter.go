package luaprocessor

import (
	"encoding/hex"
	"fmt"
	"reflect"

	"go.opentelemetry.io/collector/model/pdata"
)

// MetricsLuaConverter converts between lua and go
type LogsLuaConverter struct {
	logs *pdata.Logs
}

// NewLogsLuaConverter creates a new LogsLuaConverter.
func NewLogsLuaConverter(logs pdata.Logs) LogsLuaConverter {
	return LogsLuaConverter{logs: &logs}
}

// ConvertToLua converts Metrics data object to lua map representation
func (cv LogsLuaConverter) ConvertToLua() map[string]interface{} {
	resourceLogs := []interface{}{}
	logs := cv.logs
	rms := logs.ResourceLogs()
	for i := 0; i < rms.Len(); i++ {
		resourceLogs = append(resourceLogs, convertResourceLogsToLua(rms.At(i)))
	}

	return map[string]interface{}{
		"logRecordCount": logs.LogRecordCount(),
		"resourceLogs":   resourceLogs,
	}
}

// ConvertFromLua converts lua representation to Logs
func (cv LogsLuaConverter) ConvertFromLua(luaRs interface{}) (pdata.Logs, error) {
	luaRsMap := luaRs.(map[string]interface{})
	rs := pdata.NewLogs()

	luaResourceLogs := luaRsMap["resourceLogs"].([]interface{})
	resourceLogsSlice := rs.ResourceLogs()
	resourceLogsSlice.EnsureCapacity(len(luaResourceLogs))
	for _, luaResourceLog := range luaResourceLogs {
		convertResourceLogsFromLua(resourceLogsSlice.AppendEmpty(), luaResourceLog.(map[string]interface{}))
	}

	return rs, nil
}

// convertResourceLogsToLua converts ResourceLogs to map representation
func convertResourceLogsToLua(rl pdata.ResourceLogs) map[string]interface{} {
	libraryLogs := []interface{}{}
	ilms := rl.InstrumentationLibraryLogs()
	for j := 0; j < ilms.Len(); j++ {
		libraryLogs = append(libraryLogs, convertLibraryLogsToLua(ilms.At(j)))
	}

	return map[string]interface{}{
		"schemaUrl":   rl.SchemaUrl(),
		"resource":    convertResourceToLua(rl.Resource()),
		"libraryLogs": libraryLogs,
	}
}

// convertResourceLogsFromLua converts from lua and fills ResourceLogs
func convertResourceLogsFromLua(rl pdata.ResourceLogs, lua map[string]interface{}) {
	if schemaUrl, ok := lua["schemaUrl"]; ok {
		rl.SetSchemaUrl(schemaUrl.(string))
	}

	convertResourceFromLua(rl.Resource(), extractMapValueFrom(lua, "resource"))

	ll := extractSliceValueFrom(lua, "libraryLogs")
	libSlice := rl.InstrumentationLibraryLogs()
	libSlice.EnsureCapacity(len(ll))
	for _, luaLl := range ll {
		withMap(luaLl, func(luaLlMap map[string]interface{}) {
			convertLibraryLogsFromLua(libSlice.AppendEmpty(), luaLlMap)
		})
	}
}

// convertLibraryLogsToLua converts InstrumentationLibraryLogs to map representation
func convertLibraryLogsToLua(llogs pdata.InstrumentationLibraryLogs) map[string]interface{} {
	luaLogs := []interface{}{}
	lsl := llogs.Logs()
	for j := 0; j < lsl.Len(); j++ {
		luaLogs = append(luaLogs, convertLogToLua(lsl.At(j)))
	}

	return map[string]interface{}{
		"schemaUrl": llogs.SchemaUrl(),
		"library":   convertInstrumentationLibraryToLua(llogs.InstrumentationLibrary()),
		"logs":      luaLogs,
	}
}

// convertLibraryLogsFromLua converts from lua and fills InstrumentationLibraryLogs
func convertLibraryLogsFromLua(ll pdata.InstrumentationLibraryLogs, lua map[string]interface{}) {
	if schemaUrl, ok := lua["schemaUrl"]; ok {
		ll.SetSchemaUrl(schemaUrl.(string))
	}

	convertInstrumentationLibraryFromLua(ll.InstrumentationLibrary(), extractMapValueFrom(lua, "library"))

	logs := extractSliceValueFrom(lua, "logs")
	lSlice := ll.Logs()
	lSlice.EnsureCapacity(len(logs))
	for _, log := range logs {
		withMap(log, func(logMap map[string]interface{}) {
			convertLogFromLua(lSlice.AppendEmpty(), logMap)
		})
	}
}

// convertLogToLua converts LogRecord to map representation
func convertLogToLua(lr pdata.LogRecord) map[string]interface{} {
	return map[string]interface{}{
		"timestamp":              uint64(lr.Timestamp()),
		"traceID":                lr.TraceID().HexString(),
		"spanID":                 lr.SpanID().HexString(),
		"flags":                  uint64(lr.Flags()),
		"severityText":           lr.SeverityText(),
		"severityNumber":         int32(lr.SeverityNumber()),
		"name":                   lr.Name(),
		"body":                   convertAttributeValueToLua(lr.Body()),
		"attributes":             convertAttributeMapToLua(lr.Attributes()),
		"droppedAttributesCount": lr.DroppedAttributesCount(),
	}
}

// convertLogFromLua converts from lua and fills LogRecord
func convertLogFromLua(lr pdata.LogRecord, lua map[string]interface{}) {
	if timestamp, ok := lua["timestamp"]; ok {
		lr.SetTimestamp(pdata.Timestamp(uint64(timestamp.(int64))))
	}

	if traceID, ok := lua["traceID"]; ok {
		if id, err := convertTraceIdFromLua(traceID.(string)); err == nil {
			lr.SetTraceID(id)
		}
	}
	if spanID, ok := lua["spanID"]; ok {
		if id, err := convertSpanIdFromLua(spanID.(string)); err == nil {
			lr.SetSpanID(id)
		}
	}

	if flags, ok := lua["flags"]; ok {
		lr.SetFlags(uint32(flags.(int64)))
	}

	if severityText, ok := lua["severityText"]; ok {
		lr.SetSeverityText(severityText.(string))
	}

	if severityNumber, ok := lua["severityNumber"]; ok {
		lr.SetSeverityNumber(pdata.SeverityNumber(int32(severityNumber.(int64))))
	}

	convertAttributeValueFromLua(extractMapValueFrom(lua, "body")).CopyTo(lr.Body())

	if name, ok := lua["name"]; ok {
		lr.SetName(name.(string))
	}

	convertAttributeMapFromLua(lr.Attributes(), extractMapValueFrom(lua, "attributes"))

	if droppedAttributesCount, ok := lua["droppedAttributesCount"]; ok {
		lr.SetDroppedAttributesCount(uint32(droppedAttributesCount.(int64)))
	}
}

// MetricsLuaConverter converts between lua and go
type MetricsLuaConverter struct {
	metrics *pdata.Metrics
}

// NewMetricsLuaConverter creates a new MetricsLuaConverter.
func NewMetricsLuaConverter(md pdata.Metrics) MetricsLuaConverter {
	return MetricsLuaConverter{metrics: &md}
}

// ConvertToLua converts Metrics data object to lua map representation
func (cv MetricsLuaConverter) ConvertToLua() map[string]interface{} {
	resourceMetrics := []interface{}{}
	md := cv.metrics
	rms := md.ResourceMetrics()
	for i := 0; i < rms.Len(); i++ {
		resourceMetrics = append(resourceMetrics, convertResourceMetricsToLua(rms.At(i)))
	}

	return map[string]interface{}{
		"metricCount":     md.MetricCount(),
		"dataPointCount":  md.DataPointCount(),
		"resourceMetrics": resourceMetrics,
	}
}

// ConvertFromLua converts back result from lua and updates Metrics datat object
func (cv MetricsLuaConverter) ConvertFromLua(luaRs interface{}) (pdata.Metrics, error) {
	luaRsMap := luaRs.(map[string]interface{})
	rs := pdata.NewMetrics()

	luaResourceMetrics := luaRsMap["resourceMetrics"].([]interface{})
	resourceMetricsSlice := rs.ResourceMetrics()
	resourceMetricsSlice.EnsureCapacity(len(luaResourceMetrics))
	for _, luaResourceMetric := range luaResourceMetrics {
		convertResourceMetricsFromLua(resourceMetricsSlice.AppendEmpty(), luaResourceMetric.(map[string]interface{}))
	}

	return rs, nil
}

// convertResourceMetricsToLua converts ResourceMetrics to map representation
func convertResourceMetricsToLua(rm pdata.ResourceMetrics) map[string]interface{} {
	libraryMetrics := []interface{}{}
	ilms := rm.InstrumentationLibraryMetrics()
	for j := 0; j < ilms.Len(); j++ {
		libraryMetrics = append(libraryMetrics, convertLibraryMetricsToLua(ilms.At(j)))
	}

	return map[string]interface{}{
		"schemaUrl":      rm.SchemaUrl(),
		"resource":       convertResourceToLua(rm.Resource()),
		"libraryMetrics": libraryMetrics,
	}
}

// convertResourceMetricsFromLua converts from lua and fills ResourceMetrics
func convertResourceMetricsFromLua(rm pdata.ResourceMetrics, luaRm map[string]interface{}) {
	if schemaUrl, ok := luaRm["schemaUrl"]; ok {
		rm.SetSchemaUrl(schemaUrl.(string))
	}

	convertResourceFromLua(rm.Resource(), extractMapValueFrom(luaRm, "resource"))

	lm := extractSliceValueFrom(luaRm, "libraryMetrics")
	libSlice := rm.InstrumentationLibraryMetrics()
	libSlice.EnsureCapacity(len(lm))
	for _, luaLm := range lm {
		withMap(luaLm, func(luaLmMap map[string]interface{}) {
			convertLibraryMetricsFromLua(libSlice.AppendEmpty(), luaLmMap)
		})
	}
}

// convertLibraryMetricsToLua converts InstrumentationLibraryMetrics to map representation
func convertLibraryMetricsToLua(rm pdata.InstrumentationLibraryMetrics) map[string]interface{} {
	metrics := []interface{}{}
	ms := rm.Metrics()
	for j := 0; j < ms.Len(); j++ {
		metrics = append(metrics, convertMetricToLua(ms.At(j)))
	}

	return map[string]interface{}{
		"schemaUrl": rm.SchemaUrl(),
		"library":   convertInstrumentationLibraryToLua(rm.InstrumentationLibrary()),
		"metrics":   metrics,
	}
}

// convertLibraryMetricsFromLua converts from lua and fills InstrumentationLibraryMetrics
func convertLibraryMetricsFromLua(rm pdata.InstrumentationLibraryMetrics, luaRm map[string]interface{}) {
	if schemaUrl, ok := luaRm["schemaUrl"]; ok {
		rm.SetSchemaUrl(schemaUrl.(string))
	}

	convertInstrumentationLibraryFromLua(rm.InstrumentationLibrary(), extractMapValueFrom(luaRm, "library"))

	metrics := extractSliceValueFrom(luaRm, "metrics")
	mSlice := rm.Metrics()
	mSlice.EnsureCapacity(len(metrics))
	for _, metric := range metrics {
		withMap(metric, func(metricMap map[string]interface{}) {
			convertMetricFromLua(mSlice.AppendEmpty(), metricMap)
		})
	}
}

// convertMetricToLua converts Metric to map representation
func convertMetricToLua(m pdata.Metric) map[string]interface{} {
	metricExp := map[string]interface{}{
		"name":        m.Name(),
		"description": m.Description(),
		"unit":        m.Unit(),
	}

	switch m.DataType() {
	case pdata.MetricDataTypeGauge:
		metricExp["gauge"] = convertMetricGaugeToLua(m.Gauge())
	case pdata.MetricDataTypeSum:
		metricExp["sum"] = convertMetricSumToLua(m.Sum())
	case pdata.MetricDataTypeHistogram:
		metricExp["histogram"] = convertMetricHistogramToLua(m.Histogram())
	case pdata.MetricDataTypeExponentialHistogram:
		metricExp["exponentialHistogram"] = convertMetricExponentialHistogramToLua(m.ExponentialHistogram())
	case pdata.MetricDataTypeSummary:
		metricExp["summary"] = convertMetricSummaryToLua(m.Summary())
	}

	return metricExp
}

// convertMetricFromLua converts from lua and fills Metric
func convertMetricFromLua(m pdata.Metric, lua map[string]interface{}) {
	if name, ok := lua["name"]; ok {
		m.SetName(name.(string))
	}

	if description, ok := lua["description"]; ok {
		m.SetDescription(description.(string))
	}

	if unit, ok := lua["unit"]; ok {
		m.SetUnit(unit.(string))
	}

	if _, ok := lua["gauge"]; ok {
		m.SetDataType(pdata.MetricDataTypeGauge)
		convertMetricGaugeFromLua(m.Gauge(), extractMapValueFrom(lua, "gauge"))
	} else if _, ok := lua["sum"]; ok {
		m.SetDataType(pdata.MetricDataTypeSum)
		convertMetricSumFromLua(m.Sum(), extractMapValueFrom(lua, "sum"))
	} else if _, ok := lua["histogram"]; ok {
		m.SetDataType(pdata.MetricDataTypeHistogram)
		convertMetricHistogramFromLua(m.Histogram(), extractMapValueFrom(lua, "histogram"))
	} else if _, ok := lua["exponentialHistogram"]; ok {
		m.SetDataType(pdata.MetricDataTypeExponentialHistogram)
		convertMetricExponentialHistogramFromLua(m.ExponentialHistogram(), extractMapValueFrom(lua, "exponentialHistogram"))
	} else if _, ok := lua["summary"]; ok {
		m.SetDataType(pdata.MetricDataTypeSummary)
		convertMetricSummaryFromLua(m.Summary(), extractMapValueFrom(lua, "summary"))
	}
}

// convertMetricGaugeToLua converts Gauge to map representation
func convertMetricGaugeToLua(m pdata.Gauge) map[string]interface{} {
	return map[string]interface{}{
		"dataPoints": convertMetricNumberDataPointSliceToLua(m.DataPoints()),
	}
}

// convertMetricGaugeFromLua converts from lua and fills Gauge
func convertMetricGaugeFromLua(m pdata.Gauge, lua map[string]interface{}) {
	convertMetricNumberDataPointSliceFromLua(m.DataPoints(), extractSliceValueFrom(lua, "dataPoints"))
}

// convertMetricNumberDataPointSliceToLua converts NumberDataPointSlice to map representation
func convertMetricNumberDataPointSliceToLua(dp pdata.NumberDataPointSlice) []interface{} {
	dpExp := []interface{}{}
	for j := 0; j < dp.Len(); j++ {
		dpExp = append(dpExp, convertMetricNumberDataPointToLua(dp.At(j)))
	}

	return dpExp
}

// convertMetricNumberDataPointSliceFromLua converts from lua and fills NumberDataPointSlice
func convertMetricNumberDataPointSliceFromLua(dp pdata.NumberDataPointSlice, lua []interface{}) {
	dp.EnsureCapacity(len(lua))
	for _, luaDp := range lua {
		withMap(luaDp, func(luaDpMap map[string]interface{}) {
			convertMetricNumberDataPointFromLua(dp.AppendEmpty(), luaDpMap)
		})
	}
}

// convertMetricNumberDataPointToLua converts NumberDataPoint to map representation
func convertMetricNumberDataPointToLua(dp pdata.NumberDataPoint) map[string]interface{} {
	return map[string]interface{}{
		"attributes":     convertAttributeMapToLua(dp.Attributes()),
		"startTimestamp": uint64(dp.StartTimestamp()),
		"timestamp":      uint64(dp.Timestamp()),
		"value":          dp.DoubleVal(),
		"exemplars":      convertMetricExemplarSliceToLua(dp.Exemplars()),
		"flags":          uint64(dp.Flags()),
	}
}

// convertMetricNumberDataPointFromLua converts NumberDataPoint to map representation
func convertMetricNumberDataPointFromLua(dp pdata.NumberDataPoint, lua map[string]interface{}) {

	if startTimestamp, ok := lua["startTimestamp"]; ok {
		dp.SetStartTimestamp(pdata.Timestamp(uint64(startTimestamp.(int64))))
	}

	if timestamp, ok := lua["timestamp"]; ok {
		dp.SetTimestamp(pdata.Timestamp(uint64(timestamp.(int64))))
	}

	if flags, ok := lua["flags"]; ok {
		dp.SetFlags(pdata.MetricDataPointFlags(uint64(flags.(int64))))
	}

	convertAttributeMapFromLua(dp.Attributes(), extractMapValueFrom(lua, "attributes"))
	convertMetricExemplarSliceFromLua(dp.Exemplars(), extractSliceValueFrom(lua, "exemplars"))

	if value, ok := lua["value"]; ok {
		switch vt := value.(type) {
		case int64:
			dp.SetIntVal(vt)
		case float64:
			dp.SetDoubleVal(vt)
		default:
			// TODO error handling
			fmt.Println("Unsupported Type!")
		}
	}

}

// convertMetricSumToLua converts Sum to map representation
func convertMetricSumToLua(m pdata.Sum) map[string]interface{} {
	return map[string]interface{}{
		"dataPoints":             convertMetricNumberDataPointSliceToLua(m.DataPoints()),
		"isMonotonic":            m.IsMonotonic(),
		"aggregationTemporality": int32(m.AggregationTemporality()),
	}
}

// convertMetricSumFromLua converts from lua and fills Sum
func convertMetricSumFromLua(m pdata.Sum, lua map[string]interface{}) {
	if isMonotonic, ok := lua["isMonotonic"]; ok {
		m.SetIsMonotonic(isMonotonic.(bool))
	}

	if aggregationTemporality, ok := lua["aggregationTemporality"]; ok {
		m.SetAggregationTemporality(pdata.MetricAggregationTemporality(int32(aggregationTemporality.(int64))))
	}

	convertMetricNumberDataPointSliceFromLua(m.DataPoints(), extractSliceValueFrom(lua, "dataPoints"))
}

// convertMetricHistogramToLua converts Histogram to map representation
func convertMetricHistogramToLua(m pdata.Histogram) map[string]interface{} {

	dp := m.DataPoints()
	dpExp := []interface{}{}
	for j := 0; j < dp.Len(); j++ {
		dpExp = append(dpExp, convertMetricHistogramDataPointToLua(dp.At(j)))
	}

	return map[string]interface{}{
		"dataPoints":             dpExp,
		"aggregationTemporality": int32(m.AggregationTemporality()),
	}
}

// convertMetricHistogramFromLua converts from lua and fills Histogram
func convertMetricHistogramFromLua(m pdata.Histogram, lua map[string]interface{}) {

	dpLua := extractSliceValueFrom(lua, "dataPoints")
	dp := m.DataPoints()
	dp.EnsureCapacity(len(dpLua))
	for _, lDp := range dpLua {
		withMap(lDp, func(lMap map[string]interface{}) {
			convertMetricHistogramDataPointFromLua(dp.AppendEmpty(), lMap)
		})
	}

	if aggregationTemporality, ok := lua["aggregationTemporality"]; ok {
		m.SetAggregationTemporality(pdata.MetricAggregationTemporality(int32(aggregationTemporality.(int64))))
	}
}

// convertMetricHistogramDataPointToLua converts HistogramDataPoint to map representation
func convertMetricHistogramDataPointToLua(dp pdata.HistogramDataPoint) map[string]interface{} {
	return map[string]interface{}{
		"attributes":     convertAttributeMapToLua(dp.Attributes()),
		"startTimestamp": uint64(dp.StartTimestamp()),
		"timestamp":      uint64(dp.Timestamp()),
		"count":          dp.Count(),
		"sum":            dp.Sum(),
		"bucketCounts":   dp.BucketCounts(),
		"explicitBounds": dp.ExplicitBounds(),
		"exemplars":      convertMetricExemplarSliceToLua(dp.Exemplars()),
		"flags":          uint64(dp.Flags()),
	}
}

// convertMetricHistogramDataPointFromLua converts from lua and fills HistogramDataPoint
func convertMetricHistogramDataPointFromLua(dp pdata.HistogramDataPoint, lua map[string]interface{}) {
	convertAttributeMapFromLua(dp.Attributes(), extractMapValueFrom(lua, "attributes"))

	if startTimestamp, ok := lua["startTimestamp"]; ok {
		dp.SetStartTimestamp(pdata.Timestamp(uint64(startTimestamp.(int64))))
	}

	if timestamp, ok := lua["timestamp"]; ok {
		dp.SetTimestamp(pdata.Timestamp(uint64(timestamp.(int64))))
	}

	if count, ok := lua["count"]; ok {
		dp.SetCount(uint64(count.(int64)))
	}

	if _, ok := lua["bucketCounts"]; ok {
		dp.SetBucketCounts(convertInterfaceToUint64Slice(extractSliceValueFrom(lua, "bucketCounts")))
	}

	if _, ok := lua["explicitBounds"]; ok {
		dp.SetExplicitBounds(convertInterfaceToFloat64Slice(extractSliceValueFrom(lua, "explicitBounds")))
	}

	convertMetricExemplarSliceFromLua(dp.Exemplars(), extractSliceValueFrom(lua, "exemplars"))

	if flags, ok := lua["flags"]; ok {
		dp.SetFlags(pdata.MetricDataPointFlags(uint64(flags.(int64))))
	}
}

func convertInterfaceToUint64Slice(in []interface{}) (out []uint64) {
	out = make([]uint64, len(in), len(in))
	for i := range in {
		out[i] = uint64(in[i].(int64))
	}
	return
}

func convertInterfaceToFloat64Slice(in []interface{}) (out []float64) {
	out = make([]float64, len(in), len(in))
	for i := range in {
		out[i] = in[i].(float64)
	}
	return
}

// convertMetricExemplarSliceToLua converts ExemplarSlice to slice representation
func convertMetricExemplarSliceToLua(ex pdata.ExemplarSlice) []interface{} {
	exExp := []interface{}{}
	for j := 0; j < ex.Len(); j++ {
		exExp = append(exExp, convertMetricExemplarToLua(ex.At(j)))
	}

	return exExp
}

// convertMetricExemplarSliceFromLua converts from lua and fills ExemplarSlice
func convertMetricExemplarSliceFromLua(ex pdata.ExemplarSlice, lua []interface{}) {
	ex.EnsureCapacity(len(lua))
	for _, luaEx := range lua {
		withMap(luaEx, func(luaExMap map[string]interface{}) {
			convertMetricExemplarFromLua(ex.AppendEmpty(), luaExMap)
		})
	}
}

// convertMetricExemplarToLua converts Exemplar to map representation
func convertMetricExemplarToLua(ex pdata.Exemplar) map[string]interface{} {
	return map[string]interface{}{
		"attributes": convertAttributeMapToLua(ex.FilteredAttributes()),
		"timestamp":  uint64(ex.Timestamp()),
		"value":      ex.DoubleVal(),
		"traceID":    ex.TraceID().HexString(),
		"spanID":     ex.SpanID().HexString(),
	}
}

// convertMetricExemplarFromLua converts from lua and fills Exemplar
func convertMetricExemplarFromLua(ex pdata.Exemplar, lua map[string]interface{}) {

	convertAttributeMapFromLua(ex.FilteredAttributes(), extractMapValueFrom(lua, "attributes"))

	if timestamp, ok := lua["timestamp"]; ok {
		ex.SetTimestamp(pdata.Timestamp(uint64(timestamp.(int64))))
	}

	if value, ok := lua["value"]; ok {
		switch vt := value.(type) {
		case int64:
			ex.SetIntVal(vt)
		case float64:
			ex.SetDoubleVal(vt)
		default:
			// TODO error handling
			fmt.Println("Unsupported Type!")
		}
	}

	if traceID, ok := lua["traceID"]; ok {
		if id, err := convertTraceIdFromLua(traceID.(string)); err == nil {
			ex.SetTraceID(id)
		}
	}
	if spanID, ok := lua["spanID"]; ok {
		if id, err := convertSpanIdFromLua(spanID.(string)); err == nil {
			ex.SetSpanID(id)
		}
	}
}

func convertTraceIdFromLua(traceID string) (pdata.TraceID, error) {
	if bytes, err := hex.DecodeString(traceID); err == nil {
		var bytesArr [16]byte
		copy(bytesArr[:], bytes[:16])
		return pdata.NewTraceID(bytesArr), nil
	} else {
		return pdata.InvalidTraceID(), err
	}
}

func convertSpanIdFromLua(spanID string) (pdata.SpanID, error) {
	if bytes, err := hex.DecodeString(spanID); err == nil {
		var bytesArr [8]byte
		copy(bytesArr[:], bytes[:8])
		return pdata.NewSpanID(bytesArr), nil
	} else {
		return pdata.InvalidSpanID(), err
	}
}

// convertMetricExponentialHistogramToLua converts ExponentialHistogram to map representation
func convertMetricExponentialHistogramToLua(m pdata.ExponentialHistogram) map[string]interface{} {
	dp := m.DataPoints()
	dpExp := []interface{}{}
	for j := 0; j < dp.Len(); j++ {
		dpExp = append(dpExp, convertMetricExponentialHistogramDataPointToLua(dp.At(j)))
	}

	return map[string]interface{}{
		"dataPoints":             dpExp,
		"aggregationTemporality": int32(m.AggregationTemporality()),
	}
}

// convertMetricExponentialHistogramFromLua converts from lua and fills ExponentialHistogram
func convertMetricExponentialHistogramFromLua(m pdata.ExponentialHistogram, lua map[string]interface{}) {

	dpLua := extractSliceValueFrom(lua, "dataPoints")
	dp := m.DataPoints()
	dp.EnsureCapacity(len(dpLua))
	for _, lDp := range dpLua {
		withMap(lDp, func(lMap map[string]interface{}) {
			convertMetricExponentialHistogramDataPointFromLua(dp.AppendEmpty(), lMap)
		})
	}

	if aggregationTemporality, ok := lua["aggregationTemporality"]; ok {
		m.SetAggregationTemporality(pdata.MetricAggregationTemporality(int32(aggregationTemporality.(int64))))
	}
}

// convertMetricExponentialHistogramDataPointToLua converts ExponentialHistogramDataPoint to map representation
func convertMetricExponentialHistogramDataPointToLua(dp pdata.ExponentialHistogramDataPoint) map[string]interface{} {

	ex := dp.Exemplars()
	exExp := []interface{}{}
	for j := 0; j < ex.Len(); j++ {
		exExp = append(exExp, convertMetricExemplarToLua(ex.At(j)))
	}

	return map[string]interface{}{
		"attributes":     convertAttributeMapToLua(dp.Attributes()),
		"startTimestamp": uint64(dp.StartTimestamp()),
		"timestamp":      uint64(dp.Timestamp()),
		"count":          dp.Count(),
		"sum":            dp.Sum(),
		"scale":          dp.Scale(),
		"zeroCount":      dp.ZeroCount(),
		"positive":       convertMetricBucketsToLua(dp.Positive()),
		"negative":       convertMetricBucketsToLua(dp.Negative()),
		"exemplars":      exExp,
		"flags":          uint32(dp.Flags()),
	}
}

// convertMetricExponentialHistogramDataPointFromLua converts from lua and fills ExponentialHistogramDataPoint
func convertMetricExponentialHistogramDataPointFromLua(dp pdata.ExponentialHistogramDataPoint, lua map[string]interface{}) {

	convertAttributeMapFromLua(dp.Attributes(), extractMapValueFrom(lua, "attributes"))

	if startTimestamp, ok := lua["startTimestamp"]; ok {
		dp.SetStartTimestamp(pdata.Timestamp(uint64(startTimestamp.(int64))))
	}

	if timestamp, ok := lua["timestamp"]; ok {
		dp.SetTimestamp(pdata.Timestamp(uint64(timestamp.(int64))))
	}

	if count, ok := lua["count"]; ok {
		dp.SetCount(uint64(count.(int64)))
	}

	if sum, ok := lua["sum"]; ok {
		dp.SetSum(sum.(float64))
	}

	if scale, ok := lua["scale"]; ok {
		dp.SetScale(int32(scale.(int64)))
	}

	if zeroCount, ok := lua["zeroCount"]; ok {
		dp.SetZeroCount(uint64(zeroCount.(int64)))
	}

	// TODO positive
	// TODO negative

	convertMetricExemplarSliceFromLua(dp.Exemplars(), extractSliceValueFrom(lua, "exemplars"))

	if flags, ok := lua["flags"]; ok {
		dp.SetFlags(pdata.MetricDataPointFlags(uint64(flags.(int64))))
	}
}

// convertMetricBucketsToLua converts Buckets to map representation
func convertMetricBucketsToLua(b pdata.Buckets) map[string]interface{} {
	return map[string]interface{}{
		"offset":       b.Offset(),
		"bucketCounts": b.BucketCounts(),
	}
}

// convertMetricSummaryToLua converts Summary to map representation
func convertMetricSummaryToLua(m pdata.Summary) map[string]interface{} {
	dp := m.DataPoints()
	dpExp := []interface{}{}
	for j := 0; j < dp.Len(); j++ {
		dpExp = append(dpExp, convertMetricSummaryDataPointToLua(dp.At(j)))
	}

	return map[string]interface{}{
		"dataPoints": dpExp,
	}
}

// convertMetricSummaryFromLua converts from lua and fills Summary
func convertMetricSummaryFromLua(m pdata.Summary, lua map[string]interface{}) {
	dpLua := extractSliceValueFrom(lua, "dataPoints")
	dp := m.DataPoints()
	dp.EnsureCapacity(len(dpLua))
	for _, lDp := range dpLua {
		withMap(lDp, func(lMap map[string]interface{}) {
			convertMetricSummaryDataPointFromLua(dp.AppendEmpty(), lMap)
		})
	}
}

// convertMetricSummaryDataPointToLua converts SummaryDataPoint to map representation
func convertMetricSummaryDataPointToLua(dp pdata.SummaryDataPoint) map[string]interface{} {
	qv := dp.QuantileValues()
	qvExp := []interface{}{}
	for j := 0; j < qv.Len(); j++ {
		qvExp = append(qvExp, convertMetricValueAtQuantileToLua(qv.At(j)))
	}

	return map[string]interface{}{
		"attributes":     convertAttributeMapToLua(dp.Attributes()),
		"startTimestamp": uint64(dp.StartTimestamp()),
		"timestamp":      uint64(dp.Timestamp()),
		"count":          dp.Count(),
		"sum":            dp.Sum(),
		"quantileValues": qvExp,
		"flags":          uint32(dp.Flags()),
	}
}

// convertMetricSummaryDataPointFromLua converts from lua and fills SummaryDataPoint
func convertMetricSummaryDataPointFromLua(m pdata.SummaryDataPoint, lua map[string]interface{}) {

	qLua := extractSliceValueFrom(lua, "quantileValues")
	qv := m.QuantileValues()
	qv.EnsureCapacity(len(qLua))
	for _, lq := range qLua {
		withMap(lq, func(lqMap map[string]interface{}) {
			convertMetricValueAtQuantileFromLua(qv.AppendEmpty(), lqMap)
		})
	}

	convertAttributeMapFromLua(m.Attributes(), extractMapValueFrom(lua, "attributes"))

	if startTimestamp, ok := lua["startTimestamp"]; ok {
		m.SetStartTimestamp(pdata.Timestamp(uint64(startTimestamp.(int64))))
	}

	if timestamp, ok := lua["timestamp"]; ok {
		m.SetTimestamp(pdata.Timestamp(uint64(timestamp.(int64))))
	}

	if count, ok := lua["count"]; ok {
		m.SetCount(uint64(count.(int64)))
	}

	if sum, ok := lua["sum"]; ok {
		m.SetSum(sum.(float64))
	}

	if flags, ok := lua["flags"]; ok {
		m.SetFlags(pdata.MetricDataPointFlags(uint64(flags.(int64))))
	}
}

// convertMetricValueAtQuantileToLua converts ValueAtQuantile to map representation
func convertMetricValueAtQuantileToLua(qv pdata.ValueAtQuantile) map[string]interface{} {
	return map[string]interface{}{
		"quantile": qv.Quantile(),
		"value":    qv.Value(),
	}
}

// convertMetricValueAtQuantileFromLua converts from lua and fills ValueAtQuantile
func convertMetricValueAtQuantileFromLua(qv pdata.ValueAtQuantile, lua map[string]interface{}) {
	if quantile, ok := lua["quantile"]; ok {
		qv.SetQuantile(quantile.(float64))
	}

	if value, ok := lua["value"]; ok {
		qv.SetValue(value.(float64))
	}
}

// convertInstrumentationLibraryToLua converts InstrumentationLibrary to map representation
func convertInstrumentationLibraryToLua(il pdata.InstrumentationLibrary) map[string]interface{} {
	return map[string]interface{}{
		"name":    il.Name(),
		"version": il.Version(),
	}
}

// convertInstrumentationLibraryFromLua converts from lua and fills InstrumentationLibrary
func convertInstrumentationLibraryFromLua(il pdata.InstrumentationLibrary, lua map[string]interface{}) {
	if name, ok := lua["name"]; ok {
		il.SetName(name.(string))
	}
	if version, ok := lua["version"]; ok {
		il.SetVersion(version.(string))
	}
}

// convertResourceToLua converts Resource to map representation
func convertResourceToLua(rs pdata.Resource) map[string]interface{} {
	return map[string]interface{}{
		"attributes": convertAttributeMapToLua(rs.Attributes()),
	}
}

// convertResourceFromLua converts from lua and fills Resource
func convertResourceFromLua(rs pdata.Resource, rsLua map[string]interface{}) {
	convertAttributeMapFromLua(rs.Attributes(), extractMapValueFrom(rsLua, "attributes"))
}

// convertAttributeValueToLua converts AttributeValue to map representation
func convertAttributeValueToLua(lr pdata.AttributeValue) map[string]interface{} {
	attrVal := map[string]interface{}{}

	switch lr.Type() {
	case pdata.AttributeValueTypeString:
		attrVal["stringVal"] = lr.StringVal()
	case pdata.AttributeValueTypeBool:
		attrVal["boolVal"] = lr.BoolVal()
	case pdata.AttributeValueTypeInt:
		attrVal["intVal"] = lr.IntVal()
	case pdata.AttributeValueTypeDouble:
		attrVal["doubleVal"] = lr.DoubleVal()
	case pdata.AttributeValueTypeMap:
		attrVal["mapVal"] = convertAttributeMapToLua(lr.MapVal())
	case pdata.AttributeValueTypeArray:
		attrVal["arrayVal"] = convertAttributeValueSliceToLua(lr.SliceVal())
	case pdata.AttributeValueTypeBytes:
		attrVal["bytesVal"] = lr.BytesVal()
	default:
		fmt.Println("Type not supported:", lr.Type())
	}

	return attrVal
}

// convertAttributeValueFromLua converts from lua to AttributeValue
func convertAttributeValueFromLua(lua map[string]interface{}) pdata.AttributeValue {
	if stringVal, ok := lua["stringVal"]; ok {
		return pdata.NewAttributeValueString(stringVal.(string))
	}
	if boolVal, ok := lua["boolVal"]; ok {
		return pdata.NewAttributeValueBool(boolVal.(bool))
	}
	if intVal, ok := lua["intVal"]; ok {
		return pdata.NewAttributeValueInt(intVal.(int64))
	}
	if doubleVal, ok := lua["doubleVal"]; ok {
		return pdata.NewAttributeValueDouble(doubleVal.(float64))
	}
	if mapVal, ok := lua["mapVal"]; ok {
		m := pdata.NewAttributeValueMap()
		convertAttributeMapFromLua(m.MapVal(), mapVal.(map[string]interface{}))
		return m
	}
	if arrayVal, ok := lua["arrayVal"]; ok {
		s := pdata.NewAttributeValueArray()
		convertAttributeValueSliceFromLua(s.SliceVal(), arrayVal.([]interface{}))
		return s
	}
	if bytesVal, ok := lua["bytesVal"]; ok {
		bytesValInterface := bytesVal.([]interface{})
		var bytes = make([]byte, len(bytesValInterface), len(bytesValInterface))
		for idx, byteValInterface := range bytesValInterface {
			bytes[idx] = byte(byteValInterface.(int64))
		}
		return pdata.NewAttributeValueBytes(bytes)
	}
	return pdata.NewAttributeValueEmpty()
}

// convertAttributeMapToLua converts AttributeMap to map representation
func convertAttributeMapToLua(attr pdata.AttributeMap) map[string]interface{} {
	lua := map[string]interface{}{}
	attr.Range(func(k string, v pdata.AttributeValue) bool {
		lua[k] = convertAttributeValueToLua(v)
		return true
	})
	return lua
}

// convertAttributeMapFromLua converts from lua and fills AttributeMap
func convertAttributeMapFromLua(attr pdata.AttributeMap, lua map[string]interface{}) {
	attr.EnsureCapacity(len(lua))
	for k, v := range lua {
		attr.Insert(k, convertAttributeValueFromLua(v.(map[string]interface{})))
	}
}

// convertAttributeValueSliceToLua converts AttributeValueSlice to map representation
func convertAttributeValueSliceToLua(sl pdata.AttributeValueSlice) []interface{} {
	luaSl := []interface{}{}
	for j := 0; j < sl.Len(); j++ {
		luaSl = append(luaSl, convertAttributeValueToLua(sl.At(j)))
	}
	return luaSl
}

// convertAttributeValueSliceFromLua converts from lua and fills AttributeValueSlice
func convertAttributeValueSliceFromLua(attr pdata.AttributeValueSlice, lua []interface{}) {
	attr.EnsureCapacity(len(lua))
	for _, v := range lua {
		convertAttributeValueFromLua(v.(map[string]interface{})).CopyTo(attr.AppendEmpty())
	}
}

func extractMapValueFrom(in map[string]interface{}, key string) map[string]interface{} {
	if out, ok := in[key]; ok && reflect.TypeOf(out).Kind() == reflect.Map {
		return out.(map[string]interface{})
	} else {
		return map[string]interface{}{}
	}
}

func extractSliceValueFrom(in map[string]interface{}, key string) []interface{} {
	if out, ok := in[key]; ok && reflect.TypeOf(out).Kind() == reflect.Slice {
		return out.([]interface{})
	} else {
		return []interface{}{}
	}
}

func withMap(in interface{}, f func(map[string]interface{})) {
	if reflect.TypeOf(in).Kind() == reflect.Map {
		f(in.(map[string]interface{}))
	}
}
