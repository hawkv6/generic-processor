package processor

import (
	"reflect"
	"testing"

	"github.com/hawkv6/generic-processor/pkg/message"
	"github.com/stretchr/testify/assert"
)

func TestNewMinMaxNormalizer(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestNewMinMaxNormalizer",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, NewMinMaxNormalizer())
		})
	}
}

func TestMinMaxNormalizer_AddDataToNormalize(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		value     float64
	}{
		{
			name:      "TestMinMaxNormalizer_AddDataToNormalize",
			fieldName: "field1",
			value:     1.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizer := NewMinMaxNormalizer()
			normalizer.AddDataToNormalize(tt.fieldName, tt.value)
			assert.Equal(t, 1, len(normalizer.normalizationData[tt.fieldName]))
		})
	}
}

func TestMinMaxNormalizer_ResetNormalizationData(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "TestMinMaxNormalizer_ResetNormalizationData",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizer := NewMinMaxNormalizer()
			normalizer.normalizationData["field1"] = []float64{1.0}
			normalizer.ResetNormalizationData()
			assert.Equal(t, 0, len(normalizer.normalizationData["field1"]))
		})
	}
}

func TestMinMaxNormalizer_calculalteQuartiles(t *testing.T) {
	tests := []struct {
		name                   string
		data                   []float64
		wantQ1                 float64
		wantQ2                 float64
		wantQ3                 float64
		wantInterQuartileRange float64
		wantErr                bool
	}{
		{
			name:                   "TestMinMaxNormalizer_calculalteQuartiles success",
			data:                   []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			wantQ1:                 1.5,
			wantQ2:                 3.0,
			wantQ3:                 4.5,
			wantInterQuartileRange: 3.0,
			// calculation info: https://www.calculatorsoup.com/calculators/statistics/descriptivestatistics.php
		},
		{
			name:                   "TestMinMaxNormalizer_calculalteQuartiles empty imput error",
			data:                   []float64{},
			wantQ1:                 0,
			wantQ2:                 0,
			wantQ3:                 0,
			wantInterQuartileRange: 0,
			wantErr:                true,
			// calculation info: https://www.calculatorsoup.com/calculators/statistics/descriptivestatistics.php
		},
		{
			name:                   "TestMinMaxNormalizer_calculalteQuartiles calculation error",
			data:                   []float64{0},
			wantQ1:                 1.5,
			wantQ2:                 3.0,
			wantQ3:                 4.5,
			wantInterQuartileRange: 3.0,
			wantErr:                true,
			// calculation info: https://www.calculatorsoup.com/calculators/statistics/descriptivestatistics.php
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizer := NewMinMaxNormalizer()
			quartiles, interQuartileRange, err := normalizer.calculateQuartiles(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, quartiles)
			assert.Equal(t, tt.wantQ1, quartiles.Q1)
			assert.Equal(t, tt.wantQ2, quartiles.Q2)
			assert.Equal(t, tt.wantQ3, quartiles.Q3)
			assert.Equal(t, tt.wantInterQuartileRange, interQuartileRange)
		})
	}
}

func TestMinMaxNormalizer_getFences(t *testing.T) {
	tests := []struct {
		name            string
		data            []float64
		wantUpperFence  float64
		wantLowerFence  float64
		wantQuartileErr bool
	}{
		{
			name:            "TestMinMaxNormalizer_getFences success",
			data:            []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			wantUpperFence:  5, // max = 5, IQR = 3, Q3 = 4.5 --> 4.5 + 1.5 * 3 = 9 --> 5 is the min value
			wantLowerFence:  1, // min = 1, IQR = 3, Q1 = 1.5 --> 1.5 - 1.5 * 3 = -3 --> 1 is the max value
			wantQuartileErr: false,
			// calculation info: https://www.calculatorsoup.com/calculators/statistics/descriptivestatistics.php
		},
		{
			name:            "TestMinMaxNormalizer_getFences calculation error",
			data:            []float64{0},
			wantQuartileErr: true,
		},
		{
			name:            "TestMinMaxNormalizer_getFences empty input",
			data:            []float64{},
			wantQuartileErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizer := NewMinMaxNormalizer()
			upperFence := 0.0
			lowerFence := 0.0
			statisticalData, err := normalizer.getFences(tt.data, &upperFence, &lowerFence)
			if tt.wantQuartileErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, statisticalData)
			assert.Equal(t, tt.wantUpperFence, upperFence)
			assert.Equal(t, tt.wantLowerFence, lowerFence)
		})
	}
}

