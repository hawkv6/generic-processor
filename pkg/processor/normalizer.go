package processor

import (
	"fmt"
	"math"

	"github.com/hawkv6/generic-processor/pkg/logging"
	"github.com/hawkv6/generic-processor/pkg/message"
	"github.com/montanaflynn/stats"
	"github.com/sirupsen/logrus"
)

type StatisticalData struct {
	q1                 float64
	q3                 float64
	interQuartileRange float64
	lowerFence         float64
	upperFence         float64
}

func NewStatisticalData(q1, q3, iqr, lowerFence, upperFence float64) *StatisticalData {
	return &StatisticalData{
		q1:                 q1,
		q3:                 q3,
		interQuartileRange: iqr,
		lowerFence:         lowerFence,
		upperFence:         upperFence,
	}
}

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
	} else {
		normalizer.normalizationData[fieldName] = append(normalizer.normalizationData[fieldName], value)
	}
}

func (normalizer *MinMaxNormalizer) ResetNormalizationData() {
	normalizer.normalizationData = make(map[string][]float64)
}

func (normalizer *MinMaxNormalizer) getFences(data stats.Float64Data, upperFence, lowerFence *float64) (*StatisticalData, error) {
	quartiles, err := stats.Quartile(data)
	if err != nil {
		return nil, fmt.Errorf("Error calculating quartiles: %v", err)
	}
	interQuartileRange := quartiles.Q3 - quartiles.Q1
	normalizer.log.Debugln("Q1: ", quartiles.Q1)
	normalizer.log.Debugln("Q2 / Median: ", quartiles.Q2)
	normalizer.log.Debugln("Q3: ", quartiles.Q3)
	normalizer.log.Debugln("Interquartile range: ", interQuartileRange)
	outliers, err := stats.QuartileOutliers(data)
	if err != nil {
		return nil, fmt.Errorf("Error calculating outliers: %v", err)
	}
	normalizer.log.Debugf("Outliers: %+v", outliers)

	min, err := stats.Min(data)
	if err != nil {
		return nil, fmt.Errorf("Error calculating min: %v", err)
	}

	max, err := stats.Max(data)
	if err != nil {
		return nil, fmt.Errorf("Error calculating max: %v", err)
	}

	*upperFence = math.Min(quartiles.Q3+1.5*interQuartileRange, max)
	normalizer.log.Debugln("Upper fence: ", *upperFence)
	*lowerFence = math.Max(quartiles.Q1-1.5*interQuartileRange, min)
	normalizer.log.Debugln("Lower fence: ", *lowerFence)
	return NewStatisticalData(quartiles.Q1, quartiles.Q3, interQuartileRange, *lowerFence, *upperFence), nil
}

func (normalizer *MinMaxNormalizer) getNormalizedValue(value float64, upperFence, lowerFence float64) float64 {
	normalizedValue := value
	if upperFence != lowerFence {
		normalizedValue = (value - lowerFence) / (upperFence - lowerFence)
		if normalizedValue <= 0 {
			normalizedValue = 0.0000000001
		} else if normalizedValue > 1 {
			normalizedValue = 1
		}
	} else {
		normalizer.log.Warnf("Upper fence and lower fence are equal: %v", upperFence)
	}
	return normalizedValue
}

func (normalizer *MinMaxNormalizer) NormalizeValues(updates map[string]message.ArangoUpdate, inputField, normalizationField string) (*StatisticalData, error) {
	if _, ok := normalizer.normalizationData[inputField]; ok {
		data := stats.LoadRawData(normalizer.normalizationData[inputField])
		upperFence, lowerFence := 0.0, 0.0
		statisticalData, err := normalizer.getFences(data, &upperFence, &lowerFence)
		if err != nil {
			return nil, err
		}
		for key, update := range updates {
			normalizedValue := normalizer.getNormalizedValue(update.Values[len(update.Values)-1], upperFence, lowerFence)
			update.Fields = append(update.Fields, normalizationField)
			update.Values = append(update.Values, normalizedValue)
			updates[key] = update
		}
		return statisticalData, nil
	} else {
		return nil, fmt.Errorf("No data to normalize for field: %v", inputField)
	}
}
