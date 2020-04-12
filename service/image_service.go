package service

import (
	"bytes"
	"github.com/cellargalaxy/go-file-bed/config"
	"github.com/disintegration/imaging"
	"github.com/sirupsen/logrus"
	"math"
)

func CompressionImage(buffer *bytes.Buffer) (*bytes.Buffer, error) {
	imageBytes := buffer.Bytes()
	imageSize := len(imageBytes)
	img, err := imaging.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Error("图片解码失败")
		return nil, err
	}

	buffer = &bytes.Buffer{}
	encodeOption := createJPEGQuality(imageSize)
	imaging.Encode(buffer, img, config.ImageSaveFormat, encodeOption)
	return buffer, err
}

func createJPEGQuality(size int) imaging.EncodeOption {
	power := float64(size) / config.ImageTargetSize
	qualityRatio := math.Pow(0.99, power)
	quality := int(qualityRatio*60 + 20)
	logrus.WithFields(logrus.Fields{"power": power, "qualityRatio": qualityRatio, "quality": quality}).Info("JPEG图片质量")
	return imaging.JPEGQuality(quality)
}
