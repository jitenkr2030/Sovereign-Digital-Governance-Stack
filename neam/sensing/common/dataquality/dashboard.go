package dataquality

// DataQualityDashboard represents a Grafana dashboard for data quality monitoring
type DataQualityDashboard struct {
	Dashboard DashboardConfig `json:"dashboard"`
}

// DashboardConfig contains Grafana dashboard configuration
type DashboardConfig struct {
	ID              interface{}    `json:"id,omitempty"`
	UID             string         `json:"uid"`
	Title           string         `json:"title"`
	Tags            []string       `json:"tags"`
	Timezone        string         `json:"timezone"`
	Refresh         string         `json:"refresh"`
	Panels          []Panel        `json:"panels"`
	Annotations      AnnotationList `json:"annotations,omitempty"`
	Templating      TemplateList   `json:"templating,omitempty"`
	Time            TimeConfig     `json:"time"`
	Timepicker      Timepicker     `json:"timepicker,omitempty"`
	Variables       []Variable     `json:"variables,omitempty"`
	Links           []Link         `json:"links,omitempty"`
	Style           string         `json:"style"`
	GraphTooltip    int            `json:"graphTooltip"`
	PanelsJSON      string         `json:"panelsJson,omitempty"`
	AnnotationsJSON string         `json:"annotationsJson,omitempty"`
	SchemaVersion   int            `json:"schemaVersion"`
	Version         int            `json:"version"`
	PresetID        interface{}    `json:"presetId,omitempty"`
}

// Panel represents a Grafana panel
type Panel struct {
	ID             int               `json:"id"`
	Title          string            `json:"title"`
	Type           string            `json:"type"`
	GridPos        GridPos           `json:"gridPos"`
	Transparent    bool              `json:"transparent,omitempty"`
	Targets        []Target          `json:"targets"`
	Transformations []Transformation `json:"transformations,omitempty"`
	FieldConfig    FieldConfig       `json:"fieldConfig,omitempty"`
	Options        map[string]interface{} `json:"options,omitempty"`
	PluginVersion  string            `json:"pluginVersion,omitempty"`
	Datasource     Datasource        `json:"datasource"`
}

// GridPos contains panel position
type GridPos struct {
	H int `json:"h"`
	W int `json:"w"`
	X int `json:"x"`
	Y int `json:"y"`
}

// Target represents a panel query target
type Target struct {
	Expr          string            `json:"expr,omitempty"`
	RefID         string            `json:"refId"`
	Instant       bool              `json:"instant,omitempty"`
	Interval      string            `json:"interval,omitempty"`
	LegendFormat  string            `json:"legendFormat,omitempty"`
	Datasource    Datasource        `json:"datasource,omitempty"`
	QueryType     string            `json:"queryType,omitempty"`
	ResultFormat  string            `json:"resultFormat,omitempty"`
	Hide          bool              `json:"hide,omitempty"`
	Metric        string            `json:"metric,omitempty"`
	Metrics       []string          `json:"metrics,omitempty"`
	Filters       []Filter          `json:"filters,omitempty"`
	GroupBy       []string          `json:"groupBy,omitempty"`
	Aggregator    string            `json:"aggregator,omitempty"`
	Downsampler   string            `json:"downsampler,omitempty"`
}

// Filter represents a metric filter
type Filter struct {
	Key    string `json:"key"`
	Type   string `json:"type"`
	Filter string `json:"filter"`
	Params []string `json:"params,omitempty"`
	AdHoc  bool   `json:"adHoc,omitempty"`
}

// Datasource represents a data source reference
type Datasource struct {
	Type string `json:"type"`
	UID  string `json:"uid"`
}

// FieldConfig contains field configuration
type FieldConfig struct {
	Defaults FieldDefaults `json:"defaults"`
	Overrides []Override   `json:"overrides,omitempty"`
}

// FieldDefaults contains default field settings
type FieldDefaults struct {
	Unit        string            `json:"unit,omitempty"`
	Min         interface{}       `json:"min,omitempty"`
	Max         interface{}       `json:"max,omitempty"`
	Decimals    interface{}       `json:"decimals,omitempty"`
	DisplayName string            `json:"displayName,omitempty"`
	Thresholds  []Threshold       `json:"thresholds,omitempty"`
	ColorMode   string            `json:"colorMode,omitempty"`
	Mappings    []Mapping         `json:"mappings,omitempty"`
}

