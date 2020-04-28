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

// jpeg的图片质量范围为[min,max]
// min的取值为：分别使用几M以及几十M的图片，从图片质量100开始往下调，直到调到最小的能容忍的值，为min，默认20
// max的取值为：分别使用几十K以及几百K的图片，从图片质量100开始往下调，直到调到最小的能容忍的值，为max，默认80
// quality=min+(max-min)*(0.99^(size/targetSize))
// 大小     10K          200K  500K         1M           5M           10M          30M          50M          100M
// ratio    0.999497609  0.99  0.975187187  0.949843809  0.773145054  0.597753274  0.21358261   0.076314984  0.005823977
// quality  79.96985657  79.4  78.51123123  76.99062854  66.38870321  55.86519643  32.81495663  24.57889903  20.34943861
func createJPEGQuality(size int) imaging.EncodeOption {
	power := float64(size) / config.ImageTargetSize
	qualityRatio := math.Pow(0.99, power)
	quality := int(config.JpegMinQuality + (config.JpegMaxQuality-config.JpegMinQuality)*qualityRatio)
	logrus.WithFields(logrus.Fields{"power": power, "qualityRatio": qualityRatio, "quality": quality}).Info("JPEG图片质量")
	return imaging.JPEGQuality(quality)
}
