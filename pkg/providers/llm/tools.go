package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// TimeToolExecutor 获取当前时间的工具执行器
type TimeToolExecutor struct{}

func (t *TimeToolExecutor) Execute(ctx context.Context, arguments string) (string, error) {
	// 解析参数（如果需要的话）
	var args map[string]interface{}
	if arguments != "" {
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return "", fmt.Errorf("解析参数失败: %v", err)
		}
	}

	// 获取时区参数，默认为 UTC
	timezone := "UTC"
	if tz, ok := args["timezone"].(string); ok {
		timezone = tz
	}

	// 获取当前时间
	now := time.Now()
	if timezone != "UTC" {
		if loc, err := time.LoadLocation(timezone); err == nil {
			now = now.In(loc)
		}
	}

	// 返回格式化的时间
	result := map[string]interface{}{
		"current_time": now.Format("2006-01-02 15:04:05"),
		"timezone":     timezone,
		"timestamp":    now.Unix(),
	}

	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

func (t *TimeToolExecutor) GetName() string {
	return "get_current_time"
}

func (t *TimeToolExecutor) GetDescription() string {
	return "获取当前时间，支持指定时区"
}

func (t *TimeToolExecutor) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"timezone": map[string]interface{}{
				"type":        "string",
				"description": "时区，例如 'UTC', 'Asia/Shanghai', 'America/New_York'",
				"default":     "UTC",
			},
		},
		"required": []string{},
	}
}

// CalculatorToolExecutor 简单计算器工具执行器
type CalculatorToolExecutor struct{}

func (c *CalculatorToolExecutor) Execute(ctx context.Context, arguments string) (string, error) {
	var args struct {
		Operation string  `json:"operation"`
		A         float64 `json:"a"`
		B         float64 `json:"b"`
	}

	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("解析参数失败: %v", err)
	}

	var result float64
	switch args.Operation {
	case "add":
		result = args.A + args.B
	case "subtract":
		result = args.A - args.B
	case "multiply":
		result = args.A * args.B
	case "divide":
		if args.B == 0 {
			return "", fmt.Errorf("除数不能为零")
		}
		result = args.A / args.B
	default:
		return "", fmt.Errorf("不支持的操作: %s", args.Operation)
	}

	response := map[string]interface{}{
		"operation": args.Operation,
		"a":         args.A,
		"b":         args.B,
		"result":    result,
	}

	resultBytes, _ := json.Marshal(response)
	return string(resultBytes), nil
}

func (c *CalculatorToolExecutor) GetName() string {
	return "calculator"
}

func (c *CalculatorToolExecutor) GetDescription() string {
	return "执行基本的数学运算（加、减、乘、除）"
}

func (c *CalculatorToolExecutor) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "要执行的操作",
				"enum":        []string{"add", "subtract", "multiply", "divide"},
			},
			"a": map[string]interface{}{
				"type":        "number",
				"description": "第一个数字",
			},
			"b": map[string]interface{}{
				"type":        "number",
				"description": "第二个数字",
			},
		},
		"required": []string{"operation", "a", "b"},
	}
}

// WeatherToolExecutor 模拟天气查询工具执行器
type WeatherToolExecutor struct{}

func (w *WeatherToolExecutor) Execute(ctx context.Context, arguments string) (string, error) {
	var args struct {
		Location string `json:"location"`
		Unit     string `json:"unit,omitempty"`
	}

	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("解析参数失败: %v", err)
	}

	if args.Unit == "" {
		args.Unit = "celsius"
	}

	// 模拟天气数据
	weather := map[string]interface{}{
		"location":    args.Location,
		"temperature": 22,
		"unit":        args.Unit,
		"condition":   "晴朗",
		"humidity":    65,
		"wind_speed":  "5 km/h",
		"description": fmt.Sprintf("%s 当前天气晴朗，温度 22°C，湿度 65%%", args.Location),
	}

	if args.Unit == "fahrenheit" {
		weather["temperature"] = 72 // 转换为华氏度
	}

	resultBytes, _ := json.Marshal(weather)
	return string(resultBytes), nil
}

func (w *WeatherToolExecutor) GetName() string {
	return "get_weather"
}

func (w *WeatherToolExecutor) GetDescription() string {
	return "获取指定地点的天气信息"
}

func (w *WeatherToolExecutor) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"location": map[string]interface{}{
				"type":        "string",
				"description": "要查询天气的地点",
			},
			"unit": map[string]interface{}{
				"type":        "string",
				"description": "温度单位",
				"enum":        []string{"celsius", "fahrenheit"},
				"default":     "celsius",
			},
		},
		"required": []string{"location"},
	}
}

// RegisterDefaultTools 注册默认工具到管理器
func RegisterDefaultTools(manager *ModelManager) {
	manager.RegisterTool(&TimeToolExecutor{})
	manager.RegisterTool(&CalculatorToolExecutor{})
	manager.RegisterTool(&WeatherToolExecutor{})
}
