package analysis

import (
	"encoding/hex"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	blockGenerationTimePath = "data/gen_time/"
)

// SaveBlockGenTime saves block generation time
func SaveBlockGenTime(start, end time.Time, indentity []byte) {
	f, err := os.Create(blockGenerationTimePath + hex.EncodeToString(indentity))
	if err != nil {
		logrus.Warn("Can't Save Block Generation Time")
		return
	}
	defer f.Close()
	f.WriteString(hex.EncodeToString(indentity) + "," + strconv.FormatInt(end.Sub(start).Nanoseconds(), 10) + "\n")
}
