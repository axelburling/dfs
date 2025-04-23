package msg

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"reflect"

	"github.com/axelburling/dfs/pkg/pubsub/client/apiv1/message"
)

type Action string

var actionRegistry = map[Action]reflect.Type{}

func init() {
	registerAction[NodeAddReq](NodeAdd)
	registerAction[NodeDeleteReq](NodeDelete)
	registerAction[NodeUpdateReq](NodeUpdate)
}

const (
	NodeAdd    Action = "node:add"
	NodeDelete Action = "node:delete"
	NodeUpdate Action = "node:update"
)

type NodeAddReq struct {
	ID          string
	Address     string
	GrpcAddress string
	Hostname    string
	IsHealthy   bool
	TotalSpace  int64
	FreeSpace   int64
	Readonly    bool
}

type NodeDeleteReq struct {
	ID string
}

type NodeUpdateReq struct {
	ID  string
	Age int
}

func registerAction[T any](action Action) {
	var t T
	typ := reflect.TypeOf(t)

	// Register with gob for encoding/decoding
	gob.Register(t)

	// Store in registry
	actionRegistry[action] = typ
}

func (a Action) String() string {
	return string(a)
}

func ParseAction(actionStrs []string) ([]string, error) {
	var actions []string

	for _, actionStr := range actionStrs {
		action := Action(actionStr)
		if _, exists := actionRegistry[action]; !exists {
			return nil, fmt.Errorf("invalid action: %s", actionStr)
		}
		actions = append(actions, actionStr)
	}

	return actions, nil
}

func Create[T any](action Action, req T) (*message.Message, error) {
	// Validate the struct
	if err := validateStruct(&req); err != nil {
		return nil, fmt.Errorf("validation failed for %s: %w", action, err)
	}

	// Call the create method on the action
	return action.create(req)
}

func Decode[T any](msg *message.Message) (*T, error) {
	action := Action(msg.Action)

	// Find the type from the registry
	typ, exists := actionRegistry[action]
	if !exists {
		return nil, fmt.Errorf("unknown message action: %s", msg.Action)
	}

	// Ensure the expected type matches the registered type
	expectedType := reflect.TypeOf((*T)(nil)).Elem()
	if expectedType != typ {
		return nil, fmt.Errorf("type mismatch for action %s: expected %v, got %v", msg.Action, expectedType, typ)
	}

	// Create a new instance of the struct
	dataPtr := reflect.New(typ).Interface()

	// Decode the message data
	if err := decode(msg.Data, dataPtr); err != nil {
		return nil, err
	}

	return dataPtr.(*T), nil
}

func (a Action) create(req any) (*message.Message, error) {
	expectedType, exists := actionRegistry[a]
	if !exists {
		return nil, fmt.Errorf("unknown action: %s", a)
	}

	// Ensure the provided req matches the expected type
	if reflect.TypeOf(req) != expectedType {
		return nil, fmt.Errorf("invalid request type for action %s: expected %v, got %v", a, expectedType, reflect.TypeOf(req))
	}

	data, err := encode(req)
	if err != nil {
		return nil, err
	}

	return message.NewMessage(a.String(), data), nil
}

func encode(data any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(data); err != nil {
		return nil, fmt.Errorf("failed to encode: %w", err)
	}
	return buf.Bytes(), nil
}

func decode(data []byte, v any) error {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("failed to decode: %w", err)
	}
	return nil
}

func validateStruct[T any](obj *T) error {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New("expected a non-nil pointer to a struct")
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return errors.New("expected a struct")
	}

	var missingFields []string

	for i := range v.NumField() {
		field := v.Type().Field(i)
		value := v.Field(i)

		// Check if field is zero (empty)
		if value.Kind() != reflect.Bool && value.IsZero() {
			return fmt.Errorf("missing required field: %s", field.Name)
		}
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("missing required fields: %v", missingFields)
	}

	return nil
}
