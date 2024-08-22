package processor

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
