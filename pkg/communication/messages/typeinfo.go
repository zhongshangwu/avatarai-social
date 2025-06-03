package messages

import (
	"encoding/json"
	"fmt"
)

type TypeExtractor struct {
	Type string `json:"type"`
}

type TagExtractor struct {
	Tag string `json:"tag"`
}

func ExtractType(data []byte) (string, error) {
	var extractor TypeExtractor
	if err := json.Unmarshal(data, &extractor); err != nil {
		return "", fmt.Errorf("提取类型错误: %w", err)
	}
	return extractor.Type, nil
}

func ExtractTag(data []byte) (string, error) {
	var extractor TagExtractor
	if err := json.Unmarshal(data, &extractor); err != nil {
		return "", fmt.Errorf("提取标签错误: %w", err)
	}
	return extractor.Tag, nil
}

type TypeMismatchError struct {
	Expected string
	Actual   string
}

func (e TypeMismatchError) Error() string {
	return fmt.Sprintf("类型不匹配: 期望 %s, 实际为 %s", e.Expected, e.Actual)
}

type TagMismatchError struct {
	Expected string
	Actual   string
}

func (e TagMismatchError) Error() string {
	return fmt.Sprintf("标签不匹配: 期望 %s, 实际为 %s", e.Expected, e.Actual)
}

func ValidateType(data []byte, expectedType string) error {
	typeName, err := ExtractType(data)
	if err != nil {
		return err
	}
	if typeName != expectedType {
		return &TypeMismatchError{
			Expected: expectedType,
			Actual:   typeName,
		}
	}
	return nil
}

func ValidateTag(data []byte, expectedTag string) error {
	tagName, err := ExtractTag(data)
	if err != nil {
		return err
	}
	if tagName != expectedTag {
		return &TagMismatchError{
			Expected: expectedTag,
			Actual:   tagName,
		}
	}
	return nil
}

func UnmarshalWithType(data []byte, expectedType string, target interface{}) error {
	if err := ValidateType(data, expectedType); err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func UnmarshalWithTag(data []byte, expectedTag string, target interface{}) error {
	if err := ValidateTag(data, expectedTag); err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func ExtractTypeFrom(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("序列化错误: %w", err)
	}
	return ExtractType(data)
}

func ExtractTagFrom(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("序列化错误: %w", err)
	}
	return ExtractTag(data)
}
