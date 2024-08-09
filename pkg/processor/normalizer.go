package processor

import (
	"fmt"
	"math"

	"github.com/hawkv6/generic-processor/pkg/logging"
	"github.com/hawkv6/generic-processor/pkg/message"
	"github.com/montanaflynn/stats"
	"github.com/sirupsen/logrus"
)

type Normalizer interface {
	NormalizeValues(map[string]message.ArangoUpdate, string, string) (*StatisticalData, error)
	AddDataToNormalize(string, float64)
	ResetNormalizationData()
}

type MinMaxNormalizer struct {
	log               *logrus.Entry
	normalizationData map[string][]float64
}

func NewMinMaxNormalizer() *MinMaxNormalizer {
	return &MinMaxNormalizer{
		log:               logging.DefaultLogger.WithField("subsystem", Subsystem),
		normalizationData: make(map[string][]float64),
	}
}

func (normalizer *MinMaxNormalizer) AddDataToNormalize(fieldName string, value float64) {
	if _, ok := normalizer.normalizationData[fieldName]; !ok {
		normalizer.normalizationData[fieldName] = make([]float64, 0)
	}
	normalizer.normalizationData[fieldName] = append(normalizer.normalizationData[fieldName], value)
}

func (normalizer *MinMaxNormalizer) ResetNormalizationData() {
	normalizer.normalizationData = make(map[string][]float64)
}

func (normalizer *MinMaxNormalizer) calculateQuartiles(data stats.Float64Data) (*stats.Quartiles, float64, error) {
	quartiles, err := stats.Quartile(data)
	if err != nil || math.IsNaN(quartiles.Q1) || math.IsNaN(quartiles.Q2) || math.IsNaN(quartiles.Q3) {
		return nil, 0, fmt.Errorf("Error calculating quartiles %v", err)
	}
	interQuartileRange := quartiles.Q3 - quartiles.Q1
	normalizer.log.Debugln("Q1: ", quartiles.Q1)
	normalizer.log.Debugln("Q2 / Median: ", quartiles.Q2)
	normalizer.log.Debugln("Q3: ", quartiles.Q3)
	normalizer.log.Debugln("Interquartile range: ", interQuartileRange)
	outliers, _ := stats.QuartileOutliers(data)
	normalizer.log.Debugf("Outliers: %+v", outliers)
	return &quartiles, interQuartileRange, nil
}

func (normalizer *MinMaxNormalizer) getFences(data stats.Float64Data, upperFence, lowerFence *float64) (*StatisticalData, error) {
	quartiles, interQuartileRange, err := normalizer.calculateQuartiles(data)
	if err != nil {
		return nil, err
	}
	min, _ := stats.Min(data)
	max, _ := stats.Max(data)
	*upperFence = math.Min(quartiles.Q3+1.5*interQuartileRange, max)
	normalizer.log.Debugln("Upper fence: ", *upperFence)
	*lowerFence = math.Max(quartiles.Q1-1.5*interQuartileRange, min)
	normalizer.log.Debugln("Lower fence: ", *lowerFence)
	return NewStatisticalData(quartiles.Q1, quartiles.Q3, interQuartileRange, *lowerFence, *upperFence), nil
}

func (normalizer *MinMaxNormalizer) getNormalizedValue(value float64, upperFence, lowerFence float64) float64 {
	normalizedValue := 1.0
	if upperFence != lowerFence {
		normalizedValue = (value - lowerFence) / (upperFence - lowerFence)
		if normalizedValue <= 0 {
			normalizedValue = 0.00001 // arango does not accept zero value
		} else if normalizedValue > 1 {
			normalizedValue = 1
		}
	} else {
		normalizer.log.Warnf("Upper fence and lower fence are equal: %v", upperFence)
	}
	return normalizedValue
}

func (normalizer *MinMaxNormalizer) normalizeUpdates(updates map[string]message.ArangoUpdate, upperFence float64, lowerFence float64, normalizationField string) {
	for key, update := range updates {
		if len(update.Values) == 0 {
			continue
		}
		normalizedValue := normalizer.getNormalizedValue(update.Values[len(update.Values)-1], upperFence, lowerFence)
		update.Fields = append(update.Fields, normalizationField)
		update.Values = append(update.Values, normalizedValue)
		updates[key] = update
	}
}

func (normalizer *MinMaxNormalizer) NormalizeValues(updates map[string]message.ArangoUpdate, inputField, normalizationField string) (*StatisticalData, error) {
	if _, ok := normalizer.normalizationData[inputField]; ok {
		data := stats.LoadRawData(normalizer.normalizationData[inputField])
		upperFence, lowerFence := 0.0, 0.0
		statisticalData, err := normalizer.getFences(data, &upperFence, &lowerFence)
		if err != nil {
			return nil, err
		}
		normalizer.normalizeUpdates(updates, upperFence, lowerFence, normalizationField)
		return statisticalData, nil
	} else {
		return nil, fmt.Errorf("No data to normalize for field: %v", inputField)
	}
}
