package luaprocessor

import (
	"go.opentelemetry.io/collector/model/pdata"
)

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
	// TODO
	return *cv.metrics, nil
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

// convertMetricGaugeToLua converts Gauge to map representation
func convertMetricGaugeToLua(m pdata.Gauge) map[string]interface{} {
	return map[string]interface{}{
		"dataPoints": convertMetricNumberDataPointSliceToLua(m.DataPoints()),
	}
}

// convertMetricNumberDataPointSliceToLua converts NumberDataPointSlice to map representation
func convertMetricNumberDataPointSliceToLua(dp pdata.NumberDataPointSlice) []interface{} {
	dpExp := []interface{}{}
	for j := 0; j < dp.Len(); j++ {
		dpExp = append(dpExp, convertMetricNumberDataPointToLua(dp.At(j)))
	}

	return dpExp
}

// convertMetricNumberDataPointToLua converts NumberDataPoint to map representation
func convertMetricNumberDataPointToLua(dp pdata.NumberDataPoint) map[string]interface{} {
	return map[string]interface{}{
		"attributes":     dp.Attributes().AsRaw(),
		"startTimestamp": uint64(dp.StartTimestamp()),
		"timestamp":      uint64(dp.Timestamp()),
		"value":          dp.DoubleVal(),
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

// convertMetricHistogramDataPointToLua converts HistogramDataPoint to map representation
func convertMetricHistogramDataPointToLua(dp pdata.HistogramDataPoint) map[string]interface{} {

	ex := dp.Exemplars()
	exExp := []interface{}{}
	for j := 0; j < ex.Len(); j++ {
		exExp = append(exExp, convertMetricExemplarToLua(ex.At(j)))
	}

	return map[string]interface{}{
		"attributes":     dp.Attributes().AsRaw(),
		"startTimestamp": uint64(dp.StartTimestamp()),
		"timestamp":      uint64(dp.Timestamp()),
		"count":          dp.Count(),
		"sum":            dp.Sum(),
		"bucketCounts":   dp.BucketCounts(),
		"explicitBounds": dp.ExplicitBounds(),
		"exemplars":      exExp,
	}
}

// convertMetricExemplarToLua converts Exemplar to map representation
func convertMetricExemplarToLua(ex pdata.Exemplar) map[string]interface{} {
	return map[string]interface{}{
		"attributes": ex.FilteredAttributes().AsRaw(),
		"timestamp":  uint64(ex.Timestamp()),
		"value":      ex.DoubleVal(),
		"traceID":    ex.TraceID(),
		"spanID":     ex.SpanID(),
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

// convertMetricExponentialHistogramDataPointToLua converts ExponentialHistogramDataPoint to map representation
func convertMetricExponentialHistogramDataPointToLua(dp pdata.ExponentialHistogramDataPoint) map[string]interface{} {

	ex := dp.Exemplars()
	exExp := []interface{}{}
	for j := 0; j < ex.Len(); j++ {
		exExp = append(exExp, convertMetricExemplarToLua(ex.At(j)))
	}

	return map[string]interface{}{
		"attributes":     dp.Attributes().AsRaw(),
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

// convertMetricSummaryDataPointToLua converts SummaryDataPoint to map representation
func convertMetricSummaryDataPointToLua(dp pdata.SummaryDataPoint) map[string]interface{} {
	qv := dp.QuantileValues()
	qvExp := []interface{}{}
	for j := 0; j < qv.Len(); j++ {
		qvExp = append(qvExp, convertMetricValueAtQuantileToLua(qv.At(j)))
	}

	return map[string]interface{}{
		"attributes":     dp.Attributes().AsRaw(),
		"startTimestamp": uint64(dp.StartTimestamp()),
		"timestamp":      uint64(dp.Timestamp()),
		"count":          dp.Count(),
		"sum":            dp.Sum(),
		"quantileValues": qvExp,
		"flags":          uint32(dp.Flags()),
	}
}

// convertMetricValueAtQuantileToLua converts ValueAtQuantile to map representation
func convertMetricValueAtQuantileToLua(qv pdata.ValueAtQuantile) map[string]interface{} {
	return map[string]interface{}{
		"quantile": qv.Quantile(),
		"value":    qv.Value(),
	}
}

// convertInstrumentationLibraryToLua converts InstrumentationLibrary to map representation
func convertInstrumentationLibraryToLua(il pdata.InstrumentationLibrary) map[string]interface{} {
	return map[string]interface{}{
		"name":    il.Name(),
		"version": il.Version(),
	}
}

// convertResourceToLua converts Resource to map representation
func convertResourceToLua(rs pdata.Resource) map[string]interface{} {
	return map[string]interface{}{
		"attributes": rs.Attributes().AsRaw(),
	}
}