func TestMinMaxNormalizer_getNormalizedValue(t *testing.T) {
	tests := []struct {
		name           string
		value          float64
		upperFence     float64
		lowerFence     float64
		wantNormalized float64
	}{
		{
			name:           "TestMinMaxNormalizer_getNormalizedValue success ",
			value:          10,
			upperFence:     15,
			lowerFence:     5,
			wantNormalized: 0.5,
		},
		{
			name:           "TestMinMaxNormalizer_getNormalizedValue success normalized value smaller 0",
			value:          2,
			upperFence:     15,
			lowerFence:     5,
			wantNormalized: 0.00001,
		},
		{
			name:           "TestMinMaxNormalizer_getNormalizedValue success normalized value bigger than 1",
			value:          20.0,
			upperFence:     15.0,
			lowerFence:     5.0,
			wantNormalized: 1,
		},
		{
			name:           "TestMinMaxNormalizer_getNormalizedValue upperFence == lowerFence",
			value:          3.0,
			upperFence:     0.0,
			lowerFence:     0.0,
			wantNormalized: 1.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizer := NewMinMaxNormalizer()
			normalizedValue := normalizer.getNormalizedValue(tt.value, tt.upperFence, tt.lowerFence)
			assert.Equal(t, tt.wantNormalized, normalizedValue)
		})
	}
}

func TestMinMaxNormalizer_NormalizeValues(t *testing.T) {
	tests := []struct {
		name                string
		updates             map[string]message.ArangoUpdate
		inputField          string
		normalizationField  string
		wantStatisticalData *StatisticalData
		wantErr             bool
	}{
		{
			name: "TestMinMaxNormalizer_NormalizeValues success",
			updates: map[string]message.ArangoUpdate{
				"key1": {
					Fields: []string{"field1"},
					Values: []float64{10},
				},
			},
			inputField:         "field1",
			normalizationField: "normalizedField",
			wantStatisticalData: &StatisticalData{
				q1:                 7.5,
				q3:                 22.5,
				interQuartileRange: 15,
				lowerFence:         5,
				upperFence:         25,
			},
			wantErr: false,
		},
		{
			name: "TestMinMaxNormalizer_NormalizeValues empty/wrong field",
			updates: map[string]message.ArangoUpdate{
				"key1": {
					Fields: []string{"field1"},
					Values: []float64{10},
				},
			},
			inputField:          "",
			normalizationField:  "normalizedField",
			wantStatisticalData: nil,
			wantErr:             true,
		},
		{
			name: "TestMinMaxNormalizer_NormalizeValues no values",
			updates: map[string]message.ArangoUpdate{
				"key1": {
					Fields: []string{"field1"},
					Values: []float64{},
				},
			},
			inputField:         "field1",
			normalizationField: "normalizedField",
			wantStatisticalData: &StatisticalData{
				q1:                 7.5,
				q3:                 22.5,
				interQuartileRange: 15,
				lowerFence:         5,
				upperFence:         25,
			},
			wantErr: false,
		},
		{
			name: "TestMinMaxNormalizer_NormalizeValues can not calculate fences",
			updates: map[string]message.ArangoUpdate{
				"key1": {
					Fields: []string{"field1"},
					Values: []float64{10},
				},
			},
			inputField:          "field1",
			normalizationField:  "normalizedField",
			wantStatisticalData: nil,
			wantErr:             true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizer := NewMinMaxNormalizer()
			if tt.wantErr && tt.inputField != "" {
				normalizer.AddDataToNormalize(tt.inputField, 0)
			} else if tt.inputField != "" {
				for _, value := range []float64{5.0, 10.0, 15.0, 20.0, 25.0} {
					normalizer.AddDataToNormalize(tt.inputField, value)
				}
			}
			statisticalData, err := normalizer.NormalizeValues(tt.updates, tt.inputField, tt.normalizationField)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, statisticalData)
			if !reflect.DeepEqual(tt.wantStatisticalData, statisticalData) {
				t.Errorf("want: %+v, got: %+v", tt.wantStatisticalData, statisticalData)
			}
		})
	}
}
