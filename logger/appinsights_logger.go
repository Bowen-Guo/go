package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
	"github.com/microsoft/ApplicationInsights-Go/appinsights/contracts"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Singleton
var appInsightsSyncer *AppInsightsSyncer

type LogRequest struct {
	InstrumentationKey string
	CustomDimensions   map[string]string
}

type AppInsightsSyncer struct {
	client           appinsights.TelemetryClient
	customDimensions context.Context
}

func NewAppInsightsCore() zapcore.Core {
	// instrumentationKey := "cefb0946-dbca-4b23-bc3c-9fc355a98436" // Do not check in.
	// telemetryConf := appinsights.NewTelemetryConfiguration(instrumentationKey)
	// telemetryClient := appinsights.NewTelemetryClientFromConfig(telemetryConf)
	appInsightsSyncer = &AppInsightsSyncer{client: nil, customDimensions: nil}
	writeSyncer := new(appInsightsSyncer)

	zapConf := zap.NewProductionEncoderConfig()
	jsonEncode := zapcore.NewJSONEncoder(zapConf)
	allLevels := zap.LevelEnablerFunc(func(l zapcore.Level) bool { return true })

	return zapcore.NewCore(jsonEncode, writeSyncer, allLevels)
}

// func SetConnectionString(connectionString string) {
// 	if appInsightsSyncer == nil {
// 		return
// 	}

// 	telemetryConf := appinsights.NewTelemetryConfiguration(connectionString)
// 	telemetryClient := appinsights.NewTelemetryClientFromConfig(telemetryConf)
// 	appInsightsSyncer.client = telemetryClient
// }

func InitializeAppInsightsLogger(c *gin.Context) {

	if appInsightsSyncer == nil {
		return
	}

	// Get instrumentation key from context.
	b, _ := ioutil.ReadAll(c.Request.Body)
	// if err != nil {
	// 	// 	panic(err)
	// }

	var data LogRequest
	json.Unmarshal(b, &data)
	if err := json.Unmarshal(b, &data); err != nil {
		fmt.Printf("Failed to unmarshal request body. Error: %s\n", err.Error())
		// c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if appInsightsSyncer.client == nil {
		fmt.Printf("InstrumentationKey = %s\n", data.InstrumentationKey)
		telemetryConf := appinsights.NewTelemetryConfiguration(data.InstrumentationKey)
		telemetryClient := appinsights.NewTelemetryClientFromConfig(telemetryConf)
		appInsightsSyncer.client = telemetryClient
	}

	fmt.Printf("CustomDimensions = %s\n", data.CustomDimensions)

	// appInsightsSyncer.customDimensions.Store("customDimensions", data.CustomDimensions)
	appInsightsSyncer.customDimensions = context.WithValue(context.Background(), "customDimensions", data.CustomDimensions)
}

func (appInsightsSyncer *AppInsightsSyncer) Sync() error {
	return nil
}

func (appInsightsSyncer *AppInsightsSyncer) Write(p []byte) (int, error) {
	if appInsightsSyncer.client == nil {
		return len(p), nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal(p, &data); err != nil {
		return len(p), err
	}

	trace := appInsightsSyncer.buildTrace(data)
	appInsightsSyncer.client.Track(trace)
	return len(p), nil
}

func (appInsightsSyncer *AppInsightsSyncer) buildTrace(data map[string]interface{}) *appinsights.TraceTelemetry {
	message := data["msg"].(string)
	level := levelMap[data["level"].(string)]
	trace := appinsights.NewTraceTelemetry(message, level)

	// Add custom dimensions as well.
	custom_dimensions := appInsightsSyncer.customDimensions.Value("customDimensions")
	custom_dict, _ := custom_dimensions.(map[string]string)
	for k, v := range custom_dict {
		data[k] = v
	}

	// Add custom dimentions.
	for k, v := range data {
		switch k {
		case "msg", "level":
			break
		default:
			switch v.(type) {
			case int:
				trace.BaseTelemetry.Properties[k] = string(v.(int))
			case string:
				trace.BaseTelemetry.Properties[k] = v.(string)
			case float64:
				trace.BaseTelemetry.Properties[k] = strconv.FormatFloat(v.(float64), 'f', 6, 64)
			}
		}
	}
	return trace
}

func new(appInsightsSyncer *AppInsightsSyncer) zapcore.WriteSyncer {
	return appInsightsSyncer
}

var levelMap = map[string]contracts.SeverityLevel{
	"Critical":    appinsights.Critical,
	"Error":       appinsights.Error,
	"Warning":     appinsights.Warning,
	"Information": appinsights.Information,
	"Verbose":     appinsights.Verbose,
}
