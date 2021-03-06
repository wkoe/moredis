package moredis

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/bson"
)

type toStringTestSpec struct {
	input    interface{}
	expected string
}

var toStringTests = []toStringTestSpec{
	{bson.ObjectIdHex("ffffffffffffffffffffffff"), "ffffffffffffffffffffffff"},
	{"string", "string"},
	{nil, "<nil>"},
}

func TestToString(t *testing.T) {
	for _, testCase := range toStringTests {
		actual := toString(testCase.input)
		assert.Equal(t, testCase.expected, actual, "toString(%v) failed", testCase.input)
	}
}

type safeToLowerTestSpec struct {
	input    interface{}
	expected string
}

var safeToLowerTests = []safeToLowerTestSpec{
	{"ALL CAPS", "all caps"},
	{nil, ""},
	{"lower", "lower"},
	{[]string{"test"}, ""},
}

func TestSafeToLower(t *testing.T) {
	for _, testCase := range safeToLowerTests {
		actual := safeToLower(testCase.input)
		assert.Equal(t, testCase.expected, actual, "safeToLower(%v) failed", testCase.input)
	}
}

type toSetTestSpec struct {
	input    interface{}
	expected string
}

var toSetTests = []toSetTestSpec{
	{bson.M{"key1": true, "key2": true}, "[\"key1\",\"key2\"]"},
	{bson.M{"key1": true, "key2": 5}, ""},
	{bson.M{"key1": "true", "key2": "t"}, "[\"key1\",\"key2\"]"},
	{nil, ""},
	{bson.M{"key1": false, "key2": true}, ""},
	{bson.M{"key1": "true", "key2": "false"}, ""},
}

func TestToSet(t *testing.T) {
	for _, testCase := range toSetTests {
		actual := toSet(testCase.input)
		assert.Equal(t, testCase.expected, actual, "toSet(%v) failed", testCase.input)
	}
}

type toJSONTestSpec struct {
	input    interface{}
	expected interface{}
}

var toJSONTests = []toJSONTestSpec{
	{bson.M{"key1": "val1", "key2": "val2"}, "{\"key1\":\"val1\",\"key2\":\"val2\"}"},
	{nil, nil},
	{"notanobject", "notanobject"},
}

func TestToJSON(t *testing.T) {
	for _, testCase := range toJSONTests {
		actual := toJSON(testCase.input)
		assert.Equal(t, testCase.expected, actual, "toJson(%v) failed", testCase.input)
	}
}

type applyTemplateTestSpec struct {
	name           string
	templateString string
	payload        bson.M
	expected       string
	expectedError  bool
}

var applyTemplateTests = []applyTemplateTestSpec{
	{
		name:           "empty input",
		templateString: "",
		payload:        bson.M{},
		expected:       "",
	},
	{
		name:           "simple field sub",
		templateString: "{{.field}}",
		payload:        bson.M{"field": "value"},
		expected:       "value",
	},
	{
		name:           "field sub and text",
		templateString: "text:{{.field}}",
		payload:        bson.M{"field": "value"},
		expected:       "text:value",
	},
	{
		name:           "function calling",
		templateString: "{{toLower .field}}",
		payload:        bson.M{"field": "VALUE"},
		expected:       "value",
	},
	{
		name:           "invalid template string",
		templateString: "{{()}}",
		payload:        bson.M{},
		expectedError:  true,
	},
	{
		name:           "invalid function in template",
		templateString: "{{nonExistentFunc}}",
		payload:        bson.M{},
		expectedError:  true,
	},
}

func TestApplyTemplate(t *testing.T) {
	for _, testCase := range applyTemplateTests {
		actual, err := ApplyTemplate(testCase.templateString, testCase.payload)
		if !testCase.expectedError {
			assert.Nil(t, err)
			assert.Equal(t, testCase.expected, actual, "failed applyTemplate test: %s", testCase.name)
		} else {
			assert.Error(t, err, "wanted error, but returned %s", actual)
		}
	}
}

func TestObjectIds(t *testing.T) {
	in := map[string]interface{}{
		"nil": nil,
		"id1": "ffffffffffffffffffffffff",
		"nested": map[string]interface{}{
			"id2": "111111111111111111111111",
			"int": 5,
			"str": "test",
		},
	}

	expected := map[string]interface{}{
		"nil": nil,
		"id1": bson.ObjectIdHex("ffffffffffffffffffffffff"),
		"nested": map[string]interface{}{
			"id2": bson.ObjectIdHex("111111111111111111111111"),
			"int": 5,
			"str": "test",
		},
	}

	setObjectIds(in)

	assert.Equal(t, expected, in)
}

type parseTemplatedJSONTestSpec struct {
	name          string
	queryString   string
	params        Params
	expected      map[string]interface{}
	expectedError bool
}

var parseTemplatedJSONTests = []parseTemplatedJSONTestSpec{
	{
		name:        "simple query with ObjectId substitution",
		queryString: `{"_id": "{{.id}}"}`,
		params:      Params{"id": "111111111111111111111111"},
		expected:    map[string]interface{}{"_id": bson.ObjectIdHex("111111111111111111111111")},
	},
	{
		name:        "simple substitution and mongo operator",
		queryString: `{"{{.field}}": {"$exists": true}}`,
		params:      Params{"field": "somefield"},
		expected:    map[string]interface{}{"somefield": map[string]interface{}{"$exists": true}},
	},
	{
		name:          "invalid json (missing quotes)",
		queryString:   `{id: 5}`,
		params:        Params{},
		expectedError: true,
	},
	{
		name:          "invalid template",
		queryString:   `{"field": {{()}}}`,
		params:        Params{},
		expectedError: true,
	},
}

func TestParseTemplatedJSON(t *testing.T) {
	for _, testCase := range parseTemplatedJSONTests {
		actual, err := ParseTemplatedJSON(testCase.queryString,
			testCase.params)
		if !testCase.expectedError {
			assert.Nil(t, err)
			assert.Equal(t, testCase.expected, actual)
		} else {
			assert.Error(t, err, "expected error, but got %s", actual)
		}
	}
}