// Threshold represents a threshold for coloring
type Threshold struct {
	Value interface{} `json:"value"`
	Color string     `json:"color"`
}

// Mapping represents a value mapping
type Mapping struct {
	Type    string       `json:"type"`
	Options []MappingOpt `json:"options,omitempty"`
}

// MappingOpt represents a single value mapping option
type MappingOpt struct {
	ID        string      `json:"id"`
	Value     string      `json:"value,omitempty"`
	Text      string      `json:"text,omitempty"`
	Type      string      `json:"type,omitempty"`
	From      string      `json:"from,omitempty"`
	To        string      `json:"to,omitempty"`
	Operator  string      `json:"operator,omitempty"`
}

// Override represents a field override
type Override struct {
	Matcher Matcher     `json:"matcher"`
	Properties []Prop   `json:"properties"`
}

// Matcher represents a field matcher
type Matcher struct {
	ID   string `json:"id"`
	Options interface{} `json:"options,omitempty"`
}

// Prop represents an override property
type Prop struct {
	ID    string      `json:"id"`
	Value interface{} `json:"value"`
}

// Transformation represents a data transformation
type Transformation struct {
	ID             string                 `json:"id"`
	Options        map[string]interface{} `json:"options"`
}

// AnnotationList contains annotation configuration
type AnnotationList struct {
	List []Annotation `json:"list"`
}

// Annotation represents an annotation query
type Annotation struct {
	Datasource Datasource `json:"datasource"`
	Enable     bool       `json:"enable"`
	Hide       bool       `json:"hide,omitempty"`
	IconColor  string     `json:"iconColor"`
	Name       string     `json:"name"`
	Type       string     `json:"type"`
	BuiltIn    int        `json:"builtIn,omitempty"`
}

// TemplateList contains template variables
type TemplateList struct {
	List []Variable `json:"list"`
}

// Variable represents a template variable
type Variable struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Label       string            `json:"label,omitempty"`
	Hide        int               `json:"hide,omitempty"`
	Query       string            `json:"query,omitempty"`
	Datasource  Datasource        `json:"datasource,omitempty"`
	Options     []VariableOption  `json:"options,omitempty"`
	Current     VariableCurrent   `json:"current,omitempty"`
	Refresh     int               `json:"refresh,omitempty"`
	Regex       string            `json:"regex,omitempty"`
	Sort        int               `json:"sort,omitempty"`
	Multi       bool              `json:"multi,omitempty"`
	IncludeAll  bool              `json:"includeAll,omitempty"`
}

// VariableOption represents a variable option
type VariableOption struct {
	Text     string `json:"text"`
	Value    string `json:"value"`
	Selected bool   `json:"selected,omitempty"`
}

// VariableCurrent represents the current selection
type VariableCurrent struct {
	Text  string `json:"text,omitempty"`
	Value string `json:"value,omitempty"`
}

// TimeConfig contains time range configuration
type TimeConfig struct {
	From string `json:"from"`
	To   string `json:"to"`
	Now  bool   `json:"now,omitempty"`
}

// Timepicker contains time picker configuration
type Timepicker struct {
	RefreshIntervals []string `json:"refresh_intervals,omitempty"`
	TimeOptions      []string `json:"time_options,omitempty"`
}

