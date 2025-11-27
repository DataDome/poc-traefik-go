package modulego

import (
	"bytes"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockLogger struct {
	lastMessage string
}

func (m *MockLogger) Debug(args ...interface{}) {
	m.lastMessage = "DEBUG"
	fmt.Println(m.lastMessage, fmt.Sprint(args...))
}

func (m *MockLogger) Info(args ...interface{}) {
	m.lastMessage = "INFO"
	fmt.Println(m.lastMessage, fmt.Sprint(args...))
}

func (m *MockLogger) Warn(args ...interface{}) {
	m.lastMessage = "WARN"
	fmt.Println(m.lastMessage, fmt.Sprint(args...))
}

func (m *MockLogger) Error(args ...interface{}) {
	m.lastMessage = "ERROR"
	fmt.Println(m.lastMessage, fmt.Sprint(args...))
}

func TestDefaultLogger_Debug(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := &defaultLogger{
		logger: log.New(buffer, "", 0),
	}

	logger.Debug("Test debug message")

	expected := "DEBUG: Test debug message\n"
	assert.Equal(t, expected, buffer.String())
}

func TestDefaultLogger_Info(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := &defaultLogger{
		logger: log.New(buffer, "", 0),
	}

	logger.Info("Test info message")

	expected := "INFO: Test info message\n"
	assert.Equal(t, expected, buffer.String())
}

func TestDefaultLogger_Warn(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := &defaultLogger{
		logger: log.New(buffer, "", 0),
	}

	logger.Warn("Test warning message")

	expected := "WARN: Test warning message\n"
	assert.Equal(t, expected, buffer.String())
}

func TestDefaultLogger_Error(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := &defaultLogger{
		logger: log.New(buffer, "", 0),
	}

	logger.Error("Test error message")

	expected := "ERROR: Test error message\n"
	assert.Equal(t, expected, buffer.String())
}