// Link represents a dashboard link
type Link struct {
	Title       string     `json:"title"`
	Type        string     `json:"type"`
	Url         string     `json:"url,omitempty"`
	Tooltip      string    `json:"tooltip,omitempty"`
	Icon        string     `json:"icon,omitempty"`
	DashURI     string     `json:"dashUri,omitempty"`
	Dashboard   string     `json:"dashboard,omitempty"`
	AsDropdown  bool       `json:"asDropdown,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	TargetBlank bool       `json:"targetBlank,omitempty"`
}

// GenerateDataQualityDashboard creates a comprehensive data quality dashboard
func GenerateDataQualityDashboard() *DataQualityDashboard {
	dashboard := &DataQualityDashboard{
		Dashboard: DashboardConfig{
			UID:       "neam-data-quality",
			Title:     "NEAM Data Quality Monitor",
			Tags:      []string{"neam", "data-quality", "monitoring"},
			Timezone:  "browser",
			Refresh:   "30s",
			Style:     "dark",
			GraphTooltip: 1,
			SchemaVersion: 38,
			Version:   1,
			Time: TimeConfig{
				From: "now-6h",
				To:   "now",
			},
			Variables: []Variable{
				{
					Name:       "datasource",
					Type:       "datasource",
					Query:      "prometheus",
					Refresh:    1,
					Datasource: Datasource{Type: "prometheus", UID: "prometheus"},
				},
				{
					Name:      "source",
					Type:      "query",
					Label:     "Data Source",
					Query:     "label_values(neam_data_quality_overall_score, source)",
					Refresh:   2,
					Multi:     true,
					IncludeAll: true,
					Datasource: Datasource{Type: "prometheus", UID: "prometheus"},
				},
				{
					Name:      "adapter",
					Type:      "query",
					Label:     "Adapter",
					Query:     "label_values(neam_data_quality_overall_score, source)",
					Regex:     "/^(transport|industrial|energy)/",
					Refresh:   2,
					Multi:     true,
					IncludeAll: true,
					Datasource: Datasource{Type: "prometheus", UID: "prometheus"},
				},
			},
		},
	}

	// Add panels
	dashboard.Dashboard.Panels = []Panel{
		// Overall Quality Score Gauge
		{
			ID:      1,
			Title:   "Overall Data Quality Score",
			Type:    "gauge",
			GridPos: GridPos{H: 8, W: 6, X: 0, Y: 0},
			Targets: []Target{
				{
					Expr:   "avg(neam_data_quality_overall_score{source=~\"$source|$adapter\"})",
					RefID:  "A",
				},
			},
			FieldConfig: FieldConfig{
				Defaults: FieldDefaults{
					Unit:   "percent",
					Min:    0,
					Max:    100,
					Thresholds: []Threshold{
						{Value: nil, Color: "red"},
						{Value: 70, Color: "yellow"},
						{Value: 90, Color: "green"},
					},
				},
			},
			Options: map[string]interface{}{
				"reduceOptions": map[string]interface{}{
					"calcs": []string{"lastNotNull"},
				},
				"orientation": "auto",
				"showThresholdLabels": false,
				"showThresholdMarkers": true,
			},
			Datasource: Datasource{Type: "prometheus", UID: "prometheus"},
		},
		// Quality by Dimension Bar Gauge
		{
			ID:      2,
			Title:   "Quality by Dimension",
			Type:    "bargauge",
			GridPos: GridPos{H: 8, W: 6, X: 6, Y: 0},
			Targets: []Target{
				{
					Expr:   "neam_data_quality_completeness_score{source=~\"$source|$adapter\"}",
					RefID:  "A",
				},
				{
					Expr:   "neam_data_quality_validity_score{source=~\"$source|$adapter\"}",
					RefID:  "B",
				},
				{
					Expr:   "neam_data_quality_consistency_score{source=~\"$source|$adapter\"}",
					RefID:  "C",
				},
				{
					Expr:   "neam_data_quality_timeliness_score{source=~\"$source|$adapter\"}",
					RefID:  "D",
				},
			},
			FieldConfig: FieldConfig{
				Defaults: FieldDefaults{
					Unit:   "percent",
					Min:    0,
					Max:    100,
				},
			},
			Datasource: Datasource{Type: "prometheus", UID: "prometheus"},
		},
		// Quality Score Trend
		{
			ID:      3,
			Title:   "Quality Score Over Time",
			Type:    "timeseries",
			GridPos: GridPos{H: 8, W: 12, X: 12, Y: 0},
			Targets: []Target{
				{
					Expr:   "neam_data_quality_overall_score{source=~\"$source|$adapter\"}",
					RefID:  "A",
					LegendFormat: "{{source}}",
				},
			},
			FieldConfig: FieldConfig{
				Defaults: FieldDefaults{
					Unit:   "percent",
					Min:    0,
					Max:    100,
				},
			},
			Options: map[string]interface{}{
				"legend": map[string]interface{}{
					"displayMode": "table",
					"placement":   "bottom",
					"showLegend":  true,
				},
			},
			Datasource: Datasource{Type: "prometheus", UID: "prometheus"},
		},
		// Validation Success Rate
		{
			ID:      4,
			Title:   "Schema Validation Success Rate",
			Type:    "timeseries",
			GridPos: GridPos{H: 8, W: 8, X: 0, Y: 8},
			Targets: []Target{
				{
					Expr:   "sum(rate(neam_schema_validation_total{source=~\"$source|$adapter\",status=\"success\"}[5m])) / sum(rate(neam_schema_validation_total{source=~\"$source|$adapter\"}[5m])) * 100",
					RefID:  "A",
					LegendFormat: "Success Rate",
				},
			},
			FieldConfig: FieldConfig{
				Defaults: FieldDefaults{
					Unit:   "percent",
					Min:    0,
					Max:    100,
				},
			},
			Datasource: Datasource{Type: "prometheus", UID: "prometheus"},
		},
		// Validation Errors by Type
		{
			ID:      5,
			Title:   "Validation Errors by Type",
			Type:    "barchart",
			GridPos: GridPos{H: 8, W: 8, X: 8, Y: 8},
			Targets: []Target{
				{
					Expr:   "sum by (error_type) (rate(neam_schema_validation_errors_by_type_total{source=~\"$source|$adapter\"}[5m]))",
					RefID:  "A",
					LegendFormat: "{{error_type}}",
				},
			},
			Datasource: Datasource{Type: "prometheus", UID: "prometheus"},
		},
		// Anomaly Detection
		{
			ID:      6,
			Title:   "Anomalies Detected",
			Type:    "timeseries",
			GridPos: GridPos{H: 8, W: 8, X: 16, Y: 8},
			Targets: []Target{
				{
					Expr:   "sum by (severity) (rate(neam_data_anomaly_detected_total{source=~\"$source|$adapter\"}[5m]))",
					RefID:  "A",
					LegendFormat: "{{severity}}",
				},
			},
			Datasource: Datasource{Type: "prometheus", UID: "prometheus"},
		},
		// Data Completeness Heatmap
		{
			ID:      7,
			Title:   "Data Completeness by Field",
			Type:    "heatmap",
			GridPos: GridPos{H: 8, W: 12, X: 0, Y: 16},
			Targets: []Target{
				{
					Expr:   "neam_data_quality_completeness_score{source=~\"$source|$adapter\"}",
					RefID:  "A",
				},
			},
			Datasource: Datasource{Type: "prometheus", UID: "prometheus"},
		},
		// Quality by Source
		{
			ID:      8,
			Title:   "Quality Score Distribution by Source",
			Type:    "histogram",
			GridPos: GridPos{H: 8, W: 12, X: 12, Y: 16},
			Targets: []Target{
				{
					Expr:   "neam_data_quality_score_histogram_bucket{source=~\"$source|$adapter\"}",
					RefID:  "A",
				},
			},
			Datasource: Datasource{Type: "prometheus", UID: "prometheus"},
		},
		// Latency Impact on Quality
		{
			ID:      9,
			Title:   "Data Timeliness (Latency)",
			Type:    "timeseries",
			GridPos: GridPos{H: 8, W: 12, X: 0, Y: 24},
			Targets: []Target{
				{
					Expr:   "rate(neam_data_quality_timeliness_lag_seconds_sum{source=~\"$source|$adapter\"}[5m]) / rate(neam_data_quality_timeliness_lag_seconds_count{source=~\"$source|$adapter\"}[5m])",
					RefID:  "A",
					LegendFormat: "{{source}}",
				},
			},
			FieldConfig: FieldConfig{
				Defaults: FieldDefaults{
					Unit: "s",
				},
			},
			Datasource: Datasource{Type: "prometheus", UID: "prometheus"},
		},
		// Active Alerts Table
		{
			ID:      10,
			Title:   "Active Data Quality Alerts",
			Type:    "table",
			GridPos: GridPos{H: 8, W: 12, X: 12, Y: 24},
			Targets: []Target{
				{
					Expr:   "ALERTS{source=~\"$source|$adapter\",alertstate=\"firing\"}",
					RefID:  "A",
				},
			},
			Datasource: Datasource{Type: "prometheus", UID: "prometheus"},
		},
	}

	return dashboard
}

// ToJSON converts the dashboard to JSON
func (d *DataQualityDashboard) ToJSON() ([]byte, error) {
	return json.MarshalIndent(d, "", "  ")
}
